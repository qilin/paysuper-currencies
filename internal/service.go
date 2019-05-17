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
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/paysuper/paysuper-recurring-repository/tools"
    "go.uber.org/zap"
    "golang.org/x/net/html/charset"
    "gopkg.in/go-playground/validator.v9"
    "io/ioutil"
    "math"
    "net/http"
    "net/url"
    "time"
)

const (
    centrifugoErrorMessage = "currency_rates_message"
    centrifugoErrorError   = "currency_rates_error"

    errorCentrifugoSendMessage    = "centrifugo message send failed"
    errorEmptyUrl                 = "empty string in url"
    errorRatesRequest             = "rates request error"
    errorDbInsertFailed           = "insert rate to db failed"
    errorDbReqInvalid             = "attempt to insert invalid structure to db"
    errorFromCurrencyNotSupported = "from currency not supported"
    errorToCurrencyNotSupported   = "to currency not supported"
    errorRateTypeInvalid          = "rate type invalid"
    errorCurrencyPairNotExists    = "currency pair is not exists"
    errorDatetimeConversion       = "datetime conversion failed for central bank rate request"
    errorCorrectionRuleNotFound   = "correction rule not found"
    errorInvalidExchangeAmount    = "invalid amount for exchange"

    MIMEApplicationJSON = "application/json"
    MIMEApplicationXML  = "application/xml"
    MIMETextXML         = "text/xml"

    HeaderAccept      = "Accept"
    HeaderContentType = "Content-Type"

    collectionRatesNameTemplate           = "%s_%s"
    collectionRatesNamePrefix             = "currency_rates"
    collectionRatesNameSuffixOxr          = pkg.RateTypeOxr
    collectionRatesNameSuffixCentralbanks = pkg.RateTypeCentralbanks
    collectionRatesNameSuffixPaysuper     = pkg.RateTypePaysuper
    collectionRatesNameSuffixStock        = pkg.RateTypeStock
    collectionRatesNameSuffixCardpay      = pkg.RateTypeCardpay

    collectionNamePaysuperCorrections = "paysuper_corrections"
    collectionNamePaysuperCorridors   = "paysuper_corridors"
    collectionNameCorrectionRules     = "correction_rules"

    ratesPrecision = 10

    dateFormatLayout = "2006-01-02"

    serviceStatusOK   = "ok"
    serviceStatusFail = "fail"
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
        return serviceStatusFail, err
    }
    return serviceStatusOK, nil
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
    return s.contains(s.cfg.OxrSupportedCurrenciesParsed, cur)
}

func (s *Service) contains(slice map[string]bool, item string) bool {
    _, ok := slice[item]
    return ok
}

func (s *Service) getRateByDate(collectionRatesNameSuffix string, from string, to string, date time.Time, res *currencyrates.RateData) error {
    return s.getRate(collectionRatesNameSuffix, from, to, s.getByDateQuery(date), res)
}

func (s *Service) getRate(collectionRatesNameSuffix string, from string, to string, query bson.M, res *currencyrates.RateData) error {
    if !s.isCurrencySupported(from) {
        return errors.New(errorFromCurrencyNotSupported)
    }
    if !s.isCurrencySupported(to) {
        return errors.New(errorToCurrencyNotSupported)
    }

    query["pair"] = from + to

    cName, err := s.getCollectionName(collectionRatesNameSuffix)
    if err != nil {
        return err
    }

    err = s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)
    if err != nil {
        return err
    }

    return nil
}

func (s *Service) saveRates(collectionRatesNameSuffix string, data []interface{}) error {
    cName, err := s.getCollectionName(collectionRatesNameSuffix)
    if err != nil {
        return err
    }

    err = s.db.Collection(cName).Insert(data...)

    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", data)
        return err
    }

    return nil
}

func (s *Service) getByDateQuery(date time.Time) bson.M {
    return bson.M{"created_at": bson.M{"$lte": s.Eod(date)}}
}

func (s *Service) exchangeCurrencyByDate(
    rateType string,
    from string,
    to string,
    amount float64,
    merchantId string,
    date time.Time,
    res *currencyrates.ExchangeCurrencyResponse,
) error {
    return s.exchangeCurrency(rateType, from, to, amount, merchantId, s.getByDateQuery(date), res)
}

