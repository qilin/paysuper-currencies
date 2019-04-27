package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/config"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/paysuper/paysuper-currencies-rates/utils"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "go.uber.org/zap"
    "testing"
)

var (
    r = float64(72.3096)
    c = float64(1)
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

    session, err := mgo.Dial(utils.GetMongoUrl(suite.config))
    assert.NoError(suite.T(), err, "Db connection failed")

    suite.service, err = NewService(suite.config, session.DB(suite.config.MongoDatabase))
    assert.NoError(suite.T(), err, "Service creation failed")

    rd := &currencyrates.RateData{
        Pair:          "EURRUB",
        Rate:          r,
        Correction:    c,
        CorrectedRate: r * c,
        Source:        "TEST",
        IsCbRate:      false,
    }
    err = suite.service.saveRate(rd)
    assert.NoError(suite.T(), err)

    rd = &currencyrates.RateData{
        Pair:          "EURRUB",
        Rate:          r - 1,
        Correction:    c,
        CorrectedRate: (r - 1) * c,
        Source:        "TEST",
        IsCbRate:      true,
    }
    err = suite.service.saveRate(rd)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TearDownTest() {
    if err := suite.service.db.DropDatabase(); err != nil {
        suite.FailNow("Database deletion failed", "%v", err)
    }
    suite.service.db.Session.Close()
}

func (suite *CurrenciesratesServiceTestSuite) TestService_CreatedOk() {
    assert.True(suite.T(), len(suite.service.cfg.XeSupportedCurrencies) > 0)
    assert.True(suite.T(), len(suite.service.cfg.XeBaseCurrencies) > 0)
    assert.True(suite.T(), suite.service.cfg.XeCommonCorrection > 0)
    assert.True(suite.T(), suite.service.cfg.XeCommonCorrection != float64(1))
    assert.True(suite.T(), len(suite.service.cfg.XeAuthCredentials) > 0)
    assert.True(suite.T(), len(suite.service.cfg.XePairsCorrections) > 0)
    assert.True(suite.T(), suite.service.cfg.XeRatesRequestPeriod > 0)
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
    assert.False(suite.T(), suite.service.isPairExists("USDUSD"))
    assert.False(suite.T(), suite.service.isPairExists("BLAALB"))
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCorrectionForPair_UnexistingOk() {
    assert.Equal(suite.T(), suite.service.getCorrectionForPair(""), float64(1))
    assert.Equal(suite.T(), suite.service.getCorrectionForPair("BLAHBLAH"), float64(1))
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCorrectionForPair_CommonOk() {
    c := suite.service.getCorrectionForPair("EURUSD")
    cc := suite.service.cfg.XeCommonCorrection
    assert.Equal(suite.T(), c, cc)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCorrectionForPair_PairsOk() {
    cc := suite.service.cfg.XeCommonCorrection
    for k, v := range suite.service.cfg.XePairsCorrections {
        c := suite.service.getCorrectionForPair(k)
        assert.Equal(suite.T(), c, v)
        assert.NotEqual(suite.T(), c, cc)
    }
}

func (suite *CurrenciesratesServiceTestSuite) TestSaveRate_Ok() {
    rd := &currencyrates.RateData{
        Pair:          "EURRUB",
        Rate:          r + 1,
        Correction:    c,
        CorrectedRate: (r + 1) * c,
        Source:        "TEST",
        IsCbRate:      false,
    }
    err := suite.service.saveRate(rd)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TestSaveRate_Failed() {

    rd := &currencyrates.RateData{}
    err := suite.service.saveRate(rd)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)

    rd = &currencyrates.RateData{
        Pair:     "USDEUR",
        Source:   "TEST",
        IsCbRate: false,
    }
    err = suite.service.saveRate(rd)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), "Key: 'RateData.Rate' Error:Field validation for 'Rate' failed on the 'required' tag\nKey: 'RateData.Correction' Error:Field validation for 'Correction' failed on the 'required' tag\nKey: 'RateData.CorrectedRate' Error:Field validation for 'CorrectedRate' failed on the 'required' tag")

    rd = &currencyrates.RateData{
        Pair:          "USDZWD",
        Rate:          r,
        Correction:    c,
        CorrectedRate: r * c,
        Source:        "TEST",
        IsCbRate:      false,
    }
    err = suite.service.saveRate(rd)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCurrentRate_Ok() {
    req := &currencyrates.GetCurrentRateRequest{
        From: "EUR",
        To:   "RUB",
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetCurrentRate(context.TODO(), req, res)

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "EURRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Correction, c)
    assert.Equal(suite.T(), res.CorrectedRate, r*c)
    assert.Equal(suite.T(), res.Source, "TEST")
    assert.Equal(suite.T(), res.IsCbRate, false)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCurrentRate_Fail() {
    res := &currencyrates.RateData{}

    req := &currencyrates.GetCurrentRateRequest{}
    err := suite.service.GetCurrentRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

    req = &currencyrates.GetCurrentRateRequest{
        From: "USD",
    }
    err = suite.service.GetCurrentRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencyrates.GetCurrentRateRequest{
        From: "USD",
        To:   "ZWD",
    }
    err = suite.service.GetCurrentRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencyrates.GetCurrentRateRequest{
        From: "EUR",
        To:   "JPY",
    }
    err = suite.service.GetCurrentRate(context.TODO(), req, res)
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
    assert.Equal(suite.T(), res.Rate, r-1)
    assert.Equal(suite.T(), res.Correction, c)
    assert.Equal(suite.T(), res.CorrectedRate, (r-1)*c)
    assert.Equal(suite.T(), res.Source, "TEST")
    assert.Equal(suite.T(), res.IsCbRate, true)
}

func (suite *CurrenciesratesServiceTestSuite) TestUpdateRateOk() {
    req := &currencyrates.GetCurrentRateRequest{
        From: "EUR",
        To:   "RUB",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetCurrentRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "EURRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Correction, c)
    assert.Equal(suite.T(), res.CorrectedRate, r*c)
    assert.Equal(suite.T(), res.Source, "TEST")
    assert.Equal(suite.T(), res.IsCbRate, false)

    rd := &currencyrates.RateData{
        Pair:          "EURRUB",
        Rate:          r + 1,
        Correction:    c,
        CorrectedRate: (r + 1) * c,
        Source:        "TEST",
        IsCbRate:      false,
    }
    err = suite.service.saveRate(rd)
    assert.NoError(suite.T(), err)

    err = suite.service.GetCurrentRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Rate, r+1)
    assert.Equal(suite.T(), res.CorrectedRate, (r+1)*c)
}
