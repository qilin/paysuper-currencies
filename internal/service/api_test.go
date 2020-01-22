package service

import (
	"context"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/stretchr/testify/assert"
)

var (
	supportedCurrencies = []string{"AED", "ALL", "AMD", "ARS", "AUD", "BHD", "BRL", "BYN", "CAD", "CHF", "CLP", "CNY",
		"COP", "DKK", "EGP", "EUR", "GBP", "GHS", "HKD", "IDR", "ILS", "INR", "ISK", "JPY", "KES", "KRW", "KWD", "KZT",
		"MXN", "MYR", "NOK", "NZD", "PEN", "PHP", "PLN", "QAR", "RSD", "RUB", "SAR", "SEK", "SGD", "THB", "TWD",
		"TZS", "UAH", "USD", "UYU", "VND", "ZAR", "BGN", "HUF", "RON", "HRK", "CZK", "TRY"}
	settlementCurrencies = []string{"USD", "EUR", "RUB", "GBP"}
	priceCurrencies      = []string{"AED", "ARS", "AUD", "BHD", "BRL", "CAD", "CHF", "CLP", "CNY", "COP", "CZK",
		"DKK", "EGP", "EUR", "GBP", "HKD", "HRK", "HUF", "IDR", "ILS", "INR", "JPY", "KRW", "KZT", "MXN", "MYR", "NOK",
		"NZD", "PEN", "PHP", "PLN", "QAR", "RON", "RSD", "RUB", "SAR", "SEK", "SGD", "THB", "TWD", "USD", "VND",
		"ZAR", "TRY"}
	vatCurrencies = []string{"ALL", "AMD", "AUD", "BHD", "BRL", "BYN", "CAD", "CHF", "EGP", "EUR", "GBP", "GHS",
		"ILS", "ISK", "JPY", "KES", "KRW", "PLN", "RSD", "RUB", "SGD", "TZS", "USD", "TRY"}
	accountingCurrencies = []string{"USD", "EUR", "RUB", "GBP"}
)

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Ok() {

	// common RateTypes
	req := &currenciespb.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r)
	assert.Equal(suite.T(), res.Source, "TEST")

	// and for the same currencies
	req = &currenciespb.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "USD",
		RateType: currenciespb.RateTypeOxr,
	}

	err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDUSD")
	assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(1))
	assert.Equal(suite.T(), res.Source, stubSource)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_CbFallback_Ok() {
	// fallback for centralbanks
	res := &currenciespb.RateData{}

	req := &currenciespb.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeCentralbanks,
		Source:   cbrfSource,
	}

	err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r)
	assert.Equal(suite.T(), res.Source, "TEST")

	rd := &currenciespb.RateData{
		Pair:   "USDRUB",
		Rate:   r + 1,
		Source: cbrfSource,
	}
	err = suite.service.saveRates(collectionRatesNameSuffixCentralbanks, []interface{}{rd})
	assert.NoError(suite.T(), err)

	err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r+1)
	assert.Equal(suite.T(), res.Source, cbrfSource)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentCommon_Fail() {
	res := &currenciespb.RateData{}

	req := &currenciespb.GetRateCurrentCommonRequest{}
	err := suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorFromCurrencyNotSupported)

	req = &currenciespb.GetRateCurrentCommonRequest{
		From: "USD",
	}
	err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorToCurrencyNotSupported)

	req = &currenciespb.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: "bla-bla",
	}
	err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorRateTypeInvalid)

	req = &currenciespb.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "EUR",
		RateType: currenciespb.RateTypeOxr,
	}
	err = suite.service.GetRateCurrentCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Ok() {
	// common RateTypes
	req := &currenciespb.GetRateByDateCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Datetime: ptypes.TimestampNow(),
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r)
	assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateCommon_Fail() {
	req := &currenciespb.GetRateByDateCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateByDateCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Ok() {
	// for common RateTypes
	req := &currenciespb.GetRateCurrentForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		MerchantId: bson.NewObjectId().Hex(),
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r)
	assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateCurrentForMerchant_Fail() {
	req := &currenciespb.GetRateCurrentForMerchantRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateCurrentForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Ok() {
	req := &currenciespb.GetRateByDateForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		MerchantId: bson.NewObjectId().Hex(),
		Datetime:   ptypes.TimestampNow(),
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDRUB")
	assert.Equal(suite.T(), res.Rate, r)
	assert.Equal(suite.T(), res.Source, "TEST")
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetRateByDateForMerchant_Fail() {
	req := &currenciespb.GetRateByDateForMerchantRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
	}

	res := &currenciespb.RateData{}

	err := suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

	req = &currenciespb.GetRateByDateForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		MerchantId: bson.NewObjectId().Hex(),
	}

	err = suite.service.GetRateByDateForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentCommon_Ok() {
	req := &currenciespb.ExchangeCurrencyCurrentCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))

	req = &currenciespb.ExchangeCurrencyCurrentCommonRequest{
		From:     "USD",
		To:       "USD",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
	}

	err = suite.service.ExchangeCurrencyCurrentCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(100))
	assert.Equal(suite.T(), res.ExchangeRate, float64(1))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(1))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Ok() {
	req := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		Amount:     100,
		MerchantId: bson.NewObjectId().Hex(),
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyCurrentForMerchant_Fail() {
	req := &currenciespb.ExchangeCurrencyCurrentForMerchantRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyCurrentForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Ok() {
	req := &currenciespb.ExchangeCurrencyByDateCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
		Datetime: ptypes.TimestampNow(),
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateCommon_Fail() {
	req := &currenciespb.ExchangeCurrencyByDateCommonRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyByDateCommon(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Ok() {
	req := &currenciespb.ExchangeCurrencyByDateForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		Amount:     100,
		MerchantId: bson.NewObjectId().Hex(),
		Datetime:   ptypes.TimestampNow(),
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.ExchangedAmount, float64(6463.14))
	assert.Equal(suite.T(), res.ExchangeRate, float64(64.6314))
	assert.Equal(suite.T(), res.Correction, float64(0))
	assert.Equal(suite.T(), res.OriginalRate, float64(64.6314))
}

func (suite *CurrenciesratesServiceTestSuite) Test_ExchangeCurrencyByDateForMerchant_Fail() {
	req := &currenciespb.ExchangeCurrencyByDateForMerchantRequest{
		From:     "USD",
		To:       "RUB",
		RateType: currenciespb.RateTypeOxr,
		Amount:   100,
	}

	res := &currenciespb.ExchangeCurrencyResponse{}

	err := suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)

	req = &currenciespb.ExchangeCurrencyByDateForMerchantRequest{
		From:       "USD",
		To:         "RUB",
		RateType:   currenciespb.RateTypeOxr,
		Amount:     100,
		MerchantId: bson.NewObjectId().Hex(),
	}

	err = suite.service.ExchangeCurrencyByDateForMerchant(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorTimestampRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddCommonRateCorrectionRule_Ok() {
	req1 := &currenciespb.CommonCorrectionRule{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		CommonCorrection:  1,
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddCommonRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Ok() {
	req1 := &currenciespb.CorrectionRule{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		CommonCorrection:  1,
		MerchantId:        bson.NewObjectId().Hex(),
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_AddMerchantRateCorrectionRule_Fail() {
	req1 := &currenciespb.CorrectionRule{
		RateType:         currenciespb.RateTypeOxr,
		CommonCorrection: 1,
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetCommonRateCorrectionRule_Ok() {
	req1 := &currenciespb.CommonCorrectionRule{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		CommonCorrection:  1,
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddCommonRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)

	req := &currenciespb.CommonCorrectionRuleRequest{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
	}

	res := &currenciespb.CorrectionRule{}
	err = suite.service.GetCommonRateCorrectionRule(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Ok() {
	merchantId := bson.NewObjectId().Hex()

	req1 := &currenciespb.CorrectionRule{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		CommonCorrection:  1,
		MerchantId:        merchantId,
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)

	req := &currenciespb.MerchantCorrectionRuleRequest{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		MerchantId:        merchantId,
	}

	res := &currenciespb.CorrectionRule{}
	err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), res)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetMerchantRateCorrectionRule_Fail() {
	merchantId := bson.NewObjectId().Hex()

	req1 := &currenciespb.CorrectionRule{
		RateType:          currenciespb.RateTypeOxr,
		ExchangeDirection: currenciespb.ExchangeDirectionSell,
		CommonCorrection:  1,
		MerchantId:        merchantId,
	}
	res1 := &currenciespb.EmptyResponse{}
	err := suite.service.AddMerchantRateCorrectionRule(context.TODO(), req1, res1)
	assert.NoError(suite.T(), err)

	req := &currenciespb.MerchantCorrectionRuleRequest{
		RateType: currenciespb.RateTypeOxr,
	}

	res := &currenciespb.CorrectionRule{}
	err = suite.service.GetMerchantRateCorrectionRule(context.TODO(), req, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorMerchantIdRequired)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetSupportedCurrencies_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesList{}
	err := suite.service.GetSupportedCurrencies(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Currencies, suite.config.SupportedCurrencies)
	assert.ElementsMatch(suite.T(), res.Currencies, supportedCurrencies)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetSettlementCurrencies_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesList{}
	err := suite.service.GetSettlementCurrencies(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Currencies, suite.config.SettlementCurrencies)
	assert.ElementsMatch(suite.T(), res.Currencies, settlementCurrencies)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetPriceCurrencies_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesList{}
	err := suite.service.GetPriceCurrencies(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Currencies, suite.config.PriceCurrencies)
	assert.ElementsMatch(suite.T(), res.Currencies, priceCurrencies)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetVatCurrencies_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesList{}
	err := suite.service.GetVatCurrencies(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Currencies, suite.config.VatCurrencies)
	assert.ElementsMatch(suite.T(), res.Currencies, vatCurrencies)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetAccountingCurrencies_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesList{}
	err := suite.service.GetAccountingCurrencies(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Currencies, suite.config.AccountingCurrencies)
	assert.ElementsMatch(suite.T(), res.Currencies, accountingCurrencies)
}

func (suite *CurrenciesratesServiceTestSuite) Test_GetCurrenciesPrecision_Ok() {
	req := &currenciespb.EmptyRequest{}
	res := &currenciespb.CurrenciesPrecisionResponse{}
	err := suite.service.GetCurrenciesPrecision(context.TODO(), req, res)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), res.Values["RUB"], 2)
	assert.EqualValues(suite.T(), res.Values["CLP"], 0)
	assert.EqualValues(suite.T(), res.Values["BHD"], 2)
}
