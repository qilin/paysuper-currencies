package service

import (
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_getRateStock_Ok() {
	rd, err := suite.service.getRateStock("USD", "RUB", &currencies.CorrectionRule{})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), rd.Rate, r)

	cc := float64(2)
	cr := &currencies.CorrectionRule{
		RateType:         pkg.RateTypeStock,
		CommonCorrection: cc,
	}

	ctrl := suite.service.toPrecise(r / (1 - (cc / 100)))

	rd, err = suite.service.getRateStock("USD", "RUB", cr)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), rd.Rate, ctrl)
}

// waiting for commercial oxr app_id
/*
func (suite *CurrenciesratesServiceTestSuite) TestSource_SetRatesStock_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	err = suite.CleanRatesCollection(collectionRatesNameSuffixStock)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesOxr()
	assert.NoError(suite.T(), err)

	err = suite.service.addCorrectionRule(pkg.RateTypeStock, 0, map[string]float64{}, "")
	assert.NoError(suite.T(), err)

	err = suite.service.SetRatesStock()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for _, from := range suite.config.SettlementCurrencies {
		for _, to := range suite.config.RatesRequestCurrencies {

			source := stockSource
			if from == to {
				source = stubSource
			}

			err = suite.service.getRate(pkg.RateTypeStock, from, to, bson.M{}, res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, from+to)
			assert.Equal(suite.T(), res.Source, source)

			err = suite.service.getRate(pkg.RateTypeStock, to, from, bson.M{}, res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, to+from)
			assert.Equal(suite.T(), res.Source, source)
		}
	}
}
*/
