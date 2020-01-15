package service

import (
	"context"
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-currencies/pkg"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"go.uber.org/zap"
	"time"
)

const (
	errorMerchantIdRequired                 = "merchant id required"
	errorTimestampRequired                  = "timestamp: nil Timestamp"
	errorGetRateCurrentCommonRequest        = "get current rate common failed"
	errorGetRateByDateCommonRequest         = "get rate by date failed"
	errorGetRateCurrentForMerchantRequest   = "get current rate for merchant failed"
	errorGetRateByDateForMerchantRequest    = "get rate by date for merchant failed"
	errorExchangeCurrencyCurrentCommon      = "common exchange currency by current rate failed"
	errorExchangeCurrencyCurrentForMerchant = "exchange currency by current rate for merchant failed"
	errorExchangeCurrencyByDateCommon       = "common exchange currency by date failed"
	errorExchangeCurrencyByDateForMerchant  = "exchange currency by date for merchant failed"
)

// GetRateCurrentCommon - get current rate with common correction rule applied
func (s *Service) GetRateCurrentCommon(
	ctx context.Context,
	req *currencies.GetRateCurrentCommonRequest,
	res *currencies.RateData,
) error {
	query := bson.M{}
	if req.RateType == pkg.RateTypeCardpay {
		query = s.getByDateQuery(time.Now())
	}
	err := s.getRate(req.RateType, req.From, req.To, query, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorGetRateCurrentCommonRequest, "error", err, "req", req)
		return err
	}
	s.applyCorrection(res, req.RateType, req.ExchangeDirection, "")
	return nil
}

// GetRateByDateCommon - get rate by date with common correction rule applied
func (s *Service) GetRateByDateCommon(
	ctx context.Context,
	req *currencies.GetRateByDateCommonRequest,
	res *currencies.RateData,
) error {
	dt, err := ptypes.Timestamp(req.Datetime)
	if err != nil {
		zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
		return err
	}

	err = s.getRateByDate(req.RateType, req.From, req.To, dt, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorGetRateByDateCommonRequest, "error", err, "req", req)
		return err
	}

	s.applyCorrection(res, req.RateType, req.ExchangeDirection, "")
	return nil
}

// GetRateCurrentForMerchant - get current rate with merchant's correction rule applied
func (s *Service) GetRateCurrentForMerchant(
	ctx context.Context,
	req *currencies.GetRateCurrentForMerchantRequest,
	res *currencies.RateData,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}

	query := bson.M{}
	if req.RateType == pkg.RateTypeCardpay {
		query = s.getByDateQuery(time.Now())
	}

	err := s.getRate(req.RateType, req.From, req.To, query, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorGetRateCurrentForMerchantRequest, "error", err, "req", req)
		return err
	}
	s.applyCorrection(res, req.RateType, req.ExchangeDirection, req.MerchantId)
	return nil
}

// GetRateByDateForMerchant - get rate by date with merchant's correction rule applied
func (s *Service) GetRateByDateForMerchant(
	ctx context.Context,
	req *currencies.GetRateByDateForMerchantRequest,
	res *currencies.RateData,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}

	dt, err := ptypes.Timestamp(req.Datetime)
	if err != nil {
		zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
		return err
	}

	err = s.getRateByDate(req.RateType, req.From, req.To, dt, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorGetRateByDateForMerchantRequest, "error", err, "req", req)
		return err
	}

	s.applyCorrection(res, req.RateType, req.ExchangeDirection, req.MerchantId)
	return nil
}

// ExchangeCurrencyCurrentCommon - exchange currency via current rate with common correction rule applied
func (s *Service) ExchangeCurrencyCurrentCommon(
	ctx context.Context,
	req *currencies.ExchangeCurrencyCurrentCommonRequest,
	res *currencies.ExchangeCurrencyResponse,
) error {
	query := bson.M{}
	if req.RateType == pkg.RateTypeCardpay {
		query = s.getByDateQuery(time.Now())
	}
	err := s.exchangeCurrency(req.RateType, req.ExchangeDirection, req.From, req.To, req.Amount, "", query, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorExchangeCurrencyCurrentCommon, "error", err, "req", req)
		return err
	}
	return nil
}

// ExchangeCurrencyCurrentForMerchant - exchange currency via current rate with merchant's correction rule applied
func (s *Service) ExchangeCurrencyCurrentForMerchant(
	ctx context.Context,
	req *currencies.ExchangeCurrencyCurrentForMerchantRequest,
	res *currencies.ExchangeCurrencyResponse,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}
	query := bson.M{}
	if req.RateType == pkg.RateTypeCardpay {
		query = s.getByDateQuery(time.Now())
	}
	err := s.exchangeCurrency(req.RateType, req.ExchangeDirection, req.From, req.To, req.Amount, req.MerchantId, query, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorExchangeCurrencyCurrentForMerchant, "error", err, "req", req)
		return err
	}
	return nil
}

// ExchangeCurrencyByDateCommon - exchange currency via rate by date with common correction rule applied
func (s *Service) ExchangeCurrencyByDateCommon(
	ctx context.Context,
	req *currencies.ExchangeCurrencyByDateCommonRequest,
	res *currencies.ExchangeCurrencyResponse,
) error {
	dt, err := ptypes.Timestamp(req.Datetime)
	if err != nil {
		zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
		return err
	}

	err = s.exchangeCurrencyByDate(req.RateType, req.ExchangeDirection, req.From, req.To, req.Amount, "", dt, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorExchangeCurrencyByDateCommon, "error", err, "req", req)
		return err
	}
	return nil
}

