package config

import (
    "github.com/kelseyhightower/envconfig"
    "github.com/paysuper/paysuper-currencies/pkg"
)

type Config struct {
    MongoHost     string `envconfig:"MONGO_HOST" required:"true"`
    MongoDatabase string `envconfig:"MONGO_DB" required:"true"`
    MongoUser     string `envconfig:"MONGO_USER" default:""`
    MongoPassword string `envconfig:"MONGO_PASSWORD" default:""`

    MetricsPort int `envconfig:"METRICS_PORT" required:"false" default:"80"`

    CentrifugoSecret  string `envconfig:"CENTRIFUGO_SECRET" required:"true"`
    CentrifugoURL     string `envconfig:"CENTRIFUGO_URL" required:"false" default:"http://127.0.0.1:8000"`
    CentrifugoChannel string `envconfig:"CENTRIFUGO_CHANNEL" default:"paysuper:admin"`

    MicroRegistry string `envconfig:"MICRO_REGISTRY" required:"false"`

    OxrSupportedCurrencies []string `envconfig:"OXR_SUPPORTED_CURRENCIES" default:"USD,EUR,RUB,CAD,AUD,GBP,JPY,SGD,KRW,TRY,BRL,UAH,MXN,NZD,NOK,PLN,CNY,INR,CLP,PEN,COP,ZAR,HKD,TWD,THB,VND,SAR,AED,ARS,ILS,KZT,KWD,QAR,UYU,IDR,MYR,PHP"`
    OxrBaseCurrencies      []string `envconfig:"OXR_BASE_CURRENCIES" default:"EUR,USD"`
    OxrAppId               string   `envconfig:"OXR_APP_ID" required:"true"`

    CbrfBaseCurrencies []string `envconfig:"CBRF_BASE_CURRENCIES" default:"EUR,USD"`
    CbeuBaseCurrencies []string `envconfig:"CBEU_BASE_CURRENCIES" default:"USD"`
    CbcaBaseCurrencies []string `envconfig:"CBCA_BASE_CURRENCIES" default:"EUR,USD"`
    CbauBaseCurrencies []string `envconfig:"CBAU_BASE_CURRENCIES" default:"EUR,USD"`
    CbplBaseCurrencies []string `envconfig:"CBPL_BASE_CURRENCIES" default:"EUR,USD"`

    BollingerDays   int `envconfig:"BOLLINGER_DAYS" default:"7"`
    BollingerPeriod int `envconfig:"BOLLINGER_PERIOD" default:"21"`

    OxrSupportedCurrenciesParsed map[string]bool
    CbrfBaseCurrenciesParsed     map[string]bool
    CbeuBaseCurrenciesParsed     map[string]bool
    CbauBaseCurrenciesParsed     map[string]bool
    CbplBaseCurrenciesParsed     map[string]bool

    RatesTypes map[string]bool
}

func NewConfig() (*Config, error) {
    cfg := &Config{}
    err := envconfig.Process("", cfg)

    cfg.OxrSupportedCurrenciesParsed = make(map[string]bool, len(cfg.OxrSupportedCurrencies))
    for _, v := range cfg.OxrSupportedCurrencies {
        cfg.OxrSupportedCurrenciesParsed[v] = true
    }

    cfg.CbrfBaseCurrenciesParsed = make(map[string]bool, len(cfg.CbrfBaseCurrencies))
    for _, v := range cfg.CbrfBaseCurrencies {
        cfg.CbrfBaseCurrenciesParsed[v] = true
    }

    cfg.CbeuBaseCurrenciesParsed = make(map[string]bool, len(cfg.CbeuBaseCurrencies))
    for _, v := range cfg.CbeuBaseCurrencies {
        cfg.CbeuBaseCurrenciesParsed[v] = true
    }

    cfg.CbauBaseCurrenciesParsed = make(map[string]bool, len(cfg.CbauBaseCurrencies))
    for _, v := range cfg.CbauBaseCurrencies {
        cfg.CbauBaseCurrenciesParsed[v] = true
    }

    cfg.CbplBaseCurrenciesParsed = make(map[string]bool, len(cfg.CbplBaseCurrencies))
    for _, v := range cfg.CbplBaseCurrencies {
        cfg.CbplBaseCurrenciesParsed[v] = true
    }

    cfg.RatesTypes = make(map[string]bool, 5)
    cfg.RatesTypes[pkg.RateTypeOxr] = true
    cfg.RatesTypes[pkg.RateTypeCentralbanks] = true
    cfg.RatesTypes[pkg.RateTypePaysuper] = true
    cfg.RatesTypes[pkg.RateTypeStock] = true
    cfg.RatesTypes[pkg.RateTypeCardpay] = true

    return cfg, err
}
