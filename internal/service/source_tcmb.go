package service

import (
	"encoding/xml"
	"errors"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"go.uber.org/zap"
	"net/http"
)

const (
	errorTcmbRequestFailed         = "TCMB Rates request failed"
	errorTcmbResponseParsingFailed = "TCMB Rates response parsing failed"
	errorTcmbProcessRatesFailed    = "TCMB Rates save data failed"
	errorTcmbNoResults             = "TCMB Rates no results"
	errorTcmbRateDataNotFound      = "TCMB Rate data not found"
	errorTcmbNotSupported          = "TCMB Rate for curency not supported"
	errorTcmbRateUnitZero          = "TCMB Rate unit is zero"

	tcmbTo     = "TRY"
	tcmbSource = "TCMB"
	tcmbUrl    = "https://www.tcmb.gov.tr/kurlar/today.xml"
)

type tcmbResponse struct {
	XMLName xml.Name           `xml:"Tarih_Date"`
	Rates   []tcmbResponseRate `xml:"Currency"`
}

type tcmbResponseRate struct {
	XMLName         xml.Name `xml:"Currency"`
	CurrencyCode    string   `xml:"CurrencyCode,attr"`
	Unit            float64  `xml:"Unit"`
	ForexBuying     float64  `xml:"ForexBuying"`
	ForexSelling    float64  `xml:"ForexSelling"`
	BanknoteBuying  float64  `xml:"BanknoteBuying"`
	BanknoteSelling float64  `xml:"BanknoteSelling"`
}

func (s *Service) RequestRatesTcmb() error {
	zap.S().Info("Requesting rates from TCMB")

	resp, err := s.sendRequestTcmb()
	if err != nil {
		return err
	}

	res, err := s.parseResponseTcmb(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesTcmb(res)
	if err != nil {
		zap.S().Errorw(errorTcmbProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorTcmbProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from TCMB updated")

	return nil
}

func (s *Service) sendRequestTcmb() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationXML,
		headerAccept:      mimeApplicationXML,
		headerUserAgent:   defaultUserAgent,
	}

	resp, err := s.request(http.MethodGet, tcmbUrl, nil, headers)

	if err != nil {
		zap.S().Errorw(errorTcmbRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorTcmbRequestFailed, err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) parseResponseTcmb(resp *http.Response) (*tcmbResponse, error) {
	res := &tcmbResponse{}
	err := s.decodeXml(resp, res)

	if err != nil {
		zap.S().Errorw(errorTcmbResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorTcmbResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesTcmb(res *tcmbResponse) ([]interface{}, error) {

	if len(res.Rates) == 0 {
		return nil, errors.New(errorTcmbNoResults)
	}

	var rates []interface{}

	ln := len(s.cfg.RatesRequestCurrencies)
	if s.contains(s.cfg.RatesRequestCurrenciesParsed, tcmbTo) {
		ln--
	}
	counter := make(map[string]bool, ln)

	for _, rateItem := range res.Rates {

		if !s.contains(s.cfg.RatesRequestCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == tcmbTo {
			continue
		}

		rate := rateItem.BanknoteBuying
		if rate == 0.0 {
			rate = rateItem.BanknoteSelling
		}

		if rate == 0.0 {
			zap.S().Infow(errorTcmbNotSupported, "currency", rateItem.CurrencyCode)
			continue
		}

		if rateItem.Unit == 0 {
			// Hello there! We have a trouble
			zap.S().Errorw(errorTcmbRateUnitZero, "currency", rateItem.CurrencyCode)
			continue
		}

		rate = rate / rateItem.Unit

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + tcmbTo,
			Rate:   s.toPrecise(rate),
			Source: tcmbSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   tcmbTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(1 / rate),
			Source: tcmbSource,
			Volume: 1,
		})

		counter[rateItem.CurrencyCode] = true
		if len(counter) == ln {
			break
		}
	}

	for _, cFrom := range s.cfg.RatesRequestCurrencies {
		if cFrom == tcmbTo {
			continue
		}
		if _, ok := counter[cFrom]; !ok {
			zap.S().Warnw(errorTcmbRateDataNotFound, "from", cFrom, "to", tcmbTo)
		}
	}

	return rates, nil
}
