package service

import (
	"github.com/globalsign/mgo/bson"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"go.uber.org/zap"
)

const (
	stockSource = "STOCK"

	errorStockRateCalc = "stock rate calculation error"
	errorStockRateSave = "stock rates save error"
)

// SetRatesStock - set rates for stock exchange
func (s *Service) SetRatesStock() error {

	zap.S().Info("Start calculation rates for Stock")

	var (
		cFrom string
		cTo   string
		rates []interface{}
	)

	for _, cFrom = range s.cfg.SettlementCurrencies {
		for _, cTo = range s.cfg.RatesRequestCurrencies {

			if cFrom == cTo {
				continue
			}

			rd, err := s.getRateStock(cFrom, cTo)
			if err != nil {
				zap.S().Errorw(errorStockRateCalc, "error", err)
				s.sendCentrifugoMessage(errorStockRateCalc, err)
				return err
			}
			rates = append(rates, rd)

			rd, err = s.getRateStock(cTo, cFrom)
			if err != nil {
				zap.S().Errorw(errorStockRateCalc, "error", err)
				s.sendCentrifugoMessage(errorStockRateCalc, err)
				return err
			}
			rates = append(rates, rd)
		}
	}

	err := s.saveRates(collectionRatesNameSuffixStock, rates)
	if err != nil {
		zap.S().Errorw(errorStockRateSave, "error", err)
		s.sendCentrifugoMessage(errorStockRateSave, err)
		return err
	}

	zap.S().Info("Rates for Stock updated")

	return nil
}

func (s *Service) getRateStock(cFrom string, cTo string) (*currencies.RateData, error) {
	res := &currencies.RateData{}

	err := s.getRate(collectionRatesNameSuffixOxr, cFrom, cTo, bson.M{}, "", res)
	if err != nil {
		return nil, err
	}

	rd := &currencies.RateData{
		Pair:   res.Pair,
		Rate:   res.Rate,
		Source: stockSource,
		Volume: 1,
	}

	return rd, nil
}
