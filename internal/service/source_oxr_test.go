package service

import (
	"context"
	"github.com/globalsign/mgo"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/stretchr/testify/assert"
)

var usdrate = float64(1.4408801)

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed() {
	oxrr := &oxrResponse{}
	_, err := suite.service.processRatesOxr(oxrr)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorOxrInvalidFrom)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed2() {
	oxrr := &oxrResponse{
		Base: "USD",
	}
	_, err := suite.service.processRatesOxr(oxrr)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), errorOxrNoResults)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesOk() {

	req1 := &currencies.GetRateCurrentCommonRequest{
		From:     "USD",
		To:       "AUD",
		RateType: currencies.RateTypeOxr,
	}
	req2 := &currencies.GetRateCurrentCommonRequest{
		From:     "AUD",
		To:       "USD",
		RateType: currencies.RateTypeOxr,
	}
	res := &currencies.RateData{}

	err := suite.service.GetRateCurrentCommon(context.TODO(), req1, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

	err = suite.service.GetRateCurrentCommon(context.TODO(), req2, res)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

	oxrr := &oxrResponse{
		Base: "USD",
		Rates: map[string]float64{
			"AUD": usdrate,
		},
	}
	rates, err := suite.service.processRatesOxr(oxrr)
	assert.NoError(suite.T(), err)

	err = suite.service.saveRates(collectionRatesNameSuffixOxr, rates)
	assert.NoError(suite.T(), err)

	oxrr = &oxrResponse{
		Base: "AUD",
		Rates: map[string]float64{
			"USD": 1 / usdrate,
		},
	}
	rates, err = suite.service.processRatesOxr(oxrr)
	assert.NoError(suite.T(), err)

	err = suite.service.saveRates(collectionRatesNameSuffixOxr, rates)
	assert.NoError(suite.T(), err)

	err = suite.service.GetRateCurrentCommon(context.TODO(), req1, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "USDAUD")
	assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(usdrate))
	assert.Equal(suite.T(), res.Source, oxrSource)

	err = suite.service.GetRateCurrentCommon(context.TODO(), req2, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Pair, "AUDUSD")
	assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(1/usdrate))
}

// waiting for commercial oxr app_id
/*
func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesOxr_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesOxr()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for _, from := range suite.config.SettlementCurrencies {
		for _, to := range suite.config.RatesRequestCurrencies {

			source := cbrfSource
			if from == to {
				source = stubSource
			}

			err = suite.service.getRate(pkg.RateTypeCentralbanks, from, to, bson.M{}, cbrfSource, res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, from+to)
			assert.Equal(suite.T(), res.Source, source)

			err = suite.service.getRate(pkg.RateTypeCentralbanks, to, from, bson.M{}, cbrfSource,  res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, to+from)
			assert.Equal(suite.T(), res.Source, source)
		}

	}
}
*/
