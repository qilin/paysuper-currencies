package currency

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

type CurrenciesTestSuite struct {
	suite.Suite
	log *zap.Logger
}

func Test_CurrenciesratesService(t *testing.T) {
	suite.Run(t, new(CurrenciesTestSuite))
}

func (suite *CurrenciesTestSuite) SetupTest() {
	var err error

	suite.log, err = zap.NewProduction()
	assert.NoError(suite.T(), err)
}

func (suite *CurrenciesTestSuite) Test_Currencies() {
	assert.Equal(suite.T(), len(CurrencyDefinitions), 55)
}
