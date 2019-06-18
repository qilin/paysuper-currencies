package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesCbca_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCentralbanks)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesCbca()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for from := range suite.config.SettlementCurrenciesParsed {

		// these currencies are not supported by canadian central bank
		if from == "DKK" || from == "PLN" {
			continue
		}

		source := cbcaSource
		if from == cbcaTo {
			source = stubSource
		}

		err = suite.service.getRate(pkg.RateTypeCentralbanks, from, cbcaTo, bson.M{}, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, from+cbcaTo)
		assert.Equal(suite.T(), res.Source, source)

		err = suite.service.getRate(pkg.RateTypeCentralbanks, cbcaTo, from, bson.M{}, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, cbcaTo+from)
		assert.Equal(suite.T(), res.Source, source)
	}
}