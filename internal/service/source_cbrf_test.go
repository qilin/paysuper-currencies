package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesCbrf_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCentralbanks)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesCbrf()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for from := range suite.config.SettlementCurrenciesParsed {

		source := cbrfSource
		if from == cbrfTo {
			source = stubSource
		}

		err = suite.service.getRate(pkg.RateTypeCentralbanks, from, cbrfTo, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, from+cbrfTo)
		assert.Equal(suite.T(), res.Source, source)

		err = suite.service.getRate(pkg.RateTypeCentralbanks, cbrfTo, from, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, cbrfTo+from)
		assert.Equal(suite.T(), res.Source, source)
	}
}
