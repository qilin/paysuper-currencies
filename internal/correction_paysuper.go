package internal

import (
    "errors"
    "github.com/globalsign/mgo/bson"
    "go.uber.org/zap"
    "time"
)

const (
    corridorMin = float64(0)
    corridorMax = float64(1)

    errorInvalidCorrectionCoridor = "invalid correction corridor value"
    errorCalculateCorrection      = "can't calculate correction"
    errorGetCorrection            = "can't get correction value"
)

type PaysuperCorrection struct {
    Pair      string    `bson:"pair"`
    CreatedAt time.Time `bson:"created_at"`
    Value     float64   `bson:"value"`
}

type PaysuperCorridor struct {
    CreatedAt time.Time `bson:"created_at"`
    Value     float64   `bson:"value"`
}

func (s *Service) GetPaysuperCorrection(pair string) (float64, error) {
    if !s.isPairExists(pair) {
        zap.S().Errorw(errorGetCorrection, "error", errorCurrencyPairNotExists, "pair", pair)
        return 0, errors.New(errorCurrencyPairNotExists)
    }

    query := bson.M{"pair": pair}

    res := &PaysuperCorrection{}
    err := s.db.Collection(collectionNamePaysuperCorrections).Find(query).Sort("-_id").Limit(1).One(res)
    if err != nil {
        zap.S().Errorw(errorGetCorrection, "error", err, "pair", pair)
        return 0, err
    }

    return res.Value, nil
}

func (s *Service) CalculatePaysuperCorrections() error {
    days := s.cfg.BollingerDays
    timePeriod := s.cfg.BollingerPeriod
    corridorWidth, err := s.getPaysuperCorrectionCorridorWidth()
    if err != nil {
        zap.S().Errorw(errorCalculateCorrection, "error", err)
        return err
    }

    now := time.Now()

    corrections := []interface{}{}

    for _, cFrom := range s.cfg.OxrBaseCurrencies {

        for _, cTo := range s.cfg.OxrSupportedCurrencies {

            if cFrom == cTo {
                continue
            }

            pair := cFrom + cTo

            value, err := s.getCorrectionValue(pair, days, timePeriod, corridorWidth)
            if err != nil {
                zap.S().Errorw(errorCalculateCorrection, "error", err, "pair", pair)
                return err
            }
            corrections = append(corrections, &PaysuperCorrection{
                Pair:      pair,
                Value:     value,
                CreatedAt: now,
            })

            reversePair := cTo + cFrom
            reverseValue, err := s.getCorrectionValue(reversePair, days, timePeriod, corridorWidth)
            if err != nil {
                zap.S().Errorw(errorCalculateCorrection, "error", err, "pair", reversePair)
                return err
            }
            corrections = append(corrections, &PaysuperCorrection{
                Pair:      reversePair,
                Value:     reverseValue,
                CreatedAt: now,
            })
        }
    }

    err = s.db.Collection(collectionNamePaysuperCorrections).Insert(corrections...)

    if err != nil {
        zap.S().Errorw(errorCalculateCorrection, "error", err)
        return err
    }

    return nil
}

func (s *Service) getCorrectionValue(pair string, days int, timePeriod int, corridorWidth float64) (float64, error) {

    oxrL, oxrM, oxrU, err := s.getBollingerBands(collectionSuffixOxr, pair, days, timePeriod)
    if err != nil {
        return 0, err
    }

    cpL, cpM, cpU, err := s.getBollingerBands(collectionSuffixCardpay, pair, days, timePeriod)
    if err != nil {
        return 0, err
    }

    max := float64(0)

    for i := 0; i < days; i++ {

        corridorOxr := oxrU[i] - oxrL[i]
        corridorCp := cpU[i] - cpL[i]
        deltaMed := cpM[i] - oxrM[i]
        deltaCorridor := corridorCp - corridorOxr
        corridorCorrection := deltaCorridor * corridorWidth
        correction := deltaMed + corridorCorrection

        if correction > max {
            max = correction
        }
    }

    return s.toPrecise(max), nil
}

func (s *Service) getBollingerBands(collectionSuffix string, pair string, days int, timePeriod int) ([]float64, []float64, []float64, error) {
    today := time.Now()
    startDate := today.AddDate(0, 0, -1*(days+timePeriod))
    oxrRates, err := s.GetRatesForBollinger(collectionSuffix, pair, startDate)
    if err != nil {
        return nil, nil, nil, err
    }
    oxrL, oxrM, oxrU := s.Bollinger(oxrRates, timePeriod)
    return oxrL, oxrM, oxrU, nil
}

func (s *Service) getPaysuperCorrectionCorridorWidth() (float64, error) {

    res := &PaysuperCorridor{}
    err := s.db.Collection(collectionNamePaysuperCorridors).Find(nil).Sort("-_id").Limit(1).One(res)
    if err != nil {
        return 0, err
    }

    value := res.Value
    if value < corridorMin || value > corridorMax {
        return 0, errors.New(errorInvalidCorrectionCoridor)
    }

    return value, nil
}
