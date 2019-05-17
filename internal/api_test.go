package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/pkg"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Ok() {
    req := &currencyrates.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Fail() {
    res := &currencyrates.RateData{}

    req := &currencyrates.GetRateCurrentCommonRequest{}
    err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

    req = &currencyrates.GetRateCurrentCommonRequest{
        From: "USD",
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencyrates.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: "bla-bla",
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

    req = &currencyrates.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "EUR",
        RateType: pkg.RateTypeOxr,
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Ok() {
    req := &currencyrates.GetRateByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Fail() {
    req := &currencyrates.GetRateByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Ok() {
    req := &currencyrates.GetRateCurrentForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        MerchantId: bson.NewObjectId().Hex(),
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Fail() {
    req := &currencyrates.GetRateCurrentForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Ok() {
    req := &currencyrates.GetRateByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        MerchantId: bson.NewObjectId().Hex(),
        Datetime:   ptypes.TimestampNow(),
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Fail() {
    req := &currencyrates.GetRateByDateForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.RateData{}

    err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

    req = &currencyrates.GetRateByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        MerchantId: bson.NewObjectId().Hex(),
    }

    err = suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentCommon_Ok() {
    req := &currencyrates.ExchangeCurrencyCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Ok() {
    req := &currencyrates.ExchangeCurrencyCurrentForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        Amount:     100,
        MerchantId: bson.NewObjectId().Hex(),
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Fail() {
    req := &currencyrates.ExchangeCurrencyCurrentForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Ok() {
    req := &currencyrates.ExchangeCurrencyByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Fail() {
    req := &currencyrates.ExchangeCurrencyByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Ok() {
    req := &currencyrates.ExchangeCurrencyByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        Amount:     100,
        MerchantId: bson.NewObjectId().Hex(),
        Datetime:   ptypes.TimestampNow(),
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Fail() {
    req := &currencyrates.ExchangeCurrencyByDateForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencyrates.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

    req = &currencyrates.ExchangeCurrencyByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        Amount:     100,
        MerchantId: bson.NewObjectId().Hex(),
    }

    err = suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_SetPaysuperCorrectionCorridor_Ok() {
    req := &currencyrates.CorrectionCorridor{
        Value: 0.5,
    }
    res := &currencyrates.EmptyResponse{}
    err := suite.service.SetPaysuperCorrectionCorridor(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddCommonRateCorrectionRule_Ok() {
    req1 := &currencyrates.CommonCorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddCommonRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Ok() {
    req1 := &currencyrates.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       bson.NewObjectId().Hex(),
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Fail() {
    req1 := &currencyrates.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetCommonRateCorrectionRule_Ok() {
    req1 := &currencyrates.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       bson.NewObjectId().Hex(),
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencyrates.CommonCorrectionRuleRequest{
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.CorrectionRule{}
    err = suite.service.GetCommonRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Ok() {
    merchantId := bson.NewObjectId().Hex()

    req1 := &currencyrates.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       merchantId,
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencyrates.MerchantCorrectionRuleRequest{
        RateType:   pkg.RateTypeOxr,
        MerchantId: merchantId,
    }

    res := &currencyrates.CorrectionRule{}
    err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Fail() {
    merchantId := bson.NewObjectId().Hex()

    req1 := &currencyrates.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       merchantId,
    }
    res1 := &currencyrates.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencyrates.MerchantCorrectionRuleRequest{
        RateType: pkg.RateTypeOxr,
    }

    res := &currencyrates.CorrectionRule{}
    err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}
