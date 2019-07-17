package service

import (
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"time"
)

const (
	errorCardpaySaveRatesFailed = "Cardpay Rates save data failed"

	cardpaySource     = "CARDPAY"
	cardpayStubSource = "CPSTUB"
)

// SetRatesCardpay - saving rates for Cardpay, getted from messages queue
func (s *Service) SetRatesCardpay(msg *currencies.CardpayRate, dlv amqp.Delivery) error {

	zap.S().Warn("Remove Cardpay settlements stub!")

	id := msg.From + msg.To + msg.Source

	zap.S().Info("Saving rates from Cardpay: ", id)

	var rates []interface{}

	// direct pair
	rates = append(rates, &currencies.RateData{
		Pair:      msg.From + msg.To,
		Rate:      s.toPrecise(msg.Rate),
		Source:    msg.Source,
		Volume:    msg.Volume,
		CreatedAt: msg.CreatedAt,
	})

	err := s.saveRates(collectionRatesNameSuffixCardpay, rates)
	if err != nil {
		zap.S().Errorw(errorCardpaySaveRatesFailed, "error", err, "id", id)
		return s.retry(msg, dlv, id)
	}

	zap.S().Info("Rates from Cardpay saved: ", id)
	return nil

}

// PullRecalcTrigger - switching-On trigger to recalc Paysuper Corrections after get message form special queue
func (s *Service) PullRecalcTrigger(msg *currencies.EmptyRequest, dlv amqp.Delivery) error {
	now := time.Now()
	eod := s.eod(now)
	delta := eod.Sub(now)
	return s.planDelayedTask(int64(delta.Seconds()), triggerCardpay, s.CalculatePaysuperCorrections)
}