// ExchangeCurrencyByDateForMerchant - exchange currency via rate by date with merchant's correction rule applied
func (s *Service) ExchangeCurrencyByDateForMerchant(
	ctx context.Context,
	req *currencies.ExchangeCurrencyByDateForMerchantRequest,
	res *currencies.ExchangeCurrencyResponse,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}

	dt, err := ptypes.Timestamp(req.Datetime)
	if err != nil {
		zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
		return err
	}

	err = s.exchangeCurrencyByDate(req.RateType, req.ExchangeDirection, req.From, req.To, req.Amount, req.MerchantId, dt, req.Source, res)
	if err != nil {
		zap.S().Errorw(errorExchangeCurrencyByDateForMerchant, "error", err, "req", req)
		return err
	}
	return nil
}

// GetCommonRateCorrectionRule - returns common (default) correction rule for passed rate type
func (s *Service) GetCommonRateCorrectionRule(
	ctx context.Context,
	req *currencies.CommonCorrectionRuleRequest,
	res *currencies.CorrectionRule,
) error {
	cr, err := s.getCorrectionRule(req.RateType, req.ExchangeDirection, "")
	if err != nil {
		zap.S().Errorw(errorCorrectionRuleNotFound, "error", err, "req", req)
		return err
	}

	res.MerchantId = cr.MerchantId
	res.RateType = cr.RateType
	res.ExchangeDirection = cr.ExchangeDirection
	res.CommonCorrection = cr.CommonCorrection
	res.PairCorrection = cr.PairCorrection
	res.CreatedAt = cr.CreatedAt

	return nil
}

// GetMerchantRateCorrectionRule - returns merchant's correction rule for passed rate type, with fallback to common (default) if it not exists
func (s *Service) GetMerchantRateCorrectionRule(
	ctx context.Context,
	req *currencies.MerchantCorrectionRuleRequest,
	res *currencies.CorrectionRule,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}

	cr, err := s.getCorrectionRule(req.RateType, req.ExchangeDirection, req.MerchantId)
	if err != nil {
		zap.S().Errorw(errorCorrectionRuleNotFound, "error", err, "req", req)
		return err
	}

	res.MerchantId = cr.MerchantId
	res.RateType = cr.RateType
	res.ExchangeDirection = cr.ExchangeDirection
	res.CommonCorrection = cr.CommonCorrection
	res.PairCorrection = cr.PairCorrection
	res.CreatedAt = cr.CreatedAt

	return nil
}

// AddCommonRateCorrectionRule - adding new default correction rule for passed rate type
func (s *Service) AddCommonRateCorrectionRule(
	ctx context.Context,
	req *currencies.CommonCorrectionRule,
	res *currencies.EmptyResponse,
) error {
	return s.addCorrectionRule(req.RateType, req.ExchangeDirection, req.CommonCorrection, req.PairCorrection, "")
}

// AddMerchantRateCorrectionRule - adding new merchant's correction rule for passed rate type and merchant id
func (s *Service) AddMerchantRateCorrectionRule(
	ctx context.Context,
	req *currencies.CorrectionRule,
	res *currencies.EmptyResponse,
) error {
	if req.MerchantId == "" {
		zap.S().Errorw(errorMerchantIdRequired, "req", req)
		return errors.New(errorMerchantIdRequired)
	}

	return s.addCorrectionRule(req.RateType, req.ExchangeDirection, req.CommonCorrection, req.PairCorrection, req.MerchantId)
}

// GetSupportedCurrencies - returns list of all supported currencies
func (s *Service) GetSupportedCurrencies(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesList,
) error {
	for _, v := range s.cfg.SupportedCurrencies {
		res.Currencies = append(res.Currencies, v)
	}
	return nil
}

// GetSettlementCurrencies - returns list of settlement currencies
func (s *Service) GetSettlementCurrencies(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesList,
) error {
	for _, v := range s.cfg.SettlementCurrencies {
		res.Currencies = append(res.Currencies, v)
	}
	return nil
}

// GetPriceCurrencies - returns list of price currencies
func (s *Service) GetPriceCurrencies(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesList,
) error {
	for _, v := range s.cfg.PriceCurrencies {
		res.Currencies = append(res.Currencies, v)
	}
	return nil
}

// GetVatCurrencies - returns list of vat currencies
func (s *Service) GetVatCurrencies(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesList,
) error {
	for _, v := range s.cfg.VatCurrencies {
		res.Currencies = append(res.Currencies, v)
	}
	return nil
}

// GetAccountingCurrencies - returns list of settlement accounting currencies
func (s *Service) GetAccountingCurrencies(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesList,
) error {
	for _, v := range s.cfg.AccountingCurrencies {
		res.Currencies = append(res.Currencies, v)
	}
	return nil
}

// GetCurrenciesPrecision - returns map of currencies with theirs precision
func (s *Service) GetCurrenciesPrecision(
	ctx context.Context,
	req *currencies.EmptyRequest,
	res *currencies.CurrenciesPrecisionResponse,
) error {
	res.Values = s.cfg.CurrenciesPrecision
	return nil
}
