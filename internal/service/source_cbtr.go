package service

import (
	"encoding/xml"
	"errors"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
	"go.uber.org/zap"
	"net/http"
)

const (
	errorCbtrRequestFailed         = "CBTR Rates request failed"
	errorCbtrResponseParsingFailed = "CBTR Rates response parsing failed"
	errorCbtrProcessRatesFailed    = "CBTR Rates save data failed"
	errorCbtrNoResults             = "CBTR Rates no results"
	errorCbtrRateDataNotFound      = "CBTR Rate data not found"
	errorCbtrNotSupported          = "CBTR Rate for curency not supported"
	errorCbtrRateUnitZero          = "CBTR Rate unit is zero"

	cbtrTo     = "TRY"
	cbtrSource = "CBTR"
	cbtrUrl    = "https://www.cbtr.gov.tr/kurlar/today.xml"
)

type cbtrResponse struct {
	XMLName xml.Name           `xml:"Tarih_Date"`
	Rates   []cbtrResponseRate `xml:"Currency"`
}

type cbtrResponseRate struct {
	XMLName         xml.Name `xml:"Currency"`
	CurrencyCode    string   `xml:"CurrencyCode,attr"`
	Unit            float64  `xml:"Unit"`
	ForexBuying     float64  `xml:"ForexBuying"`
	ForexSelling    float64  `xml:"ForexSelling"`
	BanknoteBuying  float64  `xml:"BanknoteBuying"`
	BanknoteSelling float64  `xml:"BanknoteSelling"`
}

func (s *Service) RequestRatesCbtr() error {
	zap.S().Info("Requesting rates from CBTR")

	resp, err := s.sendRequestCbtr()
	if err != nil {
		return err
	}

	res, err := s.parseResponseCbtr(resp)
	if err != nil {
		return err
	}

	rates, err := s.processRatesCbtr(res)
	if err != nil {
		zap.S().Errorw(errorCbtrProcessRatesFailed, "error", err)
		s.sendCentrifugoMessage(errorCbtrProcessRatesFailed, err)
		return err
	}

	err = s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
	if err != nil {
		return err
	}

	zap.S().Info("Rates from CBTR updated")

	return nil
}

func (s *Service) sendRequestCbtr() (*http.Response, error) {
	headers := map[string]string{
		headerContentType: mimeApplicationXML,
		headerAccept:      mimeApplicationXML,
		headerUserAgent:   defaultUserAgent,
	}

	resp, err := s.request(http.MethodGet, cbtrUrl, nil, headers)

	if err != nil {
		zap.S().Errorw(errorCbtrRequestFailed, "error", err)
		s.sendCentrifugoMessage(errorCbtrRequestFailed, err)
		return nil, err
	}

	return resp, nil
}

func (s *Service) parseResponseCbtr(resp *http.Response) (*cbtrResponse, error) {
	res := &cbtrResponse{}
	err := s.decodeXml(resp, res)

	if err != nil {
		zap.S().Errorw(errorCbtrResponseParsingFailed, "error", err)
		s.sendCentrifugoMessage(errorCbtrResponseParsingFailed, err)
		return nil, err
	}

	return res, nil
}

func (s *Service) processRatesCbtr(res *cbtrResponse) ([]interface{}, error) {

	if len(res.Rates) == 0 {
		return nil, errors.New(errorCbtrNoResults)
	}

	var rates []interface{}

	ln := len(s.cfg.RatesRequestCurrencies)
	if s.contains(s.cfg.RatesRequestCurrenciesParsed, cbtrTo) {
		ln--
	}
	counter := make(map[string]bool, ln)

	for _, rateItem := range res.Rates {

		if !s.contains(s.cfg.RatesRequestCurrenciesParsed, rateItem.CurrencyCode) {
			continue
		}

		if rateItem.CurrencyCode == cbtrTo {
			continue
		}

		rate := rateItem.BanknoteBuying
		if rate == 0.0 {
			rate = rateItem.BanknoteSelling
		}

		if rate == 0.0 {
			zap.S().Infow(errorCbtrNotSupported, "currency", rateItem.CurrencyCode)
			continue
		}

		if rateItem.Unit == 0 {
			// Hello there! We have a trouble
			zap.S().Errorw(errorCbtrRateUnitZero, "currency", rateItem.CurrencyCode)
			continue
		}

		rate = rate / rateItem.Unit

		// direct pair
		rates = append(rates, &currencies.RateData{
			Pair:   rateItem.CurrencyCode + cbtrTo,
			Rate:   s.toPrecise(rate),
			Source: cbtrSource,
			Volume: 1,
		})

		// inverse pair
		rates = append(rates, &currencies.RateData{
			Pair:   cbtrTo + rateItem.CurrencyCode,
			Rate:   s.toPrecise(1 / rate),
			Source: cbtrSource,
			Volume: 1,
		})

		counter[rateItem.CurrencyCode] = true
		if len(counter) == ln {
			break
		}
	}

	for _, cFrom := range s.cfg.RatesRequestCurrencies {
		if cFrom == cbtrTo {
			continue
		}
		if _, ok := counter[cFrom]; !ok {
			zap.S().Warnw(errorCbtrRateDataNotFound, "from", cFrom, "to", cbtrTo)
		}
	}

	return rates, nil
}
