package internal

import (
    "github.com/globalsign/mgo/bson"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "go.uber.org/zap"
)

const (
    stockSource = "STOCK"

    errorStockRateCalc = "stock rate calculation error"
    errorStockRateSave = "stock rates save error"
)

func (s *Service) SetRatesStock() error {

    zap.S().Info("Start calculation rates for Stock")

    rule := &currencyrates.CorrectionRule{}
    err := s.getCorrectionRule(collectionSuffixStock, "", rule)
    if err != nil {
        zap.S().Errorw(errorCorrectionRuleNotFound, "error", err)
        s.sendCentrifugoMessage(errorCorrectionRuleNotFound, err)
        return err
    }

    var (
        cFrom string
        cTo   string
        rates []interface{}
    )

    for _, cFrom = range s.cfg.OxrBaseCurrencies {
        for _, cTo = range s.cfg.OxrSupportedCurrencies {

            if cFrom == cTo {
                continue
            }

            rd, err := s.getRateStock(cFrom, cTo, rule)
            if err != nil {
                zap.S().Errorw(errorStockRateCalc, "error", err)
                s.sendCentrifugoMessage(errorStockRateCalc, err)
                return err
            }
            rates = append(rates, rd)

            rd, err = s.getRateStock(cTo, cFrom, rule)
            if err != nil {
                zap.S().Errorw(errorStockRateCalc, "error", err)
                s.sendCentrifugoMessage(errorStockRateCalc, err)
                return err
            }
            rates = append(rates, rd)
        }
    }

    err = s.saveRates(collectionSuffixStock, rates)
    if err != nil {
        zap.S().Errorw(errorStockRateSave, "error", err)
        s.sendCentrifugoMessage(errorStockRateSave, err)
        return err
    }

    zap.S().Info("Rates for Stock updated")

    return nil
}

func (s *Service) getRateStock(cFrom string, cTo string, rule *currencyrates.CorrectionRule) (*currencyrates.RateData, error) {
    res := &currencyrates.RateData{}

    err := s.getRate(collectionSuffixOxr, cFrom, cTo, bson.M{}, res)
    if err != nil {
        return nil, err
    }

    s.applyCorrectionRule(res, rule)

    res.Id = bson.NewObjectId().Hex()
    res.Source = stockSource

    rd := &currencyrates.RateData{
        Pair: res.Pair,
        Rate: res.Rate,
        Source: stockSource,
    }

    return rd, nil
}
