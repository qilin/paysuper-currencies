package internal

import (
    "github.com/globalsign/mgo/bson"
    currencyrates "github.com/paysuper/paysuper-currencies-rates/proto"
    "go.uber.org/zap"
)

const (
    paysuperSource = "PS"

    errorPaysuperRateCalc = "paysuper prediction rate calculation error"
    errorPaysuperRateSave = "paysuper prediction rates save error"
)

func (s *Service) SetRatesPaysuper() error {

    zap.S().Info("Start calculation of prediction rates for Paysuper")

    var (
        cFrom string
        cTo   string
        rates = []*currencyrates.RateData{}
    )

    for _, cFrom = range s.cfg.OxrBaseCurrencies {
        for _, cTo = range s.cfg.OxrSupportedCurrencies {

            if cFrom == cTo {
                continue
            }

            rd, err := s.getRatePaysuper(cFrom, cTo)
            if err != nil {
                zap.S().Errorw(errorPaysuperRateCalc, "error", err)
                s.sendCentrifugoMessage(errorPaysuperRateCalc, err)
                return err
            }
            rates = append(rates, rd)

            rd, err = s.getRatePaysuper(cTo, cFrom)
            if err != nil {
                zap.S().Errorw(errorPaysuperRateCalc, "error", err)
                s.sendCentrifugoMessage(errorPaysuperRateCalc, err)
                return err
            }
            rates = append(rates, rd)
        }
    }

    err := s.saveRates(collectionSuffixPaysuper, rates)
    if err != nil {
        zap.S().Errorw(errorPaysuperRateSave, "error", err)
        s.sendCentrifugoMessage(errorPaysuperRateSave, err)
        return err
    }

    zap.S().Info("Prediction rates for Paysuper updated")

    return nil
}

func (s *Service) getRatePaysuper(cFrom string, cTo string) (*currencyrates.RateData, error) {
    res := &currencyrates.RateData{}

    err := s.getRate(collectionSuffixOxr, cFrom, cTo, bson.M{}, res)
    if err != nil {
        return nil, err
    }

    correction, err := s.GetPaysuperCorrection(cFrom + cTo)
    if err != nil {
        return nil, err
    }

    rd := &currencyrates.RateData{
        Pair:   cFrom + cTo,
        Rate:   s.toPrecise(res.Rate + correction),
        Source: paysuperSource,
    }

    return rd, nil

}
