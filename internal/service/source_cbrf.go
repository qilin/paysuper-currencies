package service

import (
	"encoding/xml"
	"errors"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

const (
	errorCbrfRequestFailed         = "CBRF Rates request failed"
	errorCbrfResponseParsingFailed = "CBRF Rates response parsing failed"
	errorCbrfParseFloatError       = "CBRF Rates parse float error"
	errorCbrfProcessRatesFailed    = "CBRF Rates save data failed"
	errorCbrfNoResults             = "CBRF Rates no results"
	errorCbrfRateDataNotFound      = "CBRF Rate data not found"

	cbrfTo     = "RUB"
	cbrfSource = "CBRF"
	cbrfUrl    = "http://www.cbr.ru/scripts/XML_daily.asp"
)

type cbrfResponse struct {
	XMLName xml.Name           `xml:"ValCurs"`
	Rates   []cbrfResponseRate `xml:"Valute"`
}

type cbrfResponseRate struct {
	XMLName      xml.Name `xml:"Valute"`
	CurrencyCode string   `xml:"CharCode"`
	Value        string   `xml:"Value"`
}

// RequestRatesCbrf - retriving current rates from Central bank of Russia
func (s *Service) RequestRatesCbrf() error {
	zap.S().Info("Requesting rates from CBRF")

	resp, err := s.sendRequestCbrf()
	if err != nil {
		return err
	}

	res, err := s.parseResponseCbrf(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesCbrf(res)
	if err != nil {
		zap.S().Errorw(errorCbrfProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorCbrfProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from CBRF updated")

	return nil
}

func (s *Service) sendRequestCbrf() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationXML,
		headerAccept:      mimeApplicationXML,
	}

	// here may be 302 redirect in answer - https://toster.ru/q/149039
	resp, err := s.request(http.MethodGet, cbrfUrl, nil, headers)

	if err != nil {
		zap.S().Errorw(errorCbrfRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorCbrfRequestFailed, err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) parseResponseCbrf(resp *http.Response) (*cbrfResponse, error) {
	res := &cbrfResponse{}
	err := s.decodeXml(resp, res)

	if err != nil {
		zap.S().Errorw(errorCbrfResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorCbrfResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesCbrf(res *cbrfResponse) ([]interface{}, error) {

	if len(res.Rates) == 0 {
		return nil, errors.New(errorCbrfNoResults)
	}

	var rates []interface{}

	ln := len(s.cfg.RatesRequestCurrencies)
	if s.contains(s.cfg.RatesRequestCurrenciesParsed, cbrfTo) {
		ln--
	}
	counter := make(map[string]bool, ln)

	for _, rateItem := range res.Rates {

		if !s.contains(s.cfg.RatesRequestCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == cbrfTo {
			continue
		}

		var rate float64
		rateStr := strings.Replace(rateItem.Value, ",", ".", -1)
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			return nil, errors.New(errorCbrfParseFloatError)
		}

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + cbrfTo,
			Rate:   s.toPrecise(rate),
			Source: cbrfSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbrfTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(1 / rate),
			Source: cbrfSource,
			Volume: 1,
		})

		counter[rateItem.CurrencyCode] = true
		if len(counter) == ln {
			break
		}
	}

	for _, cFrom := range s.cfg.RatesRequestCurrencies {
		if cFrom == cbauTo {
			continue
		}
		if _, ok := counter[cFrom]; !ok {
			zap.S().Warnw(errorCbrfRateDataNotFound, "from", cFrom, "to", cbauTo)
		}
	}

	return rates, nil
}
