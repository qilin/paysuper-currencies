package service

import (
    "encoding/xml"
    "errors"
    "github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
    "go.uber.org/zap"
    "net/http"
)

const (
    errorCbauRequestFailed         = "CBAU Rates request failed"
    errorCbauResponseParsingFailed = "CBAU Rates response parsing failed"
    errorCbauSaveRatesFailed       = "CBAU Rates save data failed"
    errorCbauNoResults             = "CBAU Rates no results"

    cbauTo     = "AUD"
    cbauSource = "CBAU"
    cbauUrl    = "https://www.rba.gov.au/rss/rss-cb-exchange-rates.xml"
)

type cbauResponse struct {
    XMLName xml.Name `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
    Rates   []cbauResponseItem   `xml:"item"`
}

type cbauResponseItem struct {
    Statistics cbauResponseStatistics `xml:"http://www.cbwiki.net/wiki/index.php/Specification_1.2/ statistics"`
}

type cbauResponseStatistics struct {
    ExchangeRate cbauResponseExchangeRate `xml:"http://www.cbwiki.net/wiki/index.php/Specification_1.2/ exchangeRate"`
}

type cbauResponseExchangeRate struct {
    TargetCurrency string `xml:"http://www.cbwiki.net/wiki/index.php/Specification_1.2/ targetCurrency"`
    Observation cbauResponseObservation `xml:"http://www.cbwiki.net/wiki/index.php/Specification_1.2/ observation"`
}

type cbauResponseObservation struct {
    Value float64 `xml:"http://www.cbwiki.net/wiki/index.php/Specification_1.2/ value"`
}

func (s *Service) RequestRatesCbau() error {
    zap.S().Info("Requesting rates from CBAU")

    headers := map[string]string{
        HeaderContentType: MIMEApplicationXML,
        HeaderAccept:      MIMEApplicationXML,
    }

    resp, err := s.request(http.MethodGet, cbauUrl, nil, headers)

    if err != nil {
        zap.S().Errorw(errorCbauRequestFailed, "error", err)
        s.sendCentrifugoMessage(errorCbauRequestFailed, err)
        return err
    }

    res := &cbauResponse{}
    err = s.decodeXml(resp, res)

    if err != nil {
        zap.S().Errorw(errorCbauResponseParsingFailed, "error", err)
        s.sendCentrifugoMessage(errorCbauResponseParsingFailed, err)
        return err
    }

    err = s.processRatesCbau(res)

    if err != nil {
        zap.S().Errorw(errorCbauSaveRatesFailed, "error", err)
        s.sendCentrifugoMessage(errorCbauSaveRatesFailed, err)
        return err
    }

    zap.S().Info("Rates from CBAU updated")

    return nil
}

func (s *Service) processRatesCbau(res *cbauResponse) error {

    if len(res.Rates) == 0 {
        return errors.New(errorCbauNoResults)
    }

    var rates []interface{}

    l := len(s.cfg.CbauBaseCurrencies)
    c := 0

    for _, rateItem := range res.Rates {

        cFrom := rateItem.Statistics.ExchangeRate.TargetCurrency

        if !s.contains(s.cfg.CbauBaseCurrenciesParsed, cFrom) {
            continue
        }

        rate := rateItem.Statistics.ExchangeRate.Observation.Value

        // direct pair
        rates = append(rates, &currencies.RateData{
            Pair:   cFrom + cbauTo,
            Rate:   s.toPrecise(rate),
            Source: cbauSource,
        })

        // inverse pair
        rates = append(rates, &currencies.RateData{
            Pair:   cbauTo + cFrom,
            Rate:   s.toPrecise(1 / rate),
            Source: cbauSource,
        })

        c++
        if c == l {
            break
        }
    }

    err := s.saveRates(collectionRatesNameSuffixCentralbanks, rates)
    if err != nil {
        return err
    }
    return nil
}
