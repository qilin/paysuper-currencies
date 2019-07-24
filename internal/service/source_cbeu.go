package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"net/http"
)

const (
	errorCbeuUrlValidationFailed   = "CBEU Rates url validation failed"
	errorCbeuRequestFailed         = "CBEU Rates request failed"
	errorCbeuResponseParsingFailed = "CBEU Rates response parsing failed"
	errorCbeuProcessRatesFailed    = "CBEU Rates save data failed"
	errorCbeuNoResults             = "CBEU Rates no results"
	errorCbeuRateDataNotFound      = "CBEU Rate data not found"

	cbeuTo          = "EUR"
	cbeuSource      = "CBEU"
	cbeuUrlTemplate = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml?%s"
)

type cbeuResponse struct {
	XMLName xml.Name          `xml:"http://www.gesmes.org/xml/2002-08-01 Envelope"`
	Data    cbeuResponseCube1 `xml:"Cube"`
}

type cbeuResponseCube1 struct {
	XMLName xml.Name          `xml:"Cube"`
	Rates   cbeuResponseCube2 `xml:"Cube"`
}

type cbeuResponseCube2 struct {
	XMLName xml.Name            `xml:"Cube"`
	Rates   []cbeuResponseCube3 `xml:"Cube"`
}

type cbeuResponseCube3 struct {
	XMLName      xml.Name `xml:"Cube"`
	CurrencyCode string   `xml:"currency,attr"`
	Value        float64  `xml:"rate,attr"`
}

// RequestRatesCbeu - retriving current rates from European Central bank
func (s *Service) RequestRatesCbeu() error {
	zap.S().Info("Requesting rates from CBEU")

	resp, err := s.sendRequestCbeu()
	if err != nil {
		return err
	}

	res, err := s.parseResponseCbeu(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesCbeu(res)
	if err != nil {
		zap.S().Errorw(errorCbeuProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorCbeuProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from CBEU updated")

	return nil
}

func (s *Service) sendRequestCbeu() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationXML,
		headerAccept:      mimeApplicationXML,
	}

	reqUrl, err := s.validateUrl(fmt.Sprintf(cbeuUrlTemplate, uuid.NewV4().String()))

	if err != nil {
		zap.S().Errorw(errorCbeuUrlValidationFailed, "error", err)
		s.sendCentrifugoMessage(errorCbeuUrlValidationFailed, err)
		return nil, err
	}

	resp, err := s.request(http.MethodGet, reqUrl.String(), nil, headers)

	if err != nil {
		zap.S().Errorw(errorCbeuRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorCbeuRequestFailed, err)
		return nil, err
	}
	return resp, nil
}

func (s *Service) parseResponseCbeu(resp *http.Response) (*cbeuResponse, error) {
	res := &cbeuResponse{}
	err := s.decodeXml(resp, res)

	if err != nil {
		zap.S().Errorw(errorCbeuResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorCbeuResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesCbeu(res *cbeuResponse) ([]interface{}, error) {

	if len(res.Data.Rates.Rates) == 0 {
		return nil, errors.New(errorCbeuNoResults)
	}

	var rates []interface{}

	ln := len(s.cfg.RatesRequestCurrencies)
	if s.contains(s.cfg.RatesRequestCurrenciesParsed, cbeuTo) {
		ln--
	}
	counter := make(map[string]bool, ln)

	for _, rateItem := range res.Data.Rates.Rates {

		if !s.contains(s.cfg.RatesRequestCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == cbeuTo {
			continue
		}

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbeuTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(rateItem.Value),
			Source: cbeuSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + cbeuTo,
			Rate:   s.toPrecise(1 / rateItem.Value),
			Source: cbeuSource,
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
			zap.S().Warnw(errorCbeuRateDataNotFound, "from", cFrom, "to", cbauTo)
		}
	}

	return rates, nil
}
