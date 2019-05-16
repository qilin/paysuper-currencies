package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestGetOxrRate_Ok() {
    req := &currencyrates.GetRateRequest{
        From: "USD",
        To:   "RUB",
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req, res)

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
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
        From: "USD",
        To:   "JPY",
    }
    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) TestGetCentralBankRateForDate_Ok() {
    req := &currencyrates.GetCentralBankRateRequest{
        From:     "USD",
        To:       "RUB",
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetCentralBankRateForDate(context.TODO(), req, res)

    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) TestUpdateRateOk() {
    req := &currencyrates.GetRateRequest{
        From: "USD",
        To:   "RUB",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")

    rd := &currencyrates.RateData{
        Pair:   "USDRUB",
        Rate:   r + 1,
        Source: "TEST",
    }
    err = suite.service.saveRates(collectionSuffixOxr, []interface{}{rd})
    assert.NoError(suite.T(), err)

    err = suite.service.GetOxrRate(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Rate, r+1)
}

func (suite *CurrenciesratesServiceTestSuite) TestAddRateCorrectionRuleOk() {
    res := &currencyrates.EmptyResponse{}

    req := &currencyrates.CorrectionRule{
        RateType: "oxr",
    }
    err := suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType:   "oxr",
        MerchantId: bson.NewObjectId().Hex(),
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        MerchantId:       bson.NewObjectId().Hex(),
        CommonCorrection: 1,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        MerchantId:       bson.NewObjectId().Hex(),
        CommonCorrection: 1,
        PairCorrection: map[string]float64{
            "USDEUR": -3,
            "EURUSD": 3,
        },
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType: "oxr",
        PairCorrection: map[string]float64{
            "USDEUR": -3,
            "EURUSD": 3,
        },
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) TestAddRateCorrectionRuleFail() {
    res := &currencyrates.EmptyResponse{}

    req := &currencyrates.CorrectionRule{}
    err := suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType: "bla-bla-bla",
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType:   "oxr",
        MerchantId: "bla-bla-bla",
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 101,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType: "oxr",
        PairCorrection: map[string]float64{
            "USDEUR": -101,
        },
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRule{
        RateType: "oxr",
        PairCorrection: map[string]float64{
            "USDEUR": -3,
            "EURZWD": 3,
        },
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorCurrencyPairNotExists)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetRateCorrectionRuleFail() {
    res := &currencyrates.CorrectionRule{}

    req := &currencyrates.CorrectionRuleRequest{}
    err := suite.service.GetRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRuleRequest{
        RateType: "bla-bla-bla",
    }
    err = suite.service.GetRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRuleRequest{
        RateType:   "oxr",
        MerchantId: "bla-bla-bla",
    }
    err = suite.service.GetRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)

    req = &currencyrates.CorrectionRuleRequest{
        RateType:   "oxr",
        MerchantId: bson.NewObjectId().Hex(),
    }
    err = suite.service.GetRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) TestGetRateCorrectionRuleOk() {

    req1 := &currencyrates.CorrectionRule{
        RateType: "oxr",
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req2 := &currencyrates.CorrectionRuleRequest{
        RateType: "oxr",
    }
    res2 := &currencyrates.CorrectionRule{}
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)

    assert.Equal(suite.T(), res2.RateType, req1.RateType)
}

func (suite *CurrenciesratesServiceTestSuite) TestGetRateCorrectionRuleWithFallback() {
    merchantId := bson.NewObjectId().Hex()

    // adding default correction rule
    req1 := &currencyrates.CorrectionRule{
        RateType: "oxr",
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    // falling back to default correction rule, while requesting it for merchant,
    // if no rule for merchant is specified
    req2 := &currencyrates.CorrectionRuleRequest{
        RateType:   "oxr",
        MerchantId: merchantId,
    }
    res2 := &currencyrates.CorrectionRule{}
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res2.RateType, "oxr")
    assert.Equal(suite.T(), res2.MerchantId, "")

    // adding special correction rule for merchant
    req1 = &currencyrates.CorrectionRule{
        RateType:   "oxr",
        MerchantId: merchantId,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    // and return merchant's correction rule for the same request, because it now exists
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res2.RateType, "oxr")
    assert.Equal(suite.T(), res2.MerchantId, merchantId)

    // but still returns default rule if no merchant specified in rule request
    req3 := &currencyrates.CorrectionRuleRequest{
        RateType: "oxr",
    }
    res3 := &currencyrates.CorrectionRule{}
    err = suite.service.GetRateCorrectionRule(context.TODO(), req3, res3)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res3.RateType, "oxr")
    assert.Equal(suite.T(), res3.MerchantId, "")
}
