package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-currencies/pkg"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_RequestRatesTcmb_Ok() {
	shouldBe := require.New(suite.T())

	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCentralbanks)
	shouldBe.NoError(err)

	err = suite.service.RequestRatesCbtr()
	shouldBe.NoError(err)
	err = suite.service.RequestRatesOxr()
	shouldBe.NoError(err)

	res := &currencies.RateData{}

	for from := range suite.config.SettlementCurrenciesParsed {

		source := cbtrSource
		if from == cbtrTo {
			source = stubSource
		}

		if from == cbrfTo {
			continue
		}

		err = suite.service.getRate(pkg.RateTypeCentralbanks, from, cbtrTo, bson.M{}, source, res)
		assert.NoError(suite.T(), err, "`%s` `%s` `%s`", from, cbtrTo, source)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), from+cbtrTo, res.Pair)

		err = suite.service.getRate(pkg.RateTypeCentralbanks, cbtrTo, from, bson.M{}, source, res)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), res.Rate > 0)
		assert.Equal(suite.T(), cbtrTo+from, res.Pair)
	}
}

