package service

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/ProtocolONE/rabbitmq/pkg"
	"github.com/centrifugal/gocent"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-currencies/config"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-database-mongo"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/paysuper/paysuper-tools"
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

	errorCentrifugoSendMessage    = "centrifugo message send failed"
	errorEmptyUrl                 = "empty string in url"
	errorRatesRequest             = "rates request error"
	errorDbInsertFailed           = "insert rate to db failed"
	errorDbReqInvalid             = "attempt to insert invalid structure to db"
	errorFromCurrencyNotSupported = "from currency not supported"
	errorSourceNotSupported       = "source not supported"
	errorToCurrencyNotSupported   = "to currency not supported"
	errorRateTypeInvalid          = "rate type invalid"
	errorExchangeDirectionInvalid = "exchange direction invalid"
	errorCorrectionPercentInvalid = "correction percent invalid"
	errorCurrencyPairNotExists    = "currency pair is not exists"
	errorDatetimeConversion       = "datetime conversion failed for central bank rate request"
	errorCorrectionRuleNotFound   = "correction rule not found"

	mimeApplicationJSON = "application/json"
	mimeApplicationXML  = "application/xhtml+xml,application/xml"
	mimeTextXML         = "text/xml"
	defaultUserAgent    = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Mobile Safari/537.36"

	headerAccept      = "Accept"
	headerCookie      = "Cookie"
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"

	collectionRatesNameTemplate           = "%s_%s"
	collectionRatesNamePrefix             = "currency_rates"
	collectionRatesNameSuffixOxr          = pkg.RateTypeOxr
	collectionRatesNameSuffixCentralbanks = pkg.RateTypeCentralbanks
	collectionRatesNameSuffixPaysuper     = pkg.RateTypePaysuper
	collectionRatesNameSuffixStock        = pkg.RateTypeStock

	collectionNamePaysuperCorrections = "paysuper_corrections"
	collectionNameCorrectionRules     = "correction_rules"

	ratesPrecision = 6

	dateFormatLayout = "2006-01-02"

	serviceStatusOK   = "ok"
	serviceStatusFail = "fail"

	stubSource               = "STUB"
	defaultHttpClientTimeout = 30
)

var (
	availableCentralbanksSources = map[string]bool{
		cbeuSource: true,
		cbauSource: true,
		cbcaSource: true,
		cbplSource: true,
		cbrfSource: true,
	}
)

// Service is application entry point.
type Service struct {
	cfg                 *config.Config
	db                  *database.Source
	centrifugoClient    *gocent.Client
	validate            *validator.Validate
	cardpayBroker       *rabbitmq.Broker
	cardpayRetryBroker  *rabbitmq.Broker
	cardpayFinishBroker *rabbitmq.Broker
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

// Init rabbitMq brokers and check for active triggers for delayed tasks
func (s *Service) Init() error {
	return nil
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

	zap.S().Info("Sending request to url: ", url)

	client := tools.NewLoggedHttpClient(zap.S())
	client.Timeout = time.Duration(defaultHttpClientTimeout * time.Second)

	// prevent following to redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
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

	cookies := resp.Cookies()
	if resp.StatusCode == http.StatusFound && len(cookies) > 0 {
		c := []string{}
		for _, v := range cookies {
			c = append(c, v.Name+"="+v.Value)
		}
		headers[headerCookie] = strings.Join(c, ";")
		return s.request(method, url, req, headers)
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
	return s.contains(s.cfg.SupportedCurrenciesParsed, cur)
}

func (s *Service) contains(slice map[string]bool, item string) bool {
	_, ok := slice[item]
	return ok
}

func (s *Service) getRateByDate(collectionRatesNameSuffix string, from string, to string, date time.Time, source string, res *currencies.RateData) error {
	return s.getRate(collectionRatesNameSuffix, from, to, s.getByDateQuery(date), source, res)
}

func (s *Service) getRate(collectionRatesNameSuffix string, from string, to string, query bson.M, source string, res *currencies.RateData) error {

	var err error

	if !s.isCurrencySupported(from) {
		return errors.New(errorFromCurrencyNotSupported)
	}
	if !s.isCurrencySupported(to) {
		return errors.New(errorToCurrencyNotSupported)
	}

	pair := from + to

	// stub for rate with the same from/to currencies
	if from == to {
		res.Rate = s.toPrecise(1)
		res.Pair = pair
		res.Source = stubSource
		res.CreatedAt = ptypes.TimestampNow()
		res.Volume = 1

		return nil
	}

	query["pair"] = pair

	isCentralbank := collectionRatesNameSuffix == pkg.RateTypeCentralbanks

	cName, err := s.getCollectionName(collectionRatesNameSuffix)
	if err != nil {
		return err
	}

	if isCentralbank {
		source = strings.ToUpper(source)
		if _, ok := availableCentralbanksSources[source]; !ok {
			// temporarily ignore unsupported central banks
			zap.S().Warnw(errorSourceNotSupported, "source", source)
		}
		query["source"] = source
	}

	err = s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)
	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, cName),
			zap.Any(pkg.ErrorDatabaseFieldQuery, query),
		)
	}

	// requested pair is not found in central banks rates
	// try to fallback to OXR rate for it
	if err == mgo.ErrNotFound && isCentralbank {
		cName, err = s.getCollectionName(collectionRatesNameSuffixOxr)
		if err != nil {
			return err
		}
		delete(query, "source")
		err = s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)
		if err != nil {
			zap.L().Error(
				pkg.ErrorDatabaseQueryFailed,
				zap.Error(err),
				zap.String(pkg.ErrorDatabaseFieldCollection, cName),
				zap.Any(pkg.ErrorDatabaseFieldQuery, query),
			)
		}
	}

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
	return bson.M{"created_at": bson.M{"$lte": now.New(date).EndOfDay()}}
}

