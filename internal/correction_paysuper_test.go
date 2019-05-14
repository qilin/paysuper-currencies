package internal

import (
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "github.com/stretchr/testify/assert"
    "time"
)

var (
    fakeratesOxr = []float64{72.8398, 72.7018, 72.6515, 72.3602, 72.2440, 71.9232, 71.9232, 71.9232,
        71.7210, 71.7453, 71.7150, 72.1111, 72.0882, 72.0882, 72.0882, 72.2024, 72.3096, 72.3096, 72.3096,
        72.3096, 72.3096, 72.3096, 73.1099, 73.0817, 73.0888, 73.0888, 73.0888, 73.5231, 73.3712, 73.3712}

    fakeratesCardpay = []float64{72.7011, 72.7183, 72.1887, 71.9719, 71.9719, 71.9719, 71.9719, 71.9719,
        71.6763, 71.9850, 72.2100, 72.1131, 72.1131, 72.1131, 71.9802, 72.2099, 72.2099, 73.2535, 72.8703,
        72.8703, 72.8703, 72.9750, 72.7907, 73.0256, 73.0572, 73.3493, 73.3493, 73.3493, 73.5854, 73.5954}

    cFrom         = "EUR"
    cTo           = "RUB"
    days          = 7
    timePeriod    = 21
    corridorWidth = float64(0.5)
)

func (suite *CurrenciesratesServiceTestSuite) fillFakes(fakerates []float64, cFrom string, cTo string, collectionSuffux string) error {
    today := time.Now()
    fakes := []interface{}{}
    for day, rate := range fakerates {
        date := today.AddDate(0, 0, -1*day)
        created, _ := ptypes.TimestampProto(suite.service.Bod(date))

        rd1 := &currencyrates.RateData{
            Id:        bson.NewObjectId().Hex(),
            CreatedAt: created,
            Rate:      rate,
            Pair:      cFrom + cTo,
            Source:    "TEST",
        }

        fakes = append(fakes, rd1)

        rd2 := &currencyrates.RateData{
            Id:        bson.NewObjectId().Hex(),
            CreatedAt: created,
            Rate:      1 / rate,
            Pair:      cTo + cFrom,
            Source:    "TEST",
        }

        fakes = append(fakes, rd2)
    }
    cName := suite.service.getCollectionName(collectionSuffux)
    return suite.service.db.Collection(cName).Insert(fakes...)
}

func (suite *CurrenciesratesServiceTestSuite) TestCorrectionPaysuper_getCorrectionValueOk() {
    err := suite.fillFakes(fakeratesOxr, cFrom, cTo, collectionSuffixOxr)
    assert.NoError(suite.T(), err)

    err = suite.fillFakes(fakeratesCardpay, cFrom, cTo, collectionSuffixCardpay)
    assert.NoError(suite.T(), err)

    value1, err1 := suite.service.getCorrectionValue(cFrom+cTo, days, timePeriod, corridorWidth)
    assert.NoError(suite.T(), err1)
    assert.Equal(suite.T(), value1, float64(0.3425))

    value2, err2 := suite.service.getCorrectionValue(cTo+cFrom, days, timePeriod, corridorWidth)
    assert.NoError(suite.T(), err2)
    assert.Equal(suite.T(), value2, float64(0))
}
