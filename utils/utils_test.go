package utils

import (
    "github.com/paysuper/paysuper-currencies-rates/config"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "go.uber.org/zap"
    "testing"
)

var (
    mhost = "127.0.0.1"
    mdb   = "currenciesrates"
    muser = "user"
    mpass = "pass"
)

type CurrenciesratesMainTestSuite struct {
    suite.Suite
    log    *zap.Logger
    config *config.Config
}

func Test_CurrenciesratesMain(t *testing.T) {
    suite.Run(t, new(CurrenciesratesMainTestSuite))
}

func (suite *CurrenciesratesMainTestSuite) SetupTest() {
    var err error
    suite.log, err = zap.NewProduction()
    assert.NoError(suite.T(), err)
}

func (suite *CurrenciesratesMainTestSuite) TestGetMongoUrl() {
    cfg := &config.Config{}
    assert.Equal(suite.T(), GetMongoUrl(cfg), "")

    cfg = &config.Config{
        MongoHost: mhost,
    }
    assert.Equal(suite.T(), GetMongoUrl(cfg), mhost)

    cfg = &config.Config{
        MongoHost: mhost,
        MongoDatabase: mdb,
    }
    assert.Equal(suite.T(), GetMongoUrl(cfg), mhost)

    cfg = &config.Config{
        MongoHost: mhost,
        MongoDatabase: mdb,
        MongoUser: muser,
    }
    assert.Equal(suite.T(), GetMongoUrl(cfg), muser + "@" + mhost)


    cfg = &config.Config{
        MongoHost: mhost,
        MongoDatabase: mdb,
        MongoUser: muser,
        MongoPassword: mpass,
    }
    assert.Equal(suite.T(), GetMongoUrl(cfg), muser + ":" + mpass + "@" + mhost)
}
