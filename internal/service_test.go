package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/config"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/paysuper/paysuper-database-mongo"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "go.uber.org/zap"
    "testing"
)

var (
    r = float64(72.3096)
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

    rates := []*currencyrates.RateData{
        {
            Pair:   "EURRUB",
            Rate:   r - 1,
            Source: "TEST",
        },
        {
            Pair:   "EURRUB",
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
    assert.True(suite.T(), suite.service.isCurrencySupported("EUR"))
}

func (suite *CurrenciesratesServiceTestSuite) TestIsCurrencySupported_Fail() {
    assert.False(suite.T(), suite.service.isCurrencySupported("BLAH"))
}

func (suite *CurrenciesratesServiceTestSuite) TestIsPairExists_Ok() {
    assert.True(suite.T(), suite.service.isPairExists("EURRUB"))
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
        Pair:   "EURRUB",
        Rate:   r + 1,
        Source: "TEST",
    }
    err := suite.service.saveRates(oxrSource, []*currencyrates.RateData{rd})
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TestSaveRate_Failed() {

    rd := &currencyrates.RateData{}
    err := suite.service.saveRates(oxrSource, []*currencyrates.RateData{rd})
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)

    rd = &currencyrates.RateData{
        Pair:   "USDEUR",
        Source: "TEST",
    }
    err = suite.service.saveRates(oxrSource, []*currencyrates.RateData{rd})
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), "Key: 'RateData.Rate' Error:Field validation for 'Rate' failed on the 'required' tag")

    rd = &currencyrates.RateData{
        Pair:   "USDZWD",
        Rate:   r,
        Source: "TEST",
    }
    err = suite.service.saveRates(oxrSource, []*currencyrates.RateData{rd})
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetOxrRate_Ok() {
    req := &currencyrates.GetRateRequest{
        From: "EUR",
        To:   "RUB",
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req, res)

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "EURRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) TestGetOxrRate_Fail() {
    res := &currencyrates.RateData{}

    req := &currencyrates.GetRateRequest{}
    err := suite.service.GetOxrRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

    req = &currencyrates.GetRateRequest{
        From: "USD",
    }
    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencyrates.GetRateRequest{
        From: "USD",
        To:   "ZWD",
    }
    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencyrates.GetRateRequest{
        From: "EUR",
        To:   "JPY",
    }
    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCentralBankRateForDate_Ok() {
    req := &currencyrates.GetCentralBankRateRequest{
        From:     "EUR",
        To:       "RUB",
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetCentralBankRateForDate(context.TODO(), req, res)

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "EURRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) TestUpdateRateOk() {
    req := &currencyrates.GetRateRequest{
        From: "EUR",
        To:   "RUB",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "EURRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")

    rd := &currencyrates.RateData{
        Pair:          "EURRUB",
        Rate:          r + 1,
        Source:        "TEST",
    }
    err = suite.service.saveRates(collectionSuffixOxr, []*currencyrates.RateData{rd})
    assert.NoError(suite.T(), err)

    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Rate, r+1)
}
