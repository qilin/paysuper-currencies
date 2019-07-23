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
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-currencies/config"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/tools"
	"github.com/streadway/amqp"
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
	errorCurrencyPairNotExists    = "currency pair is not exists"
	errorDatetimeConversion       = "datetime conversion failed for central bank rate request"
	errorCorrectionRuleNotFound   = "correction rule not found"
	errorInvalidExchangeAmount    = "invalid amount for exchange"
	errorBrokerMaxRetryReached    = "broker max retry reached"
	errorBrokerRetryPublishFailed = "broker retry publishing failed"
	errorPullTrigger              = "pull trigger failed"
	errorReleaseTrigger           = "release trigger failed"
	errorDelayedFunction          = "error in delayed function"

	mimeApplicationJSON = "application/json"
	mimeApplicationXML  = "application/xml"
	mimeTextXML         = "text/xml"

	headerAccept      = "Accept"
	headerCookie      = "Cookie"
	headerContentType = "Content-Type"

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
	collectionNameTriggers            = "triggers"

	ratesPrecision = 10

	dateFormatLayout = "2006-01-02"

	serviceStatusOK   = "ok"
	serviceStatusFail = "fail"

	retryCountHeader = "x-retry-count"

	triggerCardpay = 1

	stubSource = "STUB"
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

type trigger struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Type      int           `bson:"type"`
	Active    bool          `bson:"active"`
	CreatedAt time.Time     `bson:"created_at"`
}

type callable func() error

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
	var err error
	s.cardpayBroker, s.cardpayRetryBroker, s.cardpayFinishBroker, err = s.getBrokers(
		pkg.CardpayTopicRateData,
		pkg.CardpayTopicRateDataRetry,
		pkg.CardpayTopicRateDataFinished,
	)

	tgr, err := s.getTrigger(triggerCardpay)
	if err != nil {
		return err
	}

	if tgr.Active == true {
		now := time.Now()
		eod := s.eod(now)
		delta := eod.Sub(now)
		return s.planDelayedTask(int64(delta.Seconds()), tgr.Type, s.CalculatePaysuperCorrections)
	}

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

func (s *Service) getBrokers(topicName string, retryTopicName string, finishTopicName string) (*rabbitmq.Broker, *rabbitmq.Broker, *rabbitmq.Broker, error) {
	broker, err := rabbitmq.NewBroker(s.cfg.BrokerAddress)
	if err != nil {
		return nil, nil, nil, err
	}

	retryBroker, err := rabbitmq.NewBroker(s.cfg.BrokerAddress)
	if err != nil {
		return nil, nil, nil, err
	}

	finishBroker, err := rabbitmq.NewBroker(s.cfg.BrokerAddress)
	if err != nil {
		return nil, nil, nil, err
	}

	broker.Opts.ExchangeOpts.Name = topicName

	retryBroker.Opts.QueueOpts.Args = amqp.Table{
		"x-dead-letter-exchange":    retryTopicName,
		"x-message-ttl":             int32(s.cfg.BrokerRetryTimeout * 1000),
		"x-dead-letter-routing-key": "*",
	}
	retryBroker.Opts.ExchangeOpts.Name = retryTopicName

	finishBroker.Opts.ExchangeOpts.Name = finishTopicName

	err = broker.RegisterSubscriber(topicName, s.SetRatesCardpay)

	if err != nil {
		return nil, nil, nil, err
	}

	err = finishBroker.RegisterSubscriber(topicName, s.PullRecalcTrigger)

	if err != nil {
		return nil, nil, nil, err
	}

	return broker, retryBroker, finishBroker, nil
}

