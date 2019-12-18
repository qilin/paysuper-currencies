package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesTcmb_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCentralbanks)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesCbtr()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for from := range suite.config.SettlementCurrenciesParsed {

		source := cbtrSource
		if from == cbtrTo {
			source = stubSource
		}

		err = suite.service.getRate(pkg.RateTypeCentralbanks, from, cbtrTo, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), from+cbtrTo, res.Pair)
		assert.Equal(suite.T(), source, res.Source)

		err = suite.service.getRate(pkg.RateTypeCentralbanks, cbtrTo, from, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), cbtrTo+from, res.Pair)
		assert.Equal(suite.T(), source, res.Source)
	}
}

