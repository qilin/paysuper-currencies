package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies/pkg"
    "github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
    "github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Ok() {
    req := &currencies.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Fail() {
    res := &currencies.RateData{}

    req := &currencies.GetRateCurrentCommonRequest{}
    err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

    req = &currencies.GetRateCurrentCommonRequest{
        From: "USD",
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

    req = &currencies.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: "bla-bla",
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

    req = &currencies.GetRateCurrentCommonRequest{
        From:     "USD",
        To:       "EUR",
        RateType: pkg.RateTypeOxr,
    }
    err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Ok() {
    req := &currencies.GetRateByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Fail() {
    req := &currencies.GetRateByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Ok() {
    req := &currencies.GetRateCurrentForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        MerchantId: bson.NewObjectId().Hex(),
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Fail() {
    req := &currencies.GetRateCurrentForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Ok() {
    req := &currencies.GetRateByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        MerchantId: bson.NewObjectId().Hex(),
        Datetime:   ptypes.TimestampNow(),
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, r)
    assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Fail() {
    req := &currencies.GetRateByDateForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.RateData{}

    err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

    req = &currencies.GetRateByDateForMerchantRequest{
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
    req := &currencies.ExchangeCurrencyCurrentCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Ok() {
    req := &currencies.ExchangeCurrencyCurrentForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        Amount:     100,
        MerchantId: bson.NewObjectId().Hex(),
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Fail() {
    req := &currencies.ExchangeCurrencyCurrentForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Ok() {
    req := &currencies.ExchangeCurrencyByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
        Datetime: ptypes.TimestampNow(),
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Fail() {
    req := &currencies.ExchangeCurrencyByDateCommonRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Ok() {
    req := &currencies.ExchangeCurrencyByDateForMerchantRequest{
        From:       "USD",
        To:         "RUB",
        RateType:   pkg.RateTypeOxr,
        Amount:     100,
        MerchantId: bson.NewObjectId().Hex(),
        Datetime:   ptypes.TimestampNow(),
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
    assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
    assert.Equal(suite.T(), res.Correction, float64(0))
    assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Fail() {
    req := &currencies.ExchangeCurrencyByDateForMerchantRequest{
        From:     "USD",
        To:       "RUB",
        RateType: pkg.RateTypeOxr,
        Amount:   100,
    }

    res := &currencies.ExchangeCurrencyResponse{}

    err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

    req = &currencies.ExchangeCurrencyByDateForMerchantRequest{
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
    req := &currencies.CorrectionCorridor{
        Value: 0.5,
    }
    res := &currencies.EmptyResponse{}
    err := suite.service.SetPaysuperCorrectionCorridor(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddCommonRateCorrectionRule_Ok() {
    req1 := &currencies.CommonCorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddCommonRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Ok() {
    req1 := &currencies.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       bson.NewObjectId().Hex(),
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Fail() {
    req1 := &currencies.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetCommonRateCorrectionRule_Ok() {
    req1 := &currencies.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       bson.NewObjectId().Hex(),
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencies.CommonCorrectionRuleRequest{
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.CorrectionRule{}
    err = suite.service.GetCommonRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Ok() {
    merchantId := bson.NewObjectId().Hex()

    req1 := &currencies.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       merchantId,
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencies.MerchantCorrectionRuleRequest{
        RateType:   pkg.RateTypeOxr,
        MerchantId: merchantId,
    }

    res := &currencies.CorrectionRule{}
    err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Fail() {
    merchantId := bson.NewObjectId().Hex()

    req1 := &currencies.CorrectionRule{
        RateType:         pkg.RateTypeOxr,
        CommonCorrection: 1,
        MerchantId:       merchantId,
    }
    res1 := &currencies.EmptyResponse{}
    err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
    assert.NoError(suite.T(), err)

    req := &currencies.MerchantCorrectionRuleRequest{
        RateType: pkg.RateTypeOxr,
    }

    res := &currencies.CorrectionRule{}
    err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}
