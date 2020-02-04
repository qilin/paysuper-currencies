package service

import (
	"context"
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jinzhu/now"
	"github.com/paysuper/paysuper-currencies/config"
	"github.com/paysuper/paysuper-database-mongo"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

var (
	r = float64(64.6314)
)

type CurrenciesratesServiceTestSuite struct {
	suite.Suite
	log     *zap.Logger
	config  *config.Config
	service *Service
}

func Test_CurrenciesratesService(t *testing.T) {
	suite.Run(t, new(CurrenciesratesServiceTestSuite))
}

func (suite *CurrenciesratesServiceTestSuite) SetupTest() {
	var err error

	suite.log, err = zap.NewProduction()
	assert.NoError(suite.T(), err)

	suite.config, err = config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	m, err := migrate.New(
		"file://../../migrations/tests",
		suite.config.MongoDsn)
	assert.NoError(suite.T(), err, "Migrate init failed")

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		suite.FailNow("Migrations failed", "%v", err)
	}

	db, err := database.NewDatabase()
	assert.NoError(suite.T(), err, "Db connection failed")

	suite.service, err = NewService(suite.config, db)
	assert.NoError(suite.T(), err, "Service creation failed")

	rates := []interface{}{
		&currencies.RateData{
			Pair:   "USDRUB",
			Rate:   r - 1,
			Source: "TEST",
			Volume: 1,
		},
		&currencies.RateData{
			Pair:   "USDRUB",
			Rate:   r,
			Source: "TEST",
			Volume: 1,
		},
	}
	err = suite.service.saveRates(collectionRatesNameSuffixOxr, rates)
	assert.NoError(suite.T(), err)
	err = suite.service.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TearDownTest() {
	if err := suite.service.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}
	suite.service.db.Close()
}

func (suite *CurrenciesratesServiceTestSuite) CleanRatesCollection(collectionSuffix string) error {
	// cleaning collection before test starts
	cName, err := suite.service.getCollectionName(collectionSuffix)
	if err != nil {
		return err
	}

	var selector interface{}
	_, err = suite.service.db.Collection(cName).RemoveAll(selector)
	if err != nil {
		return err
	}

	n, err := suite.service.db.Collection(cName).Count()
	if err != nil {
		return err
	}
	if n != 0 {
		return errors.New("Collection not empty")
	}

	return nil
}

func (suite *CurrenciesratesServiceTestSuite) TestService_CreatedOk() {
	assert.True(suite.T(), len(suite.service.cfg.RatesRequestCurrencies) > 0)
	assert.True(suite.T(), len(suite.service.cfg.SettlementCurrencies) > 0)
}

func (suite *CurrenciesratesServiceTestSuite) TestIsCurrencySupported_Ok() {
	assert.True(suite.T(), suite.service.isCurrencySupported("USD"))
}

func (suite *CurrenciesratesServiceTestSuite) TestIsCurrencySupported_Fail() {
	assert.False(suite.T(), suite.service.isCurrencySupported("BLAH"))
}

func (suite *CurrenciesratesServiceTestSuite) TestIsPairExists_Ok() {
	assert.True(suite.T(), suite.service.isPairExists("USDRUB"))
}

func (suite *CurrenciesratesServiceTestSuite) TestIsPairExists_Fail() {
	assert.False(suite.T(), suite.service.isPairExists(""))
	assert.False(suite.T(), suite.service.isPairExists("USD"))
	assert.False(suite.T(), suite.service.isPairExists("USDUSDUSD"))
	assert.False(suite.T(), suite.service.isPairExists("USDZWD"))
	assert.True(suite.T(), suite.service.isPairExists("USDUSD"))
	assert.False(suite.T(), suite.service.isPairExists("BLAALB"))
}

