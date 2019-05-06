package internal

import (
    "errors"
    "fmt"
    "github.com/paysuper/paysuper-currencies-rates/proto"
    "go.uber.org/zap"
    "net/http"
    "net/url"
    "strings"
)

const (
    errorXeUrlValidationFailed   = "XE Rates url validation failed"
    errorXeRequestFailed         = "XE Rates request failed"
    errorXeResponseParsingFailed = "XE Rates response parsing failed"
    errorXeSaveRatesFailed       = "XE Rates save data failed"
    errorXeNoResults             = "XE Rates no results"
    errorXeInvalidFrom           = "XE Rates invalid from"

    xeSource = "XE"

    xeUrlTemplate = "https://xecdapi.xe.com/v1/historic_rate.json?inverse=true&from=%s&%s"
)

type xeRatesResponse struct {
    Terms     string
    Privacy   string
    From      string
    Amount    float64
    Timestamp string
    To        []*xeRateItem
}

type xeRateItem struct {
    Quotecurrency string
    Mid           float64
    Inverse       float64
}

func (s *Service) RequestRatesXe() {

    params := url.Values{"to": []string{strings.Join(s.cfg.XeSupportedCurrencies, ",")}}
    to := params.Encode()

    headers := map[string]string{
        HeaderContentType:   MIMEApplicationJSON,
        HeaderAccept:        MIMEApplicationJSON,
        HeaderAuthorization: fmt.Sprintf(BasicAuthorization, s.cfg.XeAuthCredentials),
    }

    for _, from := range s.cfg.XeBaseCurrencies {

        reqUrl, err := s.validateUrl(fmt.Sprintf(xeUrlTemplate, from, to))

        if err != nil {
            zap.S().Errorw(errorXeUrlValidationFailed, "error", err)
            s.sendCentrifugoMessage(errorXeUrlValidationFailed, err)
            continue
        }

        resp, err := s.request(http.MethodGet, reqUrl.String(), nil, headers)

        if err != nil {
            zap.S().Errorw(errorXeRequestFailed, "error", err)
            s.sendCentrifugoMessage(errorXeRequestFailed, err)
            continue
        }

        res := &xeRatesResponse{}
        err = s.getJson(resp, res)

        if err != nil {
            zap.S().Errorw(errorXeResponseParsingFailed, "error", err)
            s.sendCentrifugoMessage(errorXeResponseParsingFailed, err)
            continue
        }

        err = s.processRatesXe(res)

        if err != nil {
            zap.S().Errorw(errorXeSaveRatesFailed, "error", err)
            s.sendCentrifugoMessage(errorXeSaveRatesFailed, err)
            continue
        }
    }

    return
}

func (s *Service) processRatesXe(res *xeRatesResponse) error {

    from := res.From

    if !s.isCurrencySupported(from) {
        return errors.New(errorXeInvalidFrom)
    }

    if len(res.To) == 0 {
        return errors.New(errorXeNoResults)
    }


    for _, target := range res.To {
        to := target.Quotecurrency
        if to == from {
            continue
        }

        // direct pair
        err := s.saveRateXe(from+to, target.Mid)
        if err != nil {
            return err
        }

        // inverse pair
        err = s.saveRateXe(to+from, target.Inverse)
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Service) saveRateXe(pair string, rate float64) error {
    correction := s.getCorrectionForPair(pair)
    return s.saveRate(&currencyrates.RateData{
        Pair:          pair,
        Rate:          rate,
        Correction:    correction,
        CorrectedRate: rate * correction,
        Source:        xeSource,
        IsCbRate:      false,
    })
}
