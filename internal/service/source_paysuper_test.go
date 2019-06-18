package service

import (
	"github.com/globalsign/mgo"
	"github.com/stretchr/testify/assert"
	"time"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_getRatePaysuper_Ok() {
	_, err := suite.service.getRatePaysuper("USD", "RUB")
	assert.Error(suite.T(), err, mgo.ErrNotFound)

	corrections := []interface{}{&paysuperCorrection{
		Pair:      "USDRUB",
		Value:     1.5,
		CreatedAt: time.Now(),
	}}
	err = suite.service.db.Collection(collectionNamePaysuperCorrections).Insert(corrections...)
	assert.NoError(suite.T(), err)

	rd, err := suite.service.getRatePaysuper("USD", "RUB")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), rd.Rate, r+1.5)
}

// waiting for commercial oxr app_id
/*
func (suite *CurrenciesratesServiceTestSuite) TestSource_SetRatesPaysuper_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	err = suite.CleanRatesCollection(collectionRatesNameSuffixPaysuper)
	assert.NoError(suite.T(), err)

	err = suite.service.RequestRatesOxr()
	assert.NoError(suite.T(), err)

	corrections := []interface{}{}
	for _, from := range suite.config.SettlementCurrencies {
		for _, to := range suite.config.RatesRequestCurrencies {
			corrections = append(corrections, &paysuperCorrection{
				Pair:      from + to,
				Value:     1,
				CreatedAt: time.Now(),
			})
			corrections = append(corrections, &paysuperCorrection{
				Pair:      to + from,
				Value:     1,
				CreatedAt: time.Now(),
			})
		}
	}
	err = suite.service.db.Collection(collectionNamePaysuperCorrections).Insert(corrections...)
	assert.NoError(suite.T(), err)

	err = suite.service.SetRatesPaysuper()
	assert.NoError(suite.T(), err)

	res := &currencies.RateData{}

	for _, from := range suite.config.SettlementCurrencies {
		for _, to := range suite.config.RatesRequestCurrencies {

			source := paysuperSource
			if from == to {
				source = stubSource
			}

			err = suite.service.getRate(pkg.RateTypePaysuper, from, to, bson.M{}, res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, from+to)
			assert.Equal(suite.T(), res.Source, source)

			err = suite.service.getRate(pkg.RateTypePaysuper, to, from, bson.M{}, res)
			assert.NoError(suite.T(), err)
			assert.True(suite.T(), res.Rate > 0)
			assert.Equal(suite.T(), res.Pair, to+from)
			assert.Equal(suite.T(), res.Source, source)
		}

	}
}
*/