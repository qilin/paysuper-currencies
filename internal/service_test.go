package internal

import (
    "context"
    "github.com/globalsign/mgo/bson"
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "go.uber.org/zap"
    "testing"
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

    settings := database.Connection{
        Host:     suite.config.MongoHost,
        Database: suite.config.MongoDatabase,
        User:     suite.config.MongoUser,
        Password: suite.config.MongoPassword,
    }
    db, err := database.NewDatabase(settings)
    assert.NoError(suite.T(), err, "Db connection failed")

    suite.service, err = NewService(suite.config, db)
    assert.NoError(suite.T(), err, "Service creation failed")

    rates := []interface{}{
        &currencyrates.RateData{
            Pair:   "USDRUB",
            Rate:   r - 1,
            Source: "TEST",
        },
        &currencyrates.RateData{
            Pair:   "USDRUB",
            Rate:   r,
            Source: "TEST",
        },
    }
    err = suite.service.saveRates(collectionSuffixOxr, rates)
    err = suite.service.saveRates(collectionSuffixCb, rates)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TearDownTest() {
    if err := suite.service.db.Drop(); err != nil {
        suite.FailNow("Database deletion failed", "%v", err)
    }
    suite.service.db.Close()
}

func (suite *CurrenciesratesServiceTestSuite) TestService_CreatedOk() {
    assert.True(suite.T(), len(suite.service.cfg.OxrSupportedCurrencies) > 0)
    assert.True(suite.T(), len(suite.service.cfg.OxrBaseCurrencies) > 0)
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
    rd := &currencyrates.RateData{
        Pair:   "USDRUB",
        Rate:   r + 1,
        Source: "TEST",
    }
    err := suite.service.saveRates(oxrSource, []interface{}{rd})
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetRateCorrectionRuleValue() {
    rule1 := &currencyrates.CorrectionRule{
        RateType: "oxr",
    }
    assert.Equal(suite.T(), rule1.GetCorrectionValue(""), float64(0))
    assert.Equal(suite.T(), rule1.GetCorrectionValue("Blah"), float64(0))
    assert.Equal(suite.T(), rule1.GetCorrectionValue("USDEUR"), float64(0))

    rule2 := &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 1,
    }
    assert.Equal(suite.T(), rule2.GetCorrectionValue(""), float64(1))
    assert.Equal(suite.T(), rule2.GetCorrectionValue("blah"), float64(0))
    assert.Equal(suite.T(), rule2.GetCorrectionValue("USDEUR"), float64(1))

    rule3 := &currencyrates.CorrectionRule{
        RateType: "oxr",
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

    rule4 := &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 1,
        PairCorrection: map[string]float64{
            "USDEUR": -3,
            "EURUSD": 3,
        },
    }
    assert.Equal(suite.T(), rule4.GetCorrectionValue(""), float64(1))
    assert.Equal(suite.T(), rule4.GetCorrectionValue("blah"), float64(0))
    assert.Equal(suite.T(), rule4.GetCorrectionValue("USDEUR"), float64(-3))
    assert.Equal(suite.T(), rule4.GetCorrectionValue("EURUSD"), float64(3))
    assert.Equal(suite.T(), rule4.GetCorrectionValue("EURRUB"), float64(1))
}

func (suite *CurrenciesratesServiceTestSuite) TestApplyCorrection() {
    merchantId := bson.NewObjectId().Hex()

    rd := &currencyrates.RateData{
        Pair:   "USDEUR",
        Rate:   suite.service.toPrecise(0.89),
        Source: "OXR",
    }

    // no correction rule set, rate unchanged
    suite.service.applyCorrection(rd, collectionSuffixOxr, merchantId)
    assert.Equal(suite.T(), rd.Rate, float64(0.89))

    // adding default correction rule
    req1 := &currencyrates.CorrectionRule{
        RateType:         collectionSuffixOxr,
        CommonCorrection: 1,
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    rd2 := &currencyrates.RateData{
        Pair:   "USDEUR",
        Rate:   suite.service.toPrecise(0.89),
        Source: "OXR",
    }

    // rate increased for 1%
    suite.service.applyCorrection(rd2, collectionSuffixOxr, merchantId)
    assert.Equal(suite.T(), req1.GetCorrectionValue("USDEUR"), float64(1))
    assert.Equal(suite.T(), rd2.Rate, suite.service.toPrecise(float64(0.89)/(1-(float64(1)/100))))
    assert.Equal(suite.T(), rd2.Rate, float64(0.898989899))

    // adding merchant correction rule
    req1 = &currencyrates.CorrectionRule{
        RateType:         collectionSuffixOxr,
        MerchantId:       merchantId,
        CommonCorrection: 5,
        PairCorrection: map[string]float64{
            "USDEUR": -3,
            "EURUSD": 3,
        },
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    rd3 := &currencyrates.RateData{
        Pair:   "USDEUR",
        Rate:   suite.service.toPrecise(0.89),
        Source: "OXR",
    }

    // rate decreased for 3%
    suite.service.applyCorrection(rd3, collectionSuffixOxr, merchantId)
    assert.Equal(suite.T(), req1.GetCorrectionValue("USDEUR"), float64(-3))
    assert.Equal(suite.T(), req1.GetCorrectionValue(rd3.Pair), float64(-3))
    assert.Equal(suite.T(), rd3.Rate, suite.service.toPrecise(float64(0.89)/(1-(float64(-3)/100))))
    assert.Equal(suite.T(), rd3.Rate, float64(0.8640776699))

    rd4 := &currencyrates.RateData{
        Pair:   "EURUSD",
        Rate:   suite.service.toPrecise(1.12),
        Source: "OXR",
    }
    // rate increased for 3%
    suite.service.applyCorrection(rd4, collectionSuffixOxr, merchantId)
    assert.Equal(suite.T(), req1.GetCorrectionValue("EURUSD"), float64(3))
    assert.Equal(suite.T(), rd4.Rate, suite.service.toPrecise(float64(1.12)/(1-(float64(3)/100))))

    rd5 := &currencyrates.RateData{
        Pair:   "RUBUSD",
        Rate:   suite.service.toPrecise(0.015),
        Source: "OXR",
    }
    // rate increased for 5%
    suite.service.applyCorrection(rd5, collectionSuffixOxr, merchantId)
    assert.Equal(suite.T(), req1.GetCorrectionValue("RUBUSD"), float64(5))
    assert.Equal(suite.T(), rd5.Rate, suite.service.toPrecise(float64(0.015)/(1-(float64(5)/100))))
}
