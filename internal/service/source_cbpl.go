package service

import (
	"encoding/xml"
	"errors"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"go.uber.org/zap"
	"net/http"
)

const (
	errorCbplRequestFailed         = "CBPL Rates request failed"
	errorCbplResponseParsingFailed = "CBPL Rates response parsing failed"
	errorCbplProcessRatesFailed    = "CBPL Rates save data failed"
	errorCbplNoResults             = "CBPL Rates no results"
	errorCbplRateDataNotFound      = "CBPL Rate data not found"

	cbplTo     = "PLN"
	cbplSource = "CBPL"
	cbplUrl    = "https://www.nbp.pl/kursy/xml/en/19a092en.xml"
)

type cbplResponse struct {
	XMLName xml.Name           `xml:"exchange_rates"`
	Rates   []cbplResponseRate `xml:"mid-rate"`
}

type cbplResponseRate struct {
	CurrencyCode string  `xml:"code,attr"`
	Units        float64 `xml:"units,attr"`
	Value        float64 `xml:",chardata"`
}

// RequestRatesCbpl - retriving current rates from Central bank of Poland
func (s *Service) RequestRatesCbpl() error {
	zap.S().Info("Requesting rates from CBPL")

	resp, err := s.sendRequestCbpl()
	if err != nil {
		return err
	}

	res, err := s.parseResponseCbpl(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesCbpl(res)
	if err != nil {
		zap.S().Errorw(errorCbplProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorCbplProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from CBPL updated")

	return nil
}

func (s *Service) sendRequestCbpl() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationXML,
		headerAccept:      mimeTextXML,
	}

	resp, err := s.request(http.MethodGet, cbplUrl, nil, headers)

	if err != nil {
		zap.S().Errorw(errorCbplRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorCbplRequestFailed, err)
		return nil, err
	}
	return resp, nil
}

func (s *Service) parseResponseCbpl(resp *http.Response) (*cbplResponse, error) {
	res := &cbplResponse{}
	err := s.decodeXml(resp, res)

	if err != nil {
		zap.S().Errorw(errorCbplResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorCbplResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesCbpl(res *cbplResponse) ([]interface{}, error) {

	if len(res.Rates) == 0 {
		return nil, errors.New(errorCbplNoResults)
	}

	var rates []interface{}

	ln := len(s.cfg.RatesRequestCurrencies)
	if s.contains(s.cfg.RatesRequestCurrenciesParsed, cbplTo) {
		ln--
	}
	counter := make(map[string]bool, ln)

	for _, rateItem := range res.Rates {

		if !s.contains(s.cfg.RatesRequestCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == cbplTo {
			continue
		}

		rate := rateItem.Value

		rateByNominal := rate / rateItem.Units

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + cbplTo,
			Rate:   s.toPrecise(rateByNominal),
			Source: cbplSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbplTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(1 / rateByNominal),
			Source: cbplSource,
			Volume: 1,
		})

		counter[rateItem.CurrencyCode] = true
		if len(counter) == ln {
			break
		}
	}

	for _, cFrom := range s.cfg.RatesRequestCurrencies {
		if cFrom == cbplTo {
			continue
		}
		if _, ok := counter[cFrom]; !ok {
			zap.S().Warnw(errorCbplRateDataNotFound, "from", cFrom, "to", cbplTo)
		}
	}

	return rates, nil
}
