package internal

import (
    "encoding/xml"
    "errors"
    "github.com/paysuper/paysuper-currencies-rates/pkg/proto/currencyrates"
    "go.uber.org/zap"
    "net/http"
    "strconv"
    "strings"
)

const (
    errorCbrfRequestFailed         = "CBRF Rates request failed"
    errorCbrfResponseParsingFailed = "CBRF Rates response parsing failed"
    errorCbrfParseFloatError       = "CBRF Rates parse float error"
    errorCbrfSaveRatesFailed       = "CBRF Rates save data failed"
    errorCbrfNoResults             = "CBRF Rates no results"

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

func (s *Service) RequestRatesCbrf() error {
    zap.S().Info("Requesting rates from CBRF")

    headers := map[string]string{
        HeaderContentType: MIMEApplicationXML,
        HeaderAccept:      MIMEApplicationXML,
    }

    // here may be 302 redirect in answer - https://toster.ru/q/149039
    resp, err := s.request(http.MethodGet, cbrfUrl, nil, headers)

    if err != nil {
        zap.S().Errorw(errorCbrfRequestFailed, "error", err)
        s.sendCentrifugoMessage(errorCbrfRequestFailed, err)
        return err
    }

    res := &cbrfResponse{}
    err = s.decodeXml(resp, res)

    if err != nil {
        zap.S().Errorw(errorCbrfResponseParsingFailed, "error", err)
        s.sendCentrifugoMessage(errorCbrfResponseParsingFailed, err)
        return err
    }

    err = s.processRatesCbrf(res)

    if err != nil {
        zap.S().Errorw(errorCbrfSaveRatesFailed, "error", err)
        s.sendCentrifugoMessage(errorCbrfSaveRatesFailed, err)
        return err
    }

    zap.S().Info("Rates from CBRF updated")

    return nil
}

func (s *Service) processRatesCbrf(res *cbrfResponse) error {

    if len(res.Rates) == 0 {
        return errors.New(errorCbrfNoResults)
    }

    rates := []*currencyrates.RateData{}

    l := len(s.cfg.CbrfBaseCurrencies)
    c := 0

    for _, rateItem := range res.Rates {

        if !s.contains(s.cfg.CbrfBaseCurrencies, rateItem.CurrencyCode) {
            continue
        }

        var rate float64
        rateStr := strings.Replace(rateItem.Value, ",", ".", -1)
        rate, err := strconv.ParseFloat(rateStr, 64)
        if err != nil {
            return errors.New(errorCbrfParseFloatError)
        }

        // direct pair
        rates = append(rates, &currencyrates.RateData{
            Pair:   rateItem.CurrencyCode + cbrfTo,
            Rate:   s.toPrecise(rate),
            Source: cbrfSource,
        })

        // inverse pair
        rates = append(rates, &currencyrates.RateData{
            Pair:   cbrfTo + rateItem.CurrencyCode,
            Rate:   s.toPrecise(1 / rate),
            Source: cbrfSource,
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