func (suite *CurrenciesratesServiceTestSuite) TestSaveRate_Ok() {
	rd := &currencies.RateData{
		Pair:   "USDRUB",
		Rate:   r + 1,
		Source: "TEST",
	}
	err := suite.service.saveRates(collectionRatesNameSuffixOxr, []interface{}{rd})
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetRateCorrectionRuleValue() {
	rule1 := &currencies.CorrectionRule{
		RateType:          "oxr",
		ExchangeDirection: currencies.ExchangeDirectionBuy,
	}
	assert.Equal(suite.T(), rule1.GetCorrectionValue(""), float64(0))
	assert.Equal(suite.T(), rule1.GetCorrectionValue("Blah"), float64(0))
	assert.Equal(suite.T(), rule1.GetCorrectionValue("USDEUR"), float64(0))

	rule2 := &currencies.CorrectionRule{
		RateType:          currencies.RateTypeOxr,
		ExchangeDirection: currencies.ExchangeDirectionBuy,
		CommonCorrection:  1,
	}
	assert.Equal(suite.T(), rule2.GetCorrectionValue(""), float64(1))
	assert.Equal(suite.T(), rule2.GetCorrectionValue("blah"), float64(0))
	assert.Equal(suite.T(), rule2.GetCorrectionValue("USDEUR"), float64(1))

	rule3 := &currencies.CorrectionRule{
		RateType:          "oxr",
		ExchangeDirection: currencies.ExchangeDirectionBuy,
		PairCorrection: map[string]float64{
			"USDEUR": -3,
			"EURUSD": 3,
		},
	}
	assert.Equal(suite.T(), rule3.GetCorrectionValue(""), float64(0))
	assert.Equal(suite.T(), rule3.GetCorrectionValue("blah"), float64(0))
	assert.Equal(suite.T(), rule3.GetCorrectionValue("USDEUR"), float64(-3))
	assert.Equal(suite.T(), rule3.GetCorrectionValue("EURUSD"), float64(3))
	assert.Equal(suite.T(), rule3.GetCorrectionValue("EURRUB"), float64(0))

	rule4 := &currencies.CorrectionRule{
		RateType:          currencies.RateTypeOxr,
		ExchangeDirection: currencies.ExchangeDirectionBuy,
		CommonCorrection:  1,
		PairCorrection: map[string]float64{
			"EURUSD": 3,
		},
	}
	assert.Equal(suite.T(), rule4.GetCorrectionValue(""), float64(1))
	assert.Equal(suite.T(), rule4.GetCorrectionValue("blah"), float64(0))
	assert.Equal(suite.T(), rule4.GetCorrectionValue("EURUSD"), float64(3))
	assert.Equal(suite.T(), rule4.GetCorrectionValue("EURRUB"), float64(1))
}

func (suite *CurrenciesratesServiceTestSuite) TestApplyCorrection() {
	merchantId := bson.NewObjectId().Hex()

	rd := &currencies.RateData{
		Pair:   "USDEUR",
		Rate:   suite.service.toPrecise(0.89),
		Source: "OXR",
	}

	// no correction rule set, rate unchanged
	suite.service.applyCorrection(rd, currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, merchantId)
	assert.Equal(suite.T(), rd.Rate, float64(0.89))

	// adding default correction rule for Sell
	req1 := &currencies.CommonCorrectionRule{
		RateType:          currencies.RateTypeOxr,
		ExchangeDirection: currencies.ExchangeDirectionSell,
		CommonCorrection:  1,
	}
	res1 := &currencies.EmptyResponse{}
	err := suite.service.AddCommonRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)

	// rate for Buy will be still unchanged
	suite.service.applyCorrection(rd, currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, merchantId)
	assert.Equal(suite.T(), rd.Rate, float64(0.89))

	rd2 := &currencies.RateData{
		Pair:   "USDEUR",
		Rate:   suite.service.toPrecise(0.89),
		Source: "OXR",
	}

	// rate for sell increased for 1%
	suite.service.applyCorrection(rd2, currencies.RateTypeOxr, currencies.ExchangeDirectionSell, merchantId)
	assert.Equal(suite.T(), rd2.Rate, suite.service.toPrecise(float64(0.89)/(1-(float64(1)/100))))
	assert.Equal(suite.T(), rd2.Rate, float64(0.89899))

	// adding merchant correction rule for Buy
	req2 := &currencies.CorrectionRule{
		RateType:          currencies.RateTypeOxr,
		ExchangeDirection: currencies.ExchangeDirectionBuy,
		MerchantId:        merchantId,
		CommonCorrection:  5,
		PairCorrection: map[string]float64{
			"USDEUR": 3,
		},
	}
	err = suite.service.AddMerchantRateCorrectionRule(context.TODO(), req2, res1)
	assert.NoError(suite.T(), err)

	// adding merchant correction rule for Sell
	req2 = &currencies.CorrectionRule{
		RateType:          currencies.RateTypeOxr,
		ExchangeDirection: currencies.ExchangeDirectionSell,
		MerchantId:        merchantId,
		CommonCorrection:  5,
		PairCorrection: map[string]float64{
			"USDEUR": 3,
		},
	}
	err = suite.service.AddMerchantRateCorrectionRule(context.TODO(), req2, res1)
	assert.NoError(suite.T(), err)

	rd3 := &currencies.RateData{
		Pair:   "USDEUR",
		Rate:   suite.service.toPrecise(0.89),
		Source: "OXR",
	}

	// rate for Buy decreased for 3% by pair rule for merchant
	suite.service.applyCorrection(rd3, currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, merchantId)
	assert.Equal(suite.T(), req2.GetCorrectionValue("USDEUR"), float64(3))
	assert.Equal(suite.T(), req2.GetCorrectionValue(rd3.Pair), float64(3))
	assert.Equal(suite.T(), rd3.Rate, suite.service.toPrecise(float64(0.89)/(1+(float64(3)/100))))
	assert.Equal(suite.T(), rd3.Rate, float64(0.864078))

	rd4 := &currencies.RateData{
		Pair:   "USDEUR",
		Rate:   suite.service.toPrecise(0.89),
		Source: "OXR",
	}

	// rate for Sell increased for 3% by pair rule for merchant
	suite.service.applyCorrection(rd4, currencies.RateTypeOxr, currencies.ExchangeDirectionSell, merchantId)
	assert.Equal(suite.T(), req2.GetCorrectionValue("USDEUR"), float64(3))
	assert.Equal(suite.T(), req2.GetCorrectionValue(rd4.Pair), float64(3))
	assert.Equal(suite.T(), rd4.Rate, suite.service.toPrecise(float64(0.89)/(1-(float64(3)/100))))
	assert.Equal(suite.T(), rd4.Rate, float64(0.917526))

	rd5 := &currencies.RateData{
		Pair:   "RUBUSD",
		Rate:   suite.service.toPrecise(0.89),
		Source: "OXR",
	}

	// rate increased for 5% by common rule for merchant
	suite.service.applyCorrection(rd5, currencies.RateTypeOxr, currencies.ExchangeDirectionSell, merchantId)
	assert.Equal(suite.T(), req2.GetCorrectionValue("RUBUSD"), float64(5))
	assert.Equal(suite.T(), rd5.Rate, suite.service.toPrecise(float64(0.89)/(1-(float64(5)/100))))
	assert.Equal(suite.T(), rd5.Rate, float64(0.936843))
}

