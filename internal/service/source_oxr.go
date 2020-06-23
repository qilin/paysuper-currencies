package service

import (
	"errors"
	"fmt"
	currencies "github.com/paysuper/paysuper-proto/go/currenciespb"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

const (
	errorOxrUrlValidationFailed   = "OXR Rates url validation failed"
	errorOxrRequestFailed         = "OXR Rates request failed"
	errorOxrResponseParsingFailed = "OXR Rates response parsing failed"
	errorOxrSaveRatesFailed       = "OXR Rates save data failed"
	errorOxrNoResults             = "OXR Rates no results"
	errorOxrInvalidFrom           = "OXR Rates invalid from"

	oxrSource = "OXR"

	oxrUrlTemplate = "https://openexchangerates.org/api/latest.json?base=%s&%s"
)

type oxrResponse struct {
	Disclaimer string
	License    string
	Timestamp  int64
	Base       string
	Rates      map[string]float64
}

// RequestRatesOxr - retriving current rates from openexchangerates.org
func (s *Service) RequestRatesOxr() error {
	zap.S().Info("Requesting rates from OXR")

	queryParams := url.Values{
		"app_id":  []string{s.cfg.OxrAppId},
		"symbols": []string{strings.Join(s.cfg.RatesRequestCurrencies, ",")},
	}
	queryString := queryParams.Encode()

	for _, from := range s.cfg.SettlementCurrencies {

		resp, err := s.sendRequestOxr(from, queryString)
		if err != nil {
			return err
		}

		res, err := s.parseResponseOxr(resp)
		if err != nil {
			return err
		}

		rates, err := s.processRatesOxr(res)
		if err != nil {
			zap.S().Errorw(errorOxrSaveRatesFailed, "error", err)
			s.sendCentrifugoMessage(errorOxrSaveRatesFailed, err)
			return err
		}

		err = s.saveRates(collectionRatesNameSuffixOxr, rates)
		if err != nil {
			return err
		}
	}

	zap.S().Info("Rates from OXR updated")

	return nil
}
func (s *Service) sendRequestOxr(from string, queryString string) (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationJSON,
		headerAccept:      mimeApplicationJSON,
	}

	reqUrl, err := s.validateUrl(fmt.Sprintf(oxrUrlTemplate, from, queryString))

	if err != nil {
		zap.S().Errorw(errorOxrUrlValidationFailed, "error", err)
		s.sendCentrifugoMessage(errorOxrUrlValidationFailed, err)
		return nil, err
	}

	zap.S().Info("Sending request to url: ", reqUrl.String())

	resp, err := s.request(http.MethodGet, reqUrl.String(), nil, headers)

	if err != nil {
		zap.S().Errorw(errorOxrRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorOxrRequestFailed, err)
		return nil, err
	}
	return resp, nil
}

func (s *Service) parseResponseOxr(resp *http.Response) (*oxrResponse, error) {
	res := &oxrResponse{}
	err := s.decodeJson(resp, res)

	if err != nil {
		zap.S().Errorw(errorOxrResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorOxrResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesOxr(res *oxrResponse) ([]interface{}, error) {

	from := res.Base

	if !s.isCurrencySupported(from) {
		return nil, errors.New(errorOxrInvalidFrom)
	}

	if len(res.Rates) == 0 {
		return nil, errors.New(errorOxrNoResults)
	}

	var rates []interface{}
	for to, rate := range res.Rates {

		if to == from {
			continue
		}

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   from + to,
			Rate:   s.toPrecise(rate),
			Source: oxrSource,
			Volume: 1,
		})

		// prevent duplication of inverse rates, if they will be getted as direct rates
		if _, ok := s.cfg.OxrRatesDirectPairs[to+from]; ok {
			continue
		}

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   to + from,
			Rate:   s.toPrecise(1 / rate),
			Source: oxrSource,
			Volume: 1,
		})
	}

	return rates, nil
}
