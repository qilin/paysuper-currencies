package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-currencies/pkg"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func (suite *CurrenciesratesServiceTestSuite) TestSource_SetRatesCardpay_Ok() {
	// cleaning collection before test starts
	err := suite.CleanRatesCollection(collectionRatesNameSuffixCardpay)
	assert.NoError(suite.T(), err)

	from := "USD"
	to := "EUR"
	rate1 := float64(0.9)
	volume1 := float64(1)
	rate2 := float64(0.83)
	volume2 := float64(10)
	source := "VISA"

	res := &currencies.RateData{}

	err = suite.service.getRate(pkg.RateTypeCardpay, from, to, bson.M{}, res)
	assert.Error(suite.T(), err)

	msg := &currencies.CardpayRate{
		From:      from,
		To:        to,
		Rate:      rate1,
		Volume:    volume1,
		CreatedAt: ptypes.TimestampNow(),
		Source:    source,
	}

	dlv := amqp.Delivery{}

	err = suite.service.SetRatesCardpay(msg, dlv)
	assert.NoError(suite.T(), err)

	err = suite.service.getRate(pkg.RateTypeCardpay, from, to, bson.M{}, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(rate1))
	assert.Equal(suite.T(), res.Pair, from+to)
	assert.NotEqual(suite.T(), res.Source, source)
	assert.Equal(suite.T(), res.Source, cardpaySource)

	msg = &currencies.CardpayRate{
		From:      from,
		To:        to,
		Rate:      rate2,
		Volume:    volume2,
		CreatedAt: ptypes.TimestampNow(),
		Source:    source,
	}

	err = suite.service.SetRatesCardpay(msg, dlv)
	assert.NoError(suite.T(), err)

	midRate := suite.service.toPrecise(float64((rate1*volume1 + rate2*volume2) / (volume1 + volume2)))
	midRateCtrl := suite.service.toPrecise(float64(0.8363636364))

	err = suite.service.getRate(pkg.RateTypeCardpay, from, to, bson.M{}, res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res.Rate, midRate)
	assert.Equal(suite.T(), res.Rate, midRateCtrl)
	assert.Equal(suite.T(), res.Pair, from+to)
	assert.NotEqual(suite.T(), res.Source, source)
	assert.Equal(suite.T(), res.Source, cardpaySource)
}