func (s *Service) exchangeCurrency(
    rateType string,
    from string,
    to string,
    amount float64,
    merchantId string,
    query bson.M,
    res *currencyrates.ExchangeCurrencyResponse,
) error {

    if amount < 0 {
        return errors.New(errorInvalidExchangeAmount)
    }

    rd := &currencyrates.RateData{}
    err := s.getRate(rateType, from, to, query, rd)
    if err != nil {
        return err
    }

    rule := &currencyrates.CorrectionRule{}

    // ignore error possible here, it not change workflow,
    // and a warning will be written to log in getCorrectionRule method body
    _ = s.getCorrectionRule(rateType, merchantId, rule)

    res.Correction = rule.GetCorrectionValue(rd.Pair)

    // applyCorrectionRule mutate rd object!
    // so, firstly save original rate to response,
    // than apply correction rule
    // and after that save corrected rate to response
    res.OriginalRate = rd.Rate
    s.applyCorrectionRule(rd, rule)
    res.ExchangeRate = rd.Rate

    res.ExchangedAmount = s.toPrecise(amount * res.ExchangeRate)

    return nil
}

func (s *Service) getCorrectionRule(rateType string, merchantId string, r *currencyrates.CorrectionRule) error {

    if !s.contains(s.cfg.RatesTypes, rateType) {
        return errors.New(errorRateTypeInvalid)
    }

    var sort []string
    query := bson.M{"rate_type": rateType}
    if merchantId == "" {
        query["merchant_id"] = ""
    } else {
        query["$or"] = []bson.M{{"merchant_id": merchantId}, {"merchant_id": ""}}
        sort = append(sort, "-merchant_id")
    }
    sort = append(sort, "-_id")

    err := s.db.Collection(collectionNameCorrectionRules).Find(query).Sort(sort...).Limit(1).One(&r)

    if err != nil {
        zap.S().Warnw(errorCorrectionRuleNotFound, "error", err, "rateType", rateType, "merchantId", merchantId)
        return err
    }

    return nil
}

func (s *Service) addCorrectionRule(
    rateType string,
    commonCorrection float64,
    pairCorrection map[string]float64,
    merchantId string,
) error {

    if !s.contains(s.cfg.RatesTypes, rateType) {
        return errors.New(errorRateTypeInvalid)
    }

    rule := &currencyrates.CorrectionRule{
        Id:               bson.NewObjectId().Hex(),
        CreatedAt:        ptypes.TimestampNow(),
        RateType:         rateType,
        CommonCorrection: commonCorrection,
        PairCorrection:   pairCorrection,
    }

    if len(pairCorrection) > 0 {
        for pair := range pairCorrection {
            if !s.isPairExists(pair) {
                zap.S().Errorw(errorCurrencyPairNotExists, "req", rule)
                return errors.New(errorCurrencyPairNotExists)
            }
        }
    }

    if err := s.validateReq(rule); err != nil {
        zap.S().Errorw(errorDbReqInvalid, "error", err, "req", rule)
        return err
    }

    err := s.db.Collection(collectionNameCorrectionRules).Insert(rule)
    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "req", rule)
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

    if err = s.centrifugoClient.Publish(context.Background(), s.cfg.CentrifugoChannel, b); err != nil {
        zap.S().Errorw(errorCentrifugoSendMessage, "error", err, "message", message, "original_error", error)
    }
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

func (s *Service) getCollectionName(suffix string) (string, error) {
    if !s.contains(s.cfg.RatesTypes, suffix) {
        return "", errors.New(errorRateTypeInvalid)
    }

    return fmt.Sprintf(collectionRatesNameTemplate, collectionRatesNamePrefix, suffix), nil
}

func (s *Service) toPrecise(val float64) float64 {
    p := math.Pow(10, ratesPrecision)
    return math.Round(val*p) / p
}

func (s *Service) applyCorrection(rd *currencyrates.RateData, rateType string, merchantId string) {
    rule := &currencyrates.CorrectionRule{}
    err := s.getCorrectionRule(rateType, merchantId, rule)
    if err != nil {
        // here is simple return, no error report need
        return
    }

    s.applyCorrectionRule(rd, rule)
}

func (s *Service) applyCorrectionRule(rd *currencyrates.RateData, rule *currencyrates.CorrectionRule) {
    value := rule.GetCorrectionValue(rd.Pair)
    if value == 0 {
        return
    }
    rd.Rate = s.toPrecise(rd.Rate / (1 - (value / 100)))
}
