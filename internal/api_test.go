package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/stretchr/testify/assert"
    "time"
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
    date := time.Now().AddDate(0, 0, -1)
    created, _ := ptypes.TimestampProto(suite.service.Bod(date))

    // adding default correction rule with old date
    reqAdd := &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 1,
        CreatedAt:        created,
    }
    resAdd := &currencyrates.EmptyResponse{}
    err := suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
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
    assert.Equal(suite.T(), res2.CommonCorrection, float64(1))

    // adding special correction rule for merchant
    reqAdd = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 2,
        MerchantId:       merchantId,
        CreatedAt:        created,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
    assert.NoError(suite.T(), err)

    // and return merchant's correction rule for the same request, because it now exists
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res2.RateType, "oxr")
    assert.Equal(suite.T(), res2.MerchantId, merchantId)
    assert.Equal(suite.T(), res2.CommonCorrection, float64(2))

    // but still returns default rule if no merchant specified in rule request
    req3 := &currencyrates.CorrectionRuleRequest{
        RateType: "oxr",
    }
    res3 := &currencyrates.CorrectionRule{}
    err = suite.service.GetRateCorrectionRule(context.TODO(), req3, res3)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res3.RateType, "oxr")
    assert.Equal(suite.T(), res3.MerchantId, "")
    assert.Equal(suite.T(), res3.CommonCorrection, float64(1))

    // add more fresh default correction rule
    reqAdd = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 3,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
    assert.NoError(suite.T(), err)

    // and returns old but still actual merchant's correction rule for the merchant-specified request
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res2.RateType, "oxr")
    assert.Equal(suite.T(), res2.MerchantId, merchantId)
    assert.Equal(suite.T(), res2.CommonCorrection, float64(2))

    // but returns updated default rule if no merchant specified in rule request
    err = suite.service.GetRateCorrectionRule(context.TODO(), req3, res3)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res3.RateType, "oxr")
    assert.Equal(suite.T(), res3.MerchantId, "")
    assert.Equal(suite.T(), res3.CommonCorrection, float64(3))

    // updating special correction rule for merchant
    reqAdd = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 4,
        MerchantId:       merchantId,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
    assert.NoError(suite.T(), err)

    // and returns updatedactual merchant's correction rule for the merchant-specified request
    err = suite.service.GetRateCorrectionRule(context.TODO(), req2, res2)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res2.RateType, "oxr")
    assert.Equal(suite.T(), res2.MerchantId, merchantId)
    assert.Equal(suite.T(), res2.CommonCorrection, float64(4))

    // and updated default rule if no merchant specified in rule request
    err = suite.service.GetRateCorrectionRule(context.TODO(), req3, res3)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res3.RateType, "oxr")
    assert.Equal(suite.T(), res3.MerchantId, "")
    assert.Equal(suite.T(), res3.CommonCorrection, float64(3))

}

func (suite *CurrenciesratesServiceTestSuite) TestExchangeCurrency_Ok() {
    merchantId := bson.NewObjectId().Hex()

    req := &currencyrates.ExchangeCurrencyRequest{
        From:       "USD",
        To:         "RUB",
        MerchantId: merchantId,
        Amount:     100,
        RateType:   collectionSuffixOxr,
    }
    res := &currencyrates.ExchangeCurrencyResponse{}

    // requesting exchange
    err := suite.service.ExchangeCurrency(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))

    // adding default correction rule with old date
    reqAdd := &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: 1,
    }
    resAdd := &currencyrates.EmptyResponse{}
    err = suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
    assert.NoError(suite.T(), err)

    // requesting exchange again
    err = suite.service.ExchangeCurrency(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6528.424242))
    assert.Equal(suite.T(), res.ExchangeRate, float64(65.28424242))
    assert.Equal(suite.T(), res.Correction, float64(1))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))

    // adding special correction rule for merchant
    reqAdd = &currencyrates.CorrectionRule{
        RateType:         "oxr",
        CommonCorrection: -2,
        MerchantId:       merchantId,
    }
    err = suite.service.AddRateCorrectionRule(context.TODO(), reqAdd, resAdd)
    assert.NoError(suite.T(), err)

    // requesting exchange ones more
    err = suite.service.ExchangeCurrency(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6336.411765))
    assert.Equal(suite.T(), res.ExchangeRate, float64(63.36411765))
    assert.Equal(suite.T(), res.Correction, float64(-2))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}