func (s *Service) retry(msg proto.Message, dlv amqp.Delivery, msgID string) error {
	var rtc = int32(0)

	if v, ok := dlv.Headers[retryCountHeader]; ok {
		rtc = v.(int32)
	}

	if rtc >= s.cfg.BrokerMaxRetry {
		zap.L().Error(errorBrokerMaxRetryReached, zap.String("msgid", msgID))
		s.sendCentrifugoMessage(msgID, errors.New(errorBrokerMaxRetryReached))
		return nil
	}

	err := s.cardpayRetryBroker.Publish(dlv.RoutingKey, msg, amqp.Table{retryCountHeader: rtc + 1})

	if err != nil {
		zap.L().Warn(errorBrokerRetryPublishFailed, zap.String("msgid", msgID), zap.Error(err))
		s.sendCentrifugoMessage(msgID, err)
		return err
	}

	return nil
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

	isCardpay := collectionRatesNameSuffix == pkg.RateTypeCardpay
	isCentralbank := collectionRatesNameSuffix == pkg.RateTypeCentralbanks

	cName, err := s.getCollectionName(collectionRatesNameSuffix)
	if err != nil {
		return err
	}

	if isCardpay {
		q := []bson.M{
			{"$match": query},
			{"$group": bson.M{
				"_id":         bson.M{"create_date": "$create_date"},
				"numerator":   bson.M{"$sum": bson.M{"$multiply": []string{"$rate", "$volume"}}},
				"denominator": bson.M{"$sum": "$volume"},
			}},
			{"$project": bson.M{
				"value": bson.M{"$divide": []string{"$numerator", "$denominator"}},
			}},
			{"$limit": 1},
		}
		var resp []map[string]interface{}
		err = s.db.Collection(cName).Pipe(q).All(&resp)

		if err != nil {
			return err
		}

		if len(resp) == 0 {
			return mgo.ErrNotFound
		}

		res.Pair = pair
		res.Rate = s.toPrecise(resp[0]["value"].(float64))
		res.Source = cardpaySource

	} else {
		if isCentralbank {
			source = strings.ToUpper(source)
			if _, ok := availableCentralbanksSources[source]; !ok {
				return errors.New(errorSourceNotSupported)
			}
			query["source"] = source
		}

		err = s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)

		// requested pair is not found in central banks rates
		// try to fallback to OXR rate for it
		if err == mgo.ErrNotFound && isCentralbank {
			cName, err = s.getCollectionName(collectionRatesNameSuffixOxr)
			if err != nil {
				return err
			}
			query["source"] = bson.M{"$ne": ""}
			err = s.db.Collection(cName).Find(query).Sort("-_id").Limit(1).One(&res)
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
	return bson.M{"created_at": bson.M{"$lte": s.eod(date)}}
}

func (s *Service) exchangeCurrencyByDate(
	rateType string,
	from string,
	to string,
	amount float64,
	merchantId string,
	date time.Time,
	source string,
	res *currencies.ExchangeCurrencyResponse,
) error {
	return s.exchangeCurrency(rateType, from, to, amount, merchantId, s.getByDateQuery(date), source, res)
}

func (s *Service) exchangeCurrency(
	rateType string,
	from string,
	to string,
	amount float64,
	merchantId string,
	query bson.M,
	source string,
	res *currencies.ExchangeCurrencyResponse,
) error {

	if amount < 0 {
		return errors.New(errorInvalidExchangeAmount)
	}

	rd := &currencies.RateData{}
	err := s.getRate(rateType, from, to, query, source, rd)
	if err != nil {
		return err
	}

	rule := &currencies.CorrectionRule{}

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

	zap.S().Infow("exchange currency", "from", from, "to", to, "amount", amount,
		"rateType", rateType, "merchantId", merchantId, "query", query, "res", res)

	return nil
}

func (s *Service) getCorrectionRule(rateType string, merchantId string, r *currencies.CorrectionRule) error {

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

	rule := &currencies.CorrectionRule{
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
func (s *Service) bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// returns end-of-day for passed date
func (s *Service) eod(t time.Time) time.Time {
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

func (s *Service) applyCorrection(rd *currencies.RateData, rateType string, merchantId string) {
	rule := &currencies.CorrectionRule{}
	err := s.getCorrectionRule(rateType, merchantId, rule)
	if err != nil {
		// here is simple return, no error report need
		return
	}

	s.applyCorrectionRule(rd, rule)
}

func (s *Service) applyCorrectionRule(rd *currencies.RateData, rule *currencies.CorrectionRule) {
	value := rule.GetCorrectionValue(rd.Pair)
	if value == 0 {
		return
	}
	rd.Rate = s.toPrecise(rd.Rate / (1 - (value / 100)))
}

func (s *Service) pullTrigger(triggerType int) error {
	trg := &trigger{
		Type:      triggerType,
		Active:    true,
		CreatedAt: time.Now(),
	}
	return s.db.Collection(collectionNameTriggers).Insert(trg)
}

func (s *Service) releaseTrigger(triggerType int) error {
	trg := &trigger{
		Type:      triggerType,
		Active:    false,
		CreatedAt: time.Now(),
	}
	return s.db.Collection(collectionNameTriggers).Insert(trg)
}

func (s *Service) getTrigger(triggerType int) (*trigger, error) {
	query := bson.M{"type": triggerType}
	res := &trigger{
		Type:      triggerType,
		Active:    false,
		CreatedAt: time.Now(),
	}
	err := s.db.Collection(collectionNameTriggers).Find(query).Sort("-_id").Limit(1).One(res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	return res, nil
}

func (s *Service) planDelayedTask(delay int64, trigger int, fn callable) error {
	ticker := time.NewTicker(time.Second * time.Duration(delay))
	err := s.pullTrigger(trigger)
	if err != nil {
		zap.S().Errorw(errorPullTrigger, "error", err, "trigger", trigger)
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				err := fn()
				zap.S().Errorw(errorDelayedFunction, "error", err)
				s.sendCentrifugoMessage(errorDelayedFunction, err)
				ticker.Stop()
			}
		}
	}()
	return err
}
