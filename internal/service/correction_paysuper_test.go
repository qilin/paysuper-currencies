package service

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/stretchr/testify/assert"
	"time"
)

var (
	// params
	cFrom         = "EUR"
	cTo           = "RUB"
	days          = 7
	timePeriod    = 21
	corridorWidth = float64(0.5)
	today         = time.Now()

	// rates per day, from old to new, last value is today rate
	fakeratesOxr = []float64{72.8398, 72.7018, 72.6515, 72.3602, 72.2440, 71.9232, 71.9232, 71.9232,
		71.7210, 71.7453, 71.7150, 72.1111, 72.0882, 72.0882, 72.0882, 72.2024, 72.3096, 72.3096, 72.3096,
		72.3096, 72.3096, 72.3096, 73.1099, 73.0817, 73.0888, 73.0888, 73.0888, 73.5231, 73.3712, 73.3712}

	// rates per day, from old to new, last value is today rate
	fakeratesCardpay = []float64{72.7011, 72.7183, 72.1887, 71.9719, 71.9719, 71.9719, 71.9719, 71.9719,
		71.6763, 71.9850, 72.2100, 72.1131, 72.1131, 72.1131, 71.9802, 72.2099, 72.2099, 73.2535, 72.8703,
		72.8703, 72.8703, 72.9750, 72.7907, 73.0256, 73.0572, 73.3493, 73.3493, 73.3493, 73.5854, 73.5954}

	// control values

	oxrBLCtrl = []float64{71.48847469213247, 71.4300691618775, 71.39158562004974, 71.39714097104978,
		71.36360828899619, 71.3791601420727, 71.44857525814545}

	oxrBMCtrl = []float64{72.19916190476191, 72.23385714285715, 72.27408571428572, 72.32959047619049,
		72.4057761904762, 72.47472857142859, 72.55330952380955}

	oxrBUCtrl = []float64{72.90984911739135, 73.0376451238368, 73.1565858085217, 73.2620399813312,
		73.44794409195622, 73.57029700078446, 73.65804378947364}

	corrEurRubCtrl = float64(0.337974177)
	corrRubEurCtrl = float64(0.0000107337)
)

func (suite *CurrenciesratesServiceTestSuite) reverse(numbers []float64) []float64 {
	newNumbers := make([]float64, len(numbers))
	for i, j := 0, len(numbers)-1; i < j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}

func (suite *CurrenciesratesServiceTestSuite) fillFakes(fakerates []float64, cFrom string, cTo string, collectionSuffux string) error {
	var fakes []interface{}
	startDate := today.AddDate(0, 0, -1*len(fakerates))
	for day, rate := range fakerates {
		date := startDate.AddDate(0, 0, day)
		created, _ := ptypes.TimestampProto(suite.service.Bod(date))

		rd1 := &currencies.RateData{
			Id:        bson.NewObjectId().Hex(),
			CreatedAt: created,
			Rate:      suite.service.toPrecise(rate),
			Pair:      cFrom + cTo,
			Source:    "TEST",
			Volume:    1,
		}

		fakes = append(fakes, rd1)

		rd2 := &currencies.RateData{
			Id:        bson.NewObjectId().Hex(),
			CreatedAt: created,
			Rate:      suite.service.toPrecise(1 / rate),
			Pair:      cTo + cFrom,
			Source:    "TEST",
			Volume:    1,
		}

		fakes = append(fakes, rd2)
	}

	cName, err := suite.service.getCollectionName(collectionSuffux)
	if err != nil {
		return err
	}
	return suite.service.db.Collection(cName).Insert(fakes...)
}

func (suite *CurrenciesratesServiceTestSuite) TestCorrectionPaysuper_getRatesForBollingerOk() {

	err := suite.fillFakes(fakeratesOxr, cFrom, cTo, collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	totalDays := days + timePeriod - 1
	oxrRates, err := suite.service.getRatesForBollinger(collectionRatesNameSuffixOxr, cFrom+cTo, totalDays)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), oxrRates, fakeratesOxr[len(fakeratesOxr)-totalDays:])
}

func (suite *CurrenciesratesServiceTestSuite) TestCorrectionPaysuper_getBollingerBandsOk() {

	err := suite.fillFakes(fakeratesOxr, cFrom, cTo, collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	oxrL, oxrM, oxrU, err := suite.service.getBollingerBands(collectionRatesNameSuffixOxr, cFrom+cTo, days, timePeriod)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), len(oxrL), days)
	assert.Equal(suite.T(), len(oxrM), days)
	assert.Equal(suite.T(), len(oxrU), days)

	assert.Equal(suite.T(), oxrL, oxrBLCtrl)
	assert.Equal(suite.T(), oxrM, oxrBMCtrl)
	assert.Equal(suite.T(), oxrU, oxrBUCtrl)
}

func (suite *CurrenciesratesServiceTestSuite) TestCorrectionPaysuper_getCorrectionValueOk() {

	err := suite.fillFakes(fakeratesOxr, cFrom, cTo, collectionRatesNameSuffixOxr)
	assert.NoError(suite.T(), err)

	err = suite.fillFakes(fakeratesCardpay, cFrom, cTo, collectionRatesNameSuffixCardpay)
	assert.NoError(suite.T(), err)

	value1, err1 := suite.service.getCorrectionValue(cFrom+cTo, days, timePeriod, corridorWidth)
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), value1, corrEurRubCtrl)

	value2, err2 := suite.service.getCorrectionValue(cTo+cFrom, days, timePeriod, corridorWidth)
	assert.NoError(suite.T(), err2)
	assert.Equal(suite.T(), value2, corrRubEurCtrl)
}
