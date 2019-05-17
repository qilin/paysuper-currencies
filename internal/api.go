package internal

import (
    "context"
    "errors"
    "github.com/globalsign/mgo/bson"
    "github.com/golang/protobuf/ptypes"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
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

func (s *Service) GetRateCurrentCommon(
    ctx context.Context,
    req *currencyrates.GetRateCurrentCommonRequest,
    res *currencyrates.RateData,
) error {
    err := s.getRate(req.RateType, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorGetRateCurrentCommonRequest, "error", err, "req", req)
        return err
    }
    s.applyCorrection(res, req.RateType, "")
    return nil
}

func (s *Service) GetRateByDateCommon(
    ctx context.Context,
    req *currencyrates.GetRateByDateCommonRequest,
    res *currencyrates.RateData,
) error {
    dt, err := ptypes.Timestamp(req.Datetime)
    if err != nil {
        zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
        return err
    }

    err = s.getRateByDate(req.RateType, req.From, req.To, dt, res)
    if err != nil {
        zap.S().Errorw(errorGetRateByDateCommonRequest, "error", err, "req", req)
        return err
    }

    s.applyCorrection(res, req.RateType, "")
    return nil
}

func (s *Service) GetRateCurrentForMerchant(
    ctx context.Context,
    req *currencyrates.GetRateCurrentForMerchantRequest,
    res *currencyrates.RateData,
) error {
    if req.MerchantId == "" {
        zap.S().Errorw(errorMerchantIdRequired, "req", req)
        return errors.New(errorMerchantIdRequired)
    }

    err := s.getRate(req.RateType, req.From, req.To, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorGetRateCurrentForMerchantRequest, "error", err, "req", req)
        return err
    }
    s.applyCorrection(res, req.RateType, req.MerchantId)
    return nil
}

func (s *Service) GetRateByDateForMerchant(
    ctx context.Context,
    req *currencyrates.GetRateByDateForMerchantRequest,
    res *currencyrates.RateData,
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

    err = s.getRateByDate(req.RateType, req.From, req.To, dt, res)
    if err != nil {
        zap.S().Errorw(errorGetRateByDateForMerchantRequest, "error", err, "req", req)
        return err
    }

    s.applyCorrection(res, req.RateType, req.MerchantId)
    return nil
}

func (s *Service) ExchangeCurrencyCurrentCommon(
    ctx context.Context,
    req *currencyrates.ExchangeCurrencyCurrentCommonRequest,
    res *currencyrates.ExchangeCurrencyResponse,
) error {
    err := s.exchangeCurrency(req.RateType, req.From, req.To, req.Amount, "", bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorExchangeCurrencyCurrentCommon, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) ExchangeCurrencyCurrentForMerchant(
    ctx context.Context,
    req *currencyrates.ExchangeCurrencyCurrentForMerchantRequest,
    res *currencyrates.ExchangeCurrencyResponse,
) error {
    if req.MerchantId == "" {
        zap.S().Errorw(errorMerchantIdRequired, "req", req)
        return errors.New(errorMerchantIdRequired)
    }
    err := s.exchangeCurrency(req.RateType, req.From, req.To, req.Amount, req.MerchantId, bson.M{}, res)
    if err != nil {
        zap.S().Errorw(errorExchangeCurrencyCurrentForMerchant, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) ExchangeCurrencyByDateCommon(
    ctx context.Context,
    req *currencyrates.ExchangeCurrencyByDateCommonRequest,
    res *currencyrates.ExchangeCurrencyResponse,
) error {
    dt, err := ptypes.Timestamp(req.Datetime)
    if err != nil {
        zap.S().Errorw(errorDatetimeConversion, "error", err, "req", req)
        return err
    }

    err = s.exchangeCurrencyByDate(req.RateType, req.From, req.To, req.Amount, "", dt, res)
    if err != nil {
        zap.S().Errorw(errorExchangeCurrencyByDateCommon, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) ExchangeCurrencyByDateForMerchant(
    ctx context.Context,
    req *currencyrates.ExchangeCurrencyByDateForMerchantRequest,
    res *currencyrates.ExchangeCurrencyResponse,
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

    err = s.exchangeCurrencyByDate(req.RateType, req.From, req.To, req.Amount, req.MerchantId, dt, res)
    if err != nil {
        zap.S().Errorw(errorExchangeCurrencyByDateForMerchant, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) GetCommonRateCorrectionRule(
    ctx context.Context,
    req *currencyrates.CommonCorrectionRuleRequest,
    res *currencyrates.CorrectionRule,
) error {
    err := s.getCorrectionRule(req.RateType, "", res)
    if err != nil {
        zap.S().Errorw(errorCorrectionRuleNotFound, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) GetMerchantRateCorrectionRule(
    ctx context.Context,
    req *currencyrates.MerchantCorrectionRuleRequest,
    res *currencyrates.CorrectionRule,
) error {
    if req.MerchantId == "" {
        zap.S().Errorw(errorMerchantIdRequired, "req", req)
        return errors.New(errorMerchantIdRequired)
    }

    err := s.getCorrectionRule(req.RateType, req.MerchantId, res)
    if err != nil {
        zap.S().Errorw(errorCorrectionRuleNotFound, "error", err, "req", req)
        return err
    }
    return nil
}

func (s *Service) AddCommonRateCorrectionRule(
    ctx context.Context,
    req *currencyrates.CommonCorrectionRule,
    res *currencyrates.EmptyResponse,
) error {
    return s.addCorrectionRule(req.RateType, req.CommonCorrection, req.PairCorrection, "")
}

func (s *Service) AddMerchantRateCorrectionRule(
    ctx context.Context,
    req *currencyrates.CorrectionRule,
    res *currencyrates.EmptyResponse,
) error {
    if req.MerchantId == "" {
        zap.S().Errorw(errorMerchantIdRequired, "req", req)
        return errors.New(errorMerchantIdRequired)
    }

    return s.addCorrectionRule(req.RateType, req.CommonCorrection, req.PairCorrection, req.MerchantId)
}

func (s *Service) SetPaysuperCorrectionCorridor(
    ctx context.Context,
    req *currencyrates.CorrectionCorridor,
    res *currencyrates.EmptyResponse,
) error {

    corridor := PaysuperCorridor{
        Value:     req.Value,
        CreatedAt: time.Now(),
    }

    err := s.db.Collection(collectionNamePaysuperCorridors).Insert(corridor)
    if err != nil {
        zap.S().Errorw(errorDbInsertFailed, "error", err, "data", corridor)
        return err
    }

    return nil
}
