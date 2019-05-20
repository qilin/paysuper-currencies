package config

import (
    "github.com/kelseyhightower/envconfig"
    "github.com/paysuper/paysuper-currencies/internal/currency"
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

    OxrAppId               string   `envconfig:"OXR_APP_ID" required:"true"`

    BollingerDays   int `envconfig:"BOLLINGER_DAYS" default:"7"`
    BollingerPeriod int `envconfig:"BOLLINGER_PERIOD" default:"21"`

    RatesTypes map[string]bool

    Currencies map[string]currency.CurrencyProperties

    SettlementCurrencies   []string
    PriceCurrencies        []string
    VatCurrencies          []string
    AccountingCurrencies   []string
    RatesRequestCurrencies []string
    SupportedCurrencies    []string

    SupportedCurrenciesParsed    map[string]bool
    SettlementCurrenciesParsed   map[string]bool
    RatesRequestCurrenciesParsed map[string]bool
}

func NewConfig() (*Config, error) {
    cfg := &Config{}
    err := envconfig.Process("", cfg)

    cfg.RatesTypes = make(map[string]bool, 5)
    cfg.RatesTypes[pkg.RateTypeOxr] = true
    cfg.RatesTypes[pkg.RateTypeCentralbanks] = true
    cfg.RatesTypes[pkg.RateTypePaysuper] = true
    cfg.RatesTypes[pkg.RateTypeStock] = true
    cfg.RatesTypes[pkg.RateTypeCardpay] = true

    cfg.Currencies = currency.CurrencyDefinitions
    cfg.SupportedCurrenciesParsed = make(map[string]bool, len(cfg.Currencies))

    for code, properties := range cfg.Currencies {
        cfg.SupportedCurrencies = append(cfg.SupportedCurrencies, code)
        cfg.SupportedCurrenciesParsed[code] = true

        if properties.Price || properties.Vat {
            cfg.RatesRequestCurrencies = append(cfg.RatesRequestCurrencies, code)
        }

        if properties.Price {
            cfg.PriceCurrencies = append(cfg.PriceCurrencies, code)
        }

        if properties.Vat {
            cfg.VatCurrencies = append(cfg.VatCurrencies, code)
        }

        if properties.Settlement {
            cfg.SettlementCurrencies = append(cfg.SettlementCurrencies, code)
        }

        if properties.Accounting {
            cfg.AccountingCurrencies = append(cfg.AccountingCurrencies, code)
        }
    }

    cfg.SettlementCurrenciesParsed = make(map[string]bool, len(cfg.SettlementCurrencies))
    for _, v := range cfg.SettlementCurrencies {
        cfg.SettlementCurrenciesParsed[v] = true
    }

    cfg.RatesRequestCurrenciesParsed = make(map[string]bool, len(cfg.RatesRequestCurrencies))
    for _, v := range cfg.RatesRequestCurrencies {
        cfg.RatesRequestCurrenciesParsed[v] = true
    }

    return cfg, err
}