func (suite *CurrenciesratesServiceTestSuite) Test_getCollectionName_Ok() {
	name, err := suite.service.getCollectionName(currencies.RateTypeOxr)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), name, "currency_rates_oxr")
}

func (suite *CurrenciesratesServiceTestSuite) Test_getCollectionName_Fail() {
	_, err := suite.service.getCollectionName("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

	_, err = suite.service.getCollectionName("Bla-bla-bla")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)
}

func (suite *CurrenciesratesServiceTestSuite) Test_addRateCorrectionRule_Ok() {

	err := suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 0, map[string]float64{}, "")
	assert.NoError(suite.T(), err)

	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 0, map[string]float64{}, bson.NewObjectId().Hex())
	assert.NoError(suite.T(), err)

	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 1, map[string]float64{}, bson.NewObjectId().Hex())
	assert.NoError(suite.T(), err)

	pairCorrection := map[string]float64{
		"USDEUR": 3,
		"EURUSD": 3,
	}
	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 1, pairCorrection, bson.NewObjectId().Hex())
	assert.NoError(suite.T(), err)

	pairCorrection = map[string]float64{
		"USDEUR": 3,
		"EURUSD": 3,
	}
	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 0, pairCorrection, "")
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_addRateCorrectionRule_Fail() {

	err := suite.service.addCorrectionRule("", currencies.ExchangeDirectionBuy, 0, map[string]float64{}, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

	err = suite.service.addCorrectionRule("bla-bla-bla", currencies.ExchangeDirectionBuy, 0, map[string]float64{}, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 101, map[string]float64{}, "")
	assert.Error(suite.T(), err)

	pairCorrection := map[string]float64{
		"USDEUR": 101,
	}
	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 0, pairCorrection, "")
	assert.Error(suite.T(), err)

	pairCorrection = map[string]float64{
		"USDEUR": 3,
		"EURZWD": 3,
	}
	err = suite.service.addCorrectionRule(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, 0, pairCorrection, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)
}

func (suite *CurrenciesratesServiceTestSuite) Test_Status_Ok() {
	status, err := suite.service.Status()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), status, serviceStatusOK)
}

func (suite *CurrenciesratesServiceTestSuite) Test_validateUrl_Ok() {
	_, err := suite.service.validateUrl("https://my-site.com/path?a=b&c=d#fragment")
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_validateUrl_Fail() {
	_, err := suite.service.validateUrl("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorEmptyUrl)

	_, err = suite.service.validateUrl("bla-bla-bla")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), "parse bla-bla-bla: invalid URI for request")
}

func (suite *CurrenciesratesServiceTestSuite) Test_getByDateQuery_Ok() {
	date := time.Now()
	query := suite.service.getByDateQuery(date)
	assert.Equal(suite.T(), query["created_at"], bson.M{"$lte": now.New(date).EndOfDay()})
}

func (suite *CurrenciesratesServiceTestSuite) Test_exchangeCurrency_Ok() {
	merchantId := bson.NewObjectId().Hex()
	res := &currencies.ExchangeCurrencyResponse{}

	// requesting exchange
	err := suite.service.exchangeCurrency(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, "USD", "RUB", 100, merchantId, bson.M{}, "", res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_exchangeCurrency_Fail() {
	res := &currencies.ExchangeCurrencyResponse{}

	err := suite.service.exchangeCurrency(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, "BLA", "USD", 100, "", bson.M{}, "", res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

	err = suite.service.exchangeCurrency(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, "USD", "", 100, "", bson.M{}, "", res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

	err = suite.service.exchangeCurrency("bla-bla", currencies.ExchangeDirectionBuy, "USD", "RUB", 100, "", bson.M{}, "", res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

	err = suite.service.exchangeCurrency(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, "USD", "EUR", 100, "", bson.M{}, "", res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) Test_exchangeCurrencyByDate_Ok() {
	merchantId := bson.NewObjectId().Hex()
	res := &currencies.ExchangeCurrencyResponse{}

	// requesting exchange
	err := suite.service.exchangeCurrencyByDate(currencies.RateTypeOxr, currencies.ExchangeDirectionBuy, "USD", "RUB", 100, merchantId, time.Now(), "", res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}
