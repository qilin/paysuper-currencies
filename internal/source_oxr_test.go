package internal

import (
    "context"
    "github.com/globalsign/mgo"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/stretchr/testify/assert"
)

var usdrate = float64(64.6314)

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed() {
    oxrr := &oxrRatesResponse{}
    err := suite.service.processRatesOxr(oxrr)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorOxrInvalidFrom)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesFailed2() {
    oxrr := &oxrRatesResponse{
        Base: "USD",
    }
    err := suite.service.processRatesOxr(oxrr)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorOxrNoResults)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceOxr_ProcessRatesOk() {

    req1 := &currencyrates.GetRateRequest{
        From: "USD",
        To:   "RUB",
    }
    req2 := &currencyrates.GetRateRequest{
        From: "RUB",
        To:   "USD",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetOxrRate(context.TODO(), req1, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    err = suite.service.GetOxrRate(context.TODO(), req2, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    oxrr := &oxrRatesResponse{
        Base: "USD",
        Rates: map[string]float64{
            "RUB": usdrate,
        },
    }
    err = suite.service.processRatesOxr(oxrr)
    assert.NoError(suite.T(), err)

    err = suite.service.GetOxrRate(context.TODO(), req1, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, usdrate)
    assert.Equal(suite.T(), res.Source, oxrSource)

    err = suite.service.GetOxrRate(context.TODO(), req2, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "RUBUSD")
    assert.Equal(suite.T(), res.Rate, 1/usdrate)
}
