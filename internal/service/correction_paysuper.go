package service

import (
    "errors"
    "github.com/globalsign/mgo/bson"
    "github.com/paysuper/paysuper-currencies/pkg"
    "github.com/thetruetrade/gotrade"
    "github.com/thetruetrade/gotrade/indicators"
    "go.uber.org/zap"
    "time"
)

const (
    corridorMin = float64(0)
    corridorMax = float64(1)

    errorNotEnoughRatesDataForBollinger = "not enough rates data for Bollinger"
    errorInvalidBollingerBandsLength    = "invalid bollinger bands length"
    errorInvalidCorrectionCoridor       = "invalid correction corridor value"
    errorInvalidBollingerDays           = "invalid Bollinger days"
    errorInvalidBollingerPeriod         = "invalid Bollinger period"
    errorCalculateCorrection            = "can't calculate correction"
    errorGetCorrection                  = "can't get correction value"
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
    if days < 1 {
        zap.S().Errorw(errorInvalidBollingerDays, "days", days)
        return errors.New(errorInvalidBollingerDays)
    }

    timePeriod := s.cfg.BollingerPeriod
    if timePeriod < 2 {
        zap.S().Errorw(errorInvalidBollingerPeriod, "timePeriod", timePeriod)
        return errors.New(errorInvalidBollingerPeriod)
    }

    corridorWidth, err := s.getPaysuperCorrectionCorridorWidth()
    if err != nil {
        zap.S().Errorw(errorCalculateCorrection, "error", err)
        return err
    }

    now := time.Now()

    corrections := []interface{}{}

    for _, cFrom := range s.cfg.SettlementCurrencies {

        for _, cTo := range s.cfg.RatesRequestCurrencies {

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

    err = s.releaseTrigger(triggerCardpay)
    if err != nil {
        zap.S().Errorw(errorReleaseTrigger, "error", err, "trigger", triggerCardpay)
        return err
    }

    return nil
}

func (s *Service) getCorrectionValue(pair string, days int, timePeriod int, corridorWidth float64) (float64, error) {

    oxrL, oxrM, oxrU, err := s.getBollingerBands(collectionRatesNameSuffixOxr, pair, days, timePeriod)
    if err != nil {
        return 0, err
    }

    cpL, cpM, cpU, err := s.getBollingerBands(collectionRatesNameSuffixCardpay, pair, days, timePeriod)
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

func (s *Service) getBollingerBands(collectionRatesNameSuffix string, pair string, days int, timePeriod int) ([]float64, []float64, []float64, error) {
    totalDays := days + timePeriod - 1
    rates, err := s.getRatesForBollinger(collectionRatesNameSuffix, pair, totalDays)
    if err != nil {
        return nil, nil, nil, err
    }
    l, m, u := s.bollinger(rates, timePeriod)
    if len(l) != days || len(m) != days || len(u) != days {
        return nil, nil, nil, errors.New(errorInvalidBollingerBandsLength)
    }
    return l, m, u, nil
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

func (s *Service) getRatesForBollinger(collectionRatesNameSuffix string, pair string, limit int) (res []float64, err error) {
    if !s.isPairExists(pair) {
        return nil, errors.New(errorCurrencyPairNotExists)
    }

    isCardpay := collectionRatesNameSuffix == pkg.RateTypeCardpay

    cName, err := s.getCollectionName(collectionRatesNameSuffix)
    if err != nil {
        return nil, err
    }

    groupQuery := bson.M{
        "_id": bson.M{"create_date": "$create_date"},
    }

    if isCardpay {
        groupQuery["numerator"] = bson.M{"$sum": bson.M{"$multiply": []string{"$rate", "$volume"}}}
        groupQuery["denominator"] = bson.M{"$sum": "$volume"}
    } else {
        groupQuery["value"] = bson.M{"$last": "$rate"}
    }

    q := []bson.M{
        {"$match": bson.M{"pair": pair}},
        {"$group": groupQuery},
        {"$sort": bson.M{"_id": -1}},
        {"$limit": limit},
        {"$sort": bson.M{"_id": 1}},
    }

    if isCardpay {
        q = append(q, bson.M{"$project": bson.M{
            "value": bson.M{"$divide": []string{"$numerator", "$denominator"}},
        }})
    }

    var resp []map[string]interface{}
    err = s.db.Collection(cName).Pipe(q).All(&resp)

    if err != nil {
        return nil, err
    }

    if len(resp) < limit {
        return nil, errors.New(errorNotEnoughRatesDataForBollinger)
    }

    for _, val := range resp[len(resp)-limit:] {
        res = append(res, s.toPrecise(val["value"].(float64)))
    }

    return res, nil
}

func (s *Service) bollinger(rates []float64, timePeriod int) ([]float64, []float64, []float64) {

    priceStream := gotrade.NewDailyDOHLCVStream()
    bb, _ := indicators.NewBollingerBandsForStream(priceStream, timePeriod, gotrade.UseClosePrice)

    for _, val := range rates {
        dohlcv := gotrade.NewDOHLCVDataItem(time.Now(), 0, 0, 0, val, 0)
        priceStream.ReceiveTick(dohlcv)
    }

    return bb.LowerBand, bb.MiddleBand, bb.UpperBand
}
