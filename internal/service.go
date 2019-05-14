package internal

import (
    "bytes"
    "context"
    "encoding/json"
    "encoding/xml"
    "errors"
    "fmt"
    "github.com/centrifugal/gocent"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/paysuper/paysuper-recurring-repository/tools"
    "github.com/thetruetrade/gotrade"
    "github.com/thetruetrade/gotrade/indicators"
    "go.uber.org/zap"
    "golang.org/x/net/html/charset"
    "gopkg.in/go-playground/validator.v9"
    "io/ioutil"
    "math"
    "net/http"
    "net/url"
    "strings"
    "time"
)

const (
    centrifugoChannel      = "currency_rates-failed"
    centrifugoErrorMessage = "currency_rates_message"
    centrifugoErrorError   = "currency_rates_error"

    errorCentrifugoSendMessage    = "centrifugo message send failed"
    errorEmptyUrl                 = "empty string in url"
    errorRatesRequest             = "rates request error"
    errorDbInsertFailed           = "insert rate to db failed"
    errorDbReqInvalid             = "attempt to insert invalid structure to db"
    errorFromCurrencyNotSupported = "from currency not supported"
    errorToCurrencyNotSupported   = "to currency not supported"
    errorCollectionSuffixEmpty    = "collection suffix is empty"
    errorCurrencyPairNotExists    = "currency pair is not exists"
    errorCurrentRateRequest       = "current rate request error"
    errorCentralBankRateRequest   = "central bank rate request error"
    errorDatetimeConversion       = "datetime conversion failed for central bank rate request"

    MIMEApplicationJSON = "application/json"
    MIMEApplicationXML  = "application/xml"

    HeaderAccept      = "Accept"
    HeaderContentType = "Content-Type"

    collectionNameTemplate = "%s_%s"

    collectionSuffixOxr      = "oxr"
    collectionSuffixCb       = "centralbanks"
    collectionSuffixPaysuper = "paysuper"
    collectionSuffixStock    = "stock"
    collectionSuffixCardpay  = "cardpay"

    collectionNamePaysuperCorrections = "paysuper_corrections"
    collectionNamePaysuperCorridors   = "paysuper_corridors"

    ratesPrecision = 4

    dateFormatLayout = "2006-01-02"
)

// Service is application entry point.
type Service struct {
    cfg              *config.Config
    db               *database.Source
    centrifugoClient *gocent.Client
    validate         *validator.Validate
}

// NewService create new Service.
func NewService(cfg *config.Config, db *database.Source) (*Service, error) {
    return &Service{
        cfg:      cfg,
        db:       db,
        validate: validator.New(),
        centrifugoClient: gocent.New(
            gocent.Config{
                Addr:       cfg.CentrifugoURL,
                Key:        cfg.CentrifugoSecret,
                HTTPClient: tools.NewLoggedHttpClient(zap.S()),
            },
        ),
    }, nil
}

// Status used to return micro service health.
func (s *Service) Status() (interface{}, error) {
    err := s.db.Ping()
    if err != nil {
        return "fail", err
    }
    return "ok", nil
}

func (s *Service) validateReq(req interface{}) error {
    err := s.validate.Struct(req)
    if err != nil {
        return err
    }
    return nil
}

func (s *Service) validateUrl(cUrl string) (*url.URL, error) {
    if cUrl == "" {
        return nil, errors.New(errorEmptyUrl)
    }

    u, err := url.ParseRequestURI(cUrl)

    if err != nil {
        return nil, err
    }
    return u, nil
}

func (s *Service) request(method string, url string, req []byte, headers map[string]string) (*http.Response, error) {
    client := tools.NewLoggedHttpClient(zap.S())
    httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(req))

    if err != nil {
        return nil, err
    }

    for k, v := range headers {
        httpReq.Header.Add(k, v)
    }

    resp, err := client.Do(httpReq)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent &&
        resp.StatusCode != http.StatusUnprocessableEntity {
        return nil, errors.New(errorRatesRequest)
    }

    return resp, nil
}

func (s *Service) decodeJson(resp *http.Response, target interface{}) error {
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    return json.Unmarshal(body, target)
}

func (s *Service) decodeXml(resp *http.Response, target interface{}) error {
    decoder := xml.NewDecoder(resp.Body)
    decoder.CharsetReader = charset.NewReaderLabel
    return decoder.Decode(&target)
}

func (s *Service) isPairExists(pair string) bool {
    if len(pair) != 6 {
        return false
    }

    from := string(pair[0:3])
    to := string(pair[3:6])

    if !s.isCurrencySupported(from) || !s.isCurrencySupported(to) {
        return false
    }
    return true
}

func (s *Service) isCurrencySupported(cur string) bool {
    return s.contains(s.cfg.OxrSupportedCurrencies, cur)
}

func (s *Service) contains(slice []string, item string) bool {
    set := make(map[string]struct{}, len(slice))
    for _, s := range slice {
        set[s] = struct{}{}
    }

    _, ok := set[item]
    return ok
}

func (s *Service) saveRates(collectionSuffix string, rds []*currencyrates.RateData) error {
    if collectionSuffix == "" {
        return errors.New(errorCollectionSuffixEmpty)
    }

    data := []interface{}{}

    for _, rd := range rds {

        if !s.isPairExists(rd.Pair) {
            zap.S().Errorw(errorCurrencyPairNotExists, "req", rd)
            return errors.New(errorCurrencyPairNotExists)
        }

        rd.Id = bson.NewObjectId().Hex()
        rd.CreatedAt = ptypes.TimestampNow()

        if err := s.validateReq(rd); err != nil {
            zap.S().Errorw(errorDbReqInvalid, "error", err, "data", rd)
            return err
        }

        data = append(data, rd)
    }

    cName := s.getCollectionName(collectionSuffix)

    err := s.db.Collection(cName).Insert(data...)

    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", rds)
        return err
    }

    return nil
}

