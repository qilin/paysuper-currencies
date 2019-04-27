package internal

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "github.com/centrifugal/gocent"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/paysuper/paysuper-recurring-repository/tools"
    "go.uber.org/zap"
    "gopkg.in/go-playground/validator.v9"
    "io/ioutil"
    "net/http"
    "net/url"
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
    errorCurrencyPairNotExists    = "currency pair is not exists"
    errorCurrentRateRequest       = "current rate request error"
    errorCentralBankRateRequest   = "central bank rate request error"
    errorDatetimeConversion       = "datetime conversion failed for central bank rate request"

    MIMEApplicationJSON = "application/json"

    HeaderAccept        = "Accept"
    HeaderContentType   = "Content-Type"
    HeaderAuthorization = "Authorization"

    BasicAuthorization = "Basic %s"
)

// Service is application entry point.
type Service struct {
    cfg              *config.Config
    db               *mgo.Database
    centrifugoClient *gocent.Client
    validate         *validator.Validate
}

// NewService create new Service.
func NewService(cfg *config.Config, db *mgo.Database) (*Service, error) {
    return &Service{
        cfg:      cfg,
        db:       db,
        validate: validator.New(),
    }, nil
}

// Status used to return micro service health.
func (s *Service) Status() (interface{}, error) {
    err := s.db.Session.Ping()
    if err != nil {
        return "fail", err
    }
    return "ok", nil
}

func (s *Service) Init() {

    s.centrifugoClient = gocent.New(
        gocent.Config{
            Addr:       s.cfg.CentrifugoURL,
            Key:        s.cfg.CentrifugoSecret,
            HTTPClient: tools.NewLoggedHttpClient(zap.S()),
        },
    )

    go s.initRateRequests()
}

func (s *Service) initRateRequests() {
    go s.requestRatesXe()

    xeRateTicker := time.NewTicker(time.Hour * time.Duration(s.cfg.XeRatesRequestPeriod))

    for {
        select {
        case <-xeRateTicker.C:
            go s.requestRatesXe()
        }
    }
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

func (s *Service) getJson(resp *http.Response, target interface{}) error {
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    return json.Unmarshal(body, target)
}

func (s *Service) getCorrectionForPair(pair string) float64 {
    correction := float64(1)

    if !s.isPairExists(pair) {
        return correction
    }

    if val, ok := s.cfg.XePairsCorrections[pair]; ok {
        correction = val
    } else {
        if s.cfg.XeCommonCorrection > 0 && s.cfg.XeCommonCorrection != 1 {
            correction = s.cfg.XeCommonCorrection
        }
    }
    return correction
}

func (s *Service) isPairExists(pair string) bool {
    if len(pair) != 6 {
        return false
    }

    from := string(pair[0:3])
    to := string(pair[3:6])

    if from == to {
        return false
    }

    if !s.isCurrencySupported(from) || !s.isCurrencySupported(to) {
        return false
    }
    return true
}

func (s *Service) isCurrencySupported(cur string) bool {
    return s.contains(s.cfg.XeSupportedCurrencies, cur)
}

func (s *Service) contains(slice []string, item string) bool {
    set := make(map[string]struct{}, len(slice))
    for _, s := range slice {
        set[s] = struct{}{}
    }

    _, ok := set[item]
    return ok
}

func (s *Service) saveRate(rd *currencyrates.RateData) error {
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

    err := s.db.C(pkg.CollectionRate).Insert(rd)

    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", rd)
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

func (s *Service) GetCurrentRate(
    ctx context.Context,
    req *currencyrates.GetCurrentRateRequest,
    res *currencyrates.RateData,
) error {
    query := bson.M{"is_cb_rate": false}
    err := s.getRate(req.From, req.To, query, res)
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

    query := bson.M{"is_cb_rate": true, "created_at": bson.M{"$lte": dt}}
    err = s.getRate(req.From, req.To, query, res)
    if err != nil {
        zap.S().Errorw(errorCentralBankRateRequest, "error", err, "req", req)
        return err
    }

    return nil
}


func (s *Service) getRate(from string, to string, query bson.M, res *currencyrates.RateData) error {
    if !s.isCurrencySupported(from) {
        return errors.New(errorFromCurrencyNotSupported)
    }
    if !s.isCurrencySupported(to) {
        return errors.New(errorToCurrencyNotSupported)
    }

    query["pair"] = from + to

    err := s.db.C(pkg.CollectionRate).Find(query).Sort("-_id").Limit(1).One(&res)
    if err != nil {
        return err
    }

    return nil
}