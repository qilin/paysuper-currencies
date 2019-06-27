package service

import (
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	errorCbcaUrlValidationFailed   = "CBCA Rates url validation failed"
	errorCbcaRequestFailed         = "CBCA Rates request failed"
	errorCbcaResponseParsingFailed = "CBCA Rates response parsing failed"
	errorCbcaProcessRatesFailed    = "CBCA Rates save data failed"
	errorCbcaNoResults             = "CBCA Rates no results"
	errorCbcaRateDataNotFound      = "CBCA Rate data not found"
	errorCbcaRateDataInvalidFormat = "CBCA Rate data has invalid format"

	cbcaTo          = "CAD"
	cbcaSource      = "CBCA"
	cbcaUrlTemplate = "https://www.bankofcanada.ca/valet/observations/group/FX_RATES_DAILY/json?start_date=%s"

	cbcaKeyMask = "FX%s%s"
)

type cbcaResponse struct {
	Observations []map[string]interface{} `json:"observations"`
}

// RequestRatesCbca - retriving current rates from Central bank of Canada
func (s *Service) RequestRatesCbca() error {
	zap.S().Info("Requesting rates from CBCA")

	resp, err := s.sendRequestCbca()
	if err != nil {
		return err
	}

	res, err := s.parseResponseCbca(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesCbca(res)
	if err != nil {
		zap.S().Errorw(errorCbcaProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorCbcaProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from CBCA updated")

	return nil
}

func (s *Service) sendRequestCbca() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationJSON,
		headerAccept:      mimeApplicationJSON,
	}

	today := time.Now()
	d := today.AddDate(0, 0, -7)

	reqUrl, err := s.validateUrl(fmt.Sprintf(cbcaUrlTemplate, d.Format(dateFormatLayout)))
	if err != nil {
		zap.S().Errorw(errorCbcaUrlValidationFailed, "error", err)
		s.sendCentrifugoMessage(errorCbcaUrlValidationFailed, err)
		return nil, err
	}

	resp, err := s.request(http.MethodGet, reqUrl.String(), nil, headers)

	if err != nil {
		zap.S().Errorw(errorCbcaRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorCbcaRequestFailed, err)
		return nil, err
	}
	return resp, nil
}

func (s *Service) parseResponseCbca(resp *http.Response) (*cbcaResponse, error) {
	res := &cbcaResponse{}
	err := s.decodeJson(resp, res)

	if err != nil {
		zap.S().Errorw(errorCbcaResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorCbcaResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesCbca(res *cbcaResponse) ([]interface{}, error) {

	if len(res.Observations) == 0 {
		return nil, errors.New(errorCbcaNoResults)
	}

	var rates []interface{}

	lastRates := res.Observations[len(res.Observations)-1]

	for _, cFrom := range s.cfg.SettlementCurrencies {

		if cFrom == cbcaTo {
			continue
		}

		key := fmt.Sprintf(cbcaKeyMask, cFrom, cbcaTo)

		// CBCA not supported currency rate from DKK to CAD and PLN to CAD!
		rateItem, ok := lastRates[key]
		if !ok {
			zap.S().Warnw(errorCbcaRateDataNotFound, "from", cFrom, "to", cbcaTo, "key", key)
			continue
		}

		rawRate, ok := rateItem.(map[string]interface{})["v"]
		if !ok {
			return nil, errors.New(errorCbcaRateDataInvalidFormat)
		}

		rate := rawRate.(float64)

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   cFrom + cbcaTo,
			Rate:   s.toPrecise(rate),
			Source: cbcaSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbcaTo + cFrom,
			Rate:   s.toPrecise(1 / rate),
			Source: cbcaSource,
			Volume: 1,
		})
	}

	return rates, nil
}
