package internal

import (
    "encoding/xml"
    "errors"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "go.uber.org/zap"
    "net/http"
)

const (
    errorCbplRequestFailed         = "CBPL Rates request failed"
    errorCbplResponseParsingFailed = "CBPL Rates response parsing failed"
    errorCbplSaveRatesFailed       = "CBPL Rates save data failed"
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

    headers := map[string]string{
        HeaderContentType: MIMEApplicationXML,
        HeaderAccept:      MIMETextXML,
    }

    resp, err := s.request(http.MethodGet, cbplUrl, nil, headers)

    if err != nil {
        zap.S().Errorw(errorCbplRequestFailed, "error", err)
        s.sendCentrifugoMessage(errorCbplRequestFailed, err)
        return err
    }

    res := &cbplResponse{}
    err = s.decodeXml(resp, res)

    if err != nil {
        zap.S().Errorw(errorCbplResponseParsingFailed, "error", err)
        s.sendCentrifugoMessage(errorCbplResponseParsingFailed, err)
        return err
    }

    err = s.processRatesCbpl(res)

    if err != nil {
        zap.S().Errorw(errorCbplSaveRatesFailed, "error", err)
        s.sendCentrifugoMessage(errorCbplSaveRatesFailed, err)
        return err
    }

    zap.S().Info("Rates from CBPL updated")

    return nil
}

func (s *Service) processRatesCbpl(res *cbplResponse) error {

    if len(res.Rates) == 0 {
        return errors.New(errorCbplNoResults)
    }

    rates := []*currencyrates.RateData{}

    l := len(s.cfg.CbplBaseCurrencies)
    c := 0

    for _, rateItem := range res.Rates {

        if !s.contains(s.cfg.CbplBaseCurrencies, rateItem.CurrencyCode) {
            continue
        }

        rate := rateItem.Value

        // direct pair
        rates = append(rates, &currencyrates.RateData{
            Pair:   rateItem.CurrencyCode + cbplTo,
            Rate:   s.toPrecise(rate),
            Source: cbplSource,
        })

        // inverse pair
        rates = append(rates, &currencyrates.RateData{
            Pair:   cbplTo + rateItem.CurrencyCode,
            Rate:   s.toPrecise(1 / rate),
            Source: cbplSource,
        })

        c++
        if c == l {
            break
        }
    }

    err := s.saveRates(collectionSuffixCb, rates)
    if err != nil {
        return err
    }
    return nil
}