func (s *Service) sendCentrifugoMessage(message string, error error) {
    msg := map[string]interface{}{
        centrifugoErrorMessage: message,
        centrifugoErrorError:   error.Error(),
    }

    b, err := json.Marshal(msg)

    if err != nil {
        zap.S().Errorw(errorCentrifugoSendMessage, "error", err, "message", message, "original_error", error)
        return
    }

    if err = s.centrifugoClient.Publish(context.Background(), centrifugoChannel, b); err != nil {
        zap.S().Errorw(errorCentrifugoSendMessage, "error", err, "message", message, "original_error", error)
    }
}

func (s *Service) SetPaysuperCorrectionCorridor(
    ctx context.Context,
    req *currencyrates.CorrectionCorridor,
    res *currencyrates.EmptyResponse,
) error {

    corridor := PaysuperCorridor{
        Value:     req.Value,
        CreatedAt: time.Now(),
    }

    err := s.db.Collection(collectionNamePaysuperCorridors).Insert(corridor)
    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", corridor)
        return err
    }

    return nil
}

func (s *Service) GetOxrRate(
    ctx context.Context,
    req *currencyrates.GetRateRequest,
    res *currencyrates.RateData,
) error {
    err := s.getRate(collectionSuffixOxr, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorCurrentRateRequest, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) GetCentralBankRateForDate(
    ctx context.Context,
    req *currencyrates.GetCentralBankRateRequest,
    res *currencyrates.RateData,
) error {

    dt, err := ptypes.Timestamp(req.Datetime)
    if err != nil {
        zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
        return err
    }

    query := bson.M{"created_at": bson.M{"$lte": s.Eod(dt)}}
    err = s.getRate(collectionSuffixCb, req.From, req.To, query, res)
    if err != nil {
        zap.S().Errorw(errorCentralBankRateRequest, "error", err, "req", req)
        return err
    }

    return nil
}

func (s *Service) GetPaysuperRate(
    ctx context.Context,
    req *currencyrates.GetRateRequest,
    res *currencyrates.RateData,
) error {
    err := s.getRate(collectionSuffixPaysuper, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorCurrentRateRequest, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) GetStockRate(
    ctx context.Context,
    req *currencyrates.GetRateRequest,
    res *currencyrates.RateData,
) error {
    err := s.getRate(collectionSuffixStock, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorCurrentRateRequest, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) GetCardpayRate(
    ctx context.Context,
    req *currencyrates.GetRateRequest,
    res *currencyrates.RateData,
) error {
    err := s.getRate(collectionSuffixCardpay, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorCurrentRateRequest, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) getRate(collectionSuffix string, from string, to string, query bson.M, res *currencyrates.RateData) error {
    if !s.isCurrencySupported(from) {
        return errors.New(errorFromCurrencyNotSupported)
    }
    if !s.isCurrencySupported(to) {
        return errors.New(errorToCurrencyNotSupported)
    }

    query["pair"] = from + to

    cName := s.getCollectionName(collectionSuffix)

    err := s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)
    if err != nil {
        return err
    }

    return nil
}

func (s *Service) GetRatesForBollinger(collectionSuffix string, pair string, dateFrom time.Time) (res []float64, err error) {
    if !s.isPairExists(pair) {
        return nil, errors.New(errorCurrencyPairNotExists)
    }

    cName := s.getCollectionName(collectionSuffix)

    q := []bson.M{
        {"$match": bson.M{"pair": pair, "created_at": bson.M{"$gte": s.Bod(dateFrom)}}},
        {"$group": bson.M{
            "_id":  bson.M{"create_date": "$create_date"},
            "last": bson.M{"$last": "$rate"},
        }},
    }

    var resp []map[string]interface{}
    err = s.db.Collection(cName).Pipe(q).All(&resp)

    if err != nil {
        return nil, err
    }

    for _, val := range resp {
        res = append(res, val["last"].(float64))
    }

    return res, nil
}

func (s *Service) Bollinger(rates []float64, timePeriod int) ([]float64, []float64, []float64) {

    priceStream := gotrade.NewDailyDOHLCVStream()
    bb, _ := indicators.NewBollingerBandsForStream(priceStream, timePeriod, gotrade.UseClosePrice)

    for _, val := range rates {
        dohlcv := gotrade.NewDOHLCVDataItem(time.Now(), 0, 0, 0, val, 0)
        priceStream.ReceiveTick(dohlcv)
    }

    return bb.LowerBand, bb.MiddleBand, bb.UpperBand
}

// returns begin-of-day for passed date
func (s *Service) Bod(t time.Time) time.Time {
    year, month, day := t.Date()
    return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// returns end-of-day for passed date
func (s *Service) Eod(t time.Time) time.Time {
    year, month, day := t.Date()
    return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

func (s *Service) getCollectionName(suffix string) string {
    return fmt.Sprintf(collectionNameTemplate, pkg.CollectionRate, strings.ToLower(suffix))
}

func (s *Service) toPrecise(val float64) float64 {
    p := math.Pow(10, ratesPrecision)
    return math.Round(val*p) / p
}
