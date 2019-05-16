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
    "strings"
    "time"
)

const (
    centrifugoErrorMessage = "currency_rates_message"
    centrifugoErrorError   = "currency_rates_error"

    errorCentrifugoSendMessage        = "centrifugo message send failed"
    errorEmptyUrl                     = "empty string in url"
    errorRatesRequest                 = "rates request error"
    errorDbInsertFailed               = "insert rate to db failed"
    errorDbReqInvalid                 = "attempt to insert invalid structure to db"
    errorFromCurrencyNotSupported     = "from currency not supported"
    errorToCurrencyNotSupported       = "to currency not supported"
    errorCollectionSuffixEmpty        = "collection suffix is empty"
    errorCurrencyPairNotExists        = "currency pair is not exists"
    errorCurrentRateRequest           = "current rate request error"
    errorCentralBankRateRequest       = "central bank rate request error"
    errorDatetimeConversion           = "datetime conversion failed for central bank rate request"
    errorCorrectionRuleRequestInvalid = "correction rule invalid request"
    errorCorrectionRuleNotFound       = "correction rule not found"

    MIMEApplicationJSON = "application/json"
    MIMEApplicationXML  = "application/xml"
    MIMETextXML         = "text/xml"

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
    collectionNameCorrectionRules     = "correction_rules"

    ratesPrecision = 8

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
    return s.contains(s.cfg.OxrSupportedCurrenciesParsed, cur)
}

func (s *Service) contains(slice map[string]bool, item string) bool {
    _, ok := slice[item]
    return ok
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

func (s *Service) saveRates(collectionSuffix string, data []interface{}) error {
    if collectionSuffix == "" {
        return errors.New(errorCollectionSuffixEmpty)
    }

    cName := s.getCollectionName(collectionSuffix)

    err := s.db.Collection(cName).Insert(data...)

    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", data)
        return err
    }

    return nil
}

func (s *Service) getCorrectionRule(rateType string, merchantId string, r *currencyrates.CorrectionRule) error {

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

func (s *Service) getCollectionName(suffix string) string {
    return fmt.Sprintf(collectionNameTemplate, pkg.CollectionRate, strings.ToLower(suffix))
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
