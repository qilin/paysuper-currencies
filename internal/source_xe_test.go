package internal

import (
    "context"
    "github.com/globalsign/mgo"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/stretchr/testify/assert"
)

var usdrate = float64(64.6314)

func (suite *CurrenciesratesServiceTestSuite) TestSourceXe_ProcessRatesFailed() {
    xer := &xeRatesResponse{}
    err := suite.service.processRatesXe(xer)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorXeInvalidFrom)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceXe_ProcessRatesFailed2() {
    xer := &xeRatesResponse{
        From: "USD",
    }
    err := suite.service.processRatesXe(xer)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), errorXeNoResults)
}

func (suite *CurrenciesratesServiceTestSuite) TestSourceXe_ProcessRatesOk() {

    c1 := suite.service.getCorrectionForPair("USDRUB")
    c2 := suite.service.getCorrectionForPair("RUBUSD")

    req1 := &currencyrates.GetCurrentRateRequest{
        From: "USD",
        To:   "RUB",
    }
    req2 := &currencyrates.GetCurrentRateRequest{
        From: "RUB",
        To:   "USD",
    }
    res := &currencyrates.RateData{}

    err := suite.service.GetCurrentRate(context.TODO(), req1, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    err = suite.service.GetCurrentRate(context.TODO(), req2, res)
    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), err.Error(), mgo.ErrNotFound.Error())

    xer := &xeRatesResponse{
        From: "USD",
        To: []*xeRateItem{
            {
                Quotecurrency: "RUB",
                Mid:           usdrate,
                Inverse:       1 / usdrate,
            },
        },
    }
    err = suite.service.processRatesXe(xer)
    assert.NoError(suite.T(), err)

    err = suite.service.GetCurrentRate(context.TODO(), req1, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "USDRUB")
    assert.Equal(suite.T(), res.Rate, usdrate)
    assert.Equal(suite.T(), res.Correction, c1)
    assert.Equal(suite.T(), res.CorrectedRate, usdrate*c1)
    assert.Equal(suite.T(), res.Source, xeSource)
    assert.Equal(suite.T(), res.IsCbRate, false)

    err = suite.service.GetCurrentRate(context.TODO(), req2, res)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), res.Pair, "RUBUSD")
    assert.Equal(suite.T(), res.Rate, 1/usdrate)
    assert.Equal(suite.T(), res.Correction, c2)
    assert.Equal(suite.T(), res.CorrectedRate, (1/usdrate)*c2)
    assert.Equal(suite.T(), res.Source, xeSource)
    assert.Equal(suite.T(), res.IsCbRate, false)
}