func (s *Service) exchangeCurrencyByDate(
	rateType string,
	exchangeDirection string,
	from string,
	to string,
	amount float64,
	merchantId string,
	date time.Time,
	source string,
	res *currencies.ExchangeCurrencyResponse,
) error {
	return s.exchangeCurrency(rateType, exchangeDirection, from, to, amount, merchantId, s.getByDateQuery(date), source, res)
}

func (s *Service) exchangeCurrency(
	rateType string,
	exchangeDirection string,
	from string,
	to string,
	amount float64,
	merchantId string,
	query bson.M,
	source string,
	res *currencies.ExchangeCurrencyResponse,
) error {
	rd := &currencies.RateData{}
	err := s.getRate(rateType, from, to, query, source, rd)
	if err != nil {
		return err
	}

	// ignore error possible here, it not change workflow,
	// and a warning will be written to log in getCorrectionRule method body
	rule, _ := s.getCorrectionRule(rateType, exchangeDirection, merchantId)
	if rule == nil {
		rule = &currencies.CorrectionRule{}
	}

	res.Correction = rule.GetCorrectionValue(rd.Pair)
	res.ExchangeDirection = exchangeDirection

	// applyCorrectionRule mutate rd object!
	// so, firstly save original rate to response,
	// than apply correction rule
	// and after that save corrected rate to response
	res.OriginalRate = rd.Rate
	s.applyCorrectionRule(rd, rule)
	res.ExchangeRate = rd.Rate

	res.ExchangedAmount = s.toPrecise(amount * res.ExchangeRate)

	zap.S().Infow("exchange currency", "from", from, "to", to, "amount", amount,
		"rateType", rateType, "merchantId", merchantId, "query", query, "res", res)

	return nil
}

func (s *Service) getCorrectionRule(rateType, exchangeDirection, merchantId string) (r *currencies.CorrectionRule, err error) {

	if !s.contains(s.cfg.RatesTypes, rateType) {
		return nil, errors.New(errorRateTypeInvalid)
	}

	if !s.contains(pkg.SupportedExchangeDirections, exchangeDirection) {
		return nil, errors.New(errorExchangeDirectionInvalid)
	}

	var sort []string
	query := bson.M{
		"rate_type":          rateType,
		"exchange_direction": exchangeDirection,
	}
	if merchantId == "" {
		query["merchant_id"] = ""
	} else {
		query["$or"] = []bson.M{{"merchant_id": merchantId}, {"merchant_id": ""}}
		sort = append(sort, "-merchant_id")
	}
	sort = append(sort, "-_id")

	err = s.db.Collection(collectionNameCorrectionRules).Find(query).Sort(sort...).Limit(1).One(&r)

	if err != nil {
		zap.S().Warnw(errorCorrectionRuleNotFound, "error", err, "rateType", rateType, "exchangeDirection", exchangeDirection, "merchantId", merchantId)
		return
	}

	return
}

func (s *Service) addCorrectionRule(
	rateType string,
	exchangeDirection string,
	commonCorrection float64,
	pairCorrection map[string]float64,
	merchantId string,
) error {

	if !s.contains(s.cfg.RatesTypes, rateType) {
		return errors.New(errorRateTypeInvalid)
	}

	if !s.contains(pkg.SupportedExchangeDirections, exchangeDirection) {
		return errors.New(errorExchangeDirectionInvalid)
	}

	if !s.isCorrectionPercentValid(commonCorrection) {
		return errors.New(errorCorrectionPercentInvalid)
	}

	rule := &currencies.CorrectionRule{
		Id:                bson.NewObjectId().Hex(),
		CreatedAt:         ptypes.TimestampNow(),
		RateType:          rateType,
		ExchangeDirection: exchangeDirection,
		CommonCorrection:  commonCorrection,
		PairCorrection:    pairCorrection,
		MerchantId:        merchantId,
	}

	if len(pairCorrection) > 0 {
		for pair, val := range pairCorrection {
			if !s.isPairExists(pair) {
				zap.S().Errorw(errorCurrencyPairNotExists, "req", rule)
				return errors.New(errorCurrencyPairNotExists)
			}
			if !s.isCorrectionPercentValid(val) {
				return errors.New(errorCorrectionPercentInvalid)
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

func (s *Service) isCorrectionPercentValid(val float64) bool {
	return val >= 0 && val <= 100
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

func (s *Service) getCollectionName(suffix string) (string, error) {
	if !s.contains(s.cfg.RatesTypes, suffix) {
		return "", errors.New(errorRateTypeInvalid)
	}

	return fmt.Sprintf(collectionRatesNameTemplate, collectionRatesNamePrefix, suffix), nil
}

func (s *Service) toPrecise(val float64) float64 {
	p := math.Pow(10, ratesPrecision)
	return math.Ceil(val*p) / p
}

func (s *Service) applyCorrection(rd *currencies.RateData, rateType, exchangeDirection, merchantId string) {
	rule, err := s.getCorrectionRule(rateType, exchangeDirection, merchantId)
	if err != nil {
		// here is simple return, no error report need
		return
	}
	if rule == nil {
		rule = &currencies.CorrectionRule{}
	}

	s.applyCorrectionRule(rd, rule)
}

func (s *Service) applyCorrectionRule(rd *currencies.RateData, rule *currencies.CorrectionRule) {
	value := rule.GetCorrectionValue(rd.Pair)
	if value == 0 {
		return
	}

	divider := float64(1)

	switch rule.ExchangeDirection {

	case pkg.ExchangeDirectionSell:
		divider = 1 - (value / 100)

	case pkg.ExchangeDirectionBuy:
		divider = 1 + (value / 100)
	}

	rd.Rate = s.toPrecise(rd.Rate / divider)
}
