package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesCbau_Ok() {

	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCentralbanks)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesCbau()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for from := range suite.config.SettlementCurrenciesParsed {

		// these currencies are not supported by australian central bank
		if from == "AUD" || from == "DKK" || from == "NOK" || from == "PLN" || from == "RUB" || from == "SEK" {
			continue
		}

		source := cbauSource
		if from == cbauTo {
			source = stubSource
		}

		err = suite.service.getRate(pkg.RateTypeCentralbanks, from, cbauTo, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, from+cbauTo)
		assert.Equal(suite.T(), res.Source, source)

		err = suite.service.getRate(pkg.RateTypeCentralbanks, cbauTo, from, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), res.Pair, cbauTo+from)
		assert.Equal(suite.T(), res.Source, source)
	}
}
