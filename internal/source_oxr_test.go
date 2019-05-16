package internal

import (
    "context"
    "github.com/globalsign/mgo"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "github.com/stretchr/testify/assert"
)

var usdrate = float64(1.4408801)

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed() {
    oxrr := &oxrResponse{}
    err := suite.service.processRatesOxr(oxrr)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorOxrInvalidFrom)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed2() {
    oxrr := &oxrResponse{
        Base: "USD",
    }
    err := suite.service.processRatesOxr(oxrr)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorOxrNoResults)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesOk() {

    req1 := &currencyrates.GetRateRequest{
        From: "USD",
        To:   "AUD",
    }
    req2 := &currencyrates.GetRateRequest{
        From: "AUD",
        To:   "USD",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req1, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    err = suite.service.GetOxrRate(context.TODO(), req2, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    oxrr := &oxrResponse{
        Base: "USD",
        Rates: map[string]float64{
            "AUD": usdrate,
        },
    }
    err = suite.service.processRatesOxr(oxrr)
    assert.NoError(suite.T(), err)

    err = suite.service.GetOxrRate(context.TODO(), req1, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDAUD")
    assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(usdrate))
    assert.Equal(suite.T(), res.Source, oxrSource)

    err = suite.service.GetOxrRate(context.TODO(), req2, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "AUDUSD")
    assert.Equal(suite.T(), res.Rate, suite.service.toPrecise(1/usdrate))
}
