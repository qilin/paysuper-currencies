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
	Value        float64 `xml:",chardata"`
}

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
		HeaderContentType: MIMEApplicationXML,
		HeaderAccept:      MIMETextXML,
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

	ln := len(s.cfg.SettlementCurrencies)
	if s.contains(s.cfg.SettlementCurrenciesParsed, cbplTo) {
		ln--
	}
	c := 0

	for _, rateItem := range res.Rates {

		if !s.contains(s.cfg.SettlementCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == cbplTo {
			continue
		}

		rate := rateItem.Value

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + cbplTo,
			Rate:   s.toPrecise(rate),
			Source: cbplSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbplTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(1 / rate),
			Source: cbplSource,
			Volume: 1,
		})

		c++
		if c == ln {
			break
		}
	}

	return rates, nil
}
