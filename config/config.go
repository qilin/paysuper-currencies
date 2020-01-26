package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/paysuper/paysuper-currencies/internal/currency"
	"github.com/paysuper/paysuper-proto/go/currenciespb"
)

// Config is struct for store service configuration
type Config struct {
	MongoDsn         string `envconfig:"MONGO_DSN" required:"true"`
	MongoDialTimeout string `envconfig:"MONGO_DIAL_TIMEOUT" required:"false" default:"10"`

	MetricsPort int `envconfig:"METRICS_PORT" required:"false" default:"80"`

	CentrifugoSecret  string `envconfig:"CENTRIFUGO_SECRET" required:"true"`
	CentrifugoURL     string `envconfig:"CENTRIFUGO_URL" required:"false" default:"http://127.0.0.1:8000"`
	CentrifugoChannel string `envconfig:"CENTRIFUGO_CHANNEL" default:"paysuper:admin"`

	OxrAppId string `envconfig:"OXR_APP_ID" required:"true"`

	RatesTypes map[string]bool

	Currencies map[string]currency.CurrencyProperties

	CurrenciesPrecision map[string]int32

	SettlementCurrencies   []string
	PriceCurrencies        []string
	VatCurrencies          []string
	AccountingCurrencies   []string
	RatesRequestCurrencies []string
	SupportedCurrencies    []string

	SupportedCurrenciesParsed    map[string]bool
	SettlementCurrenciesParsed   map[string]bool
	RatesRequestCurrenciesParsed map[string]bool

	OxrRatesDirectPairs map[string]bool

	MicroSelector string `envconfig:"MICRO_SELECTOR" default:""`
}

// NewConfig return new config
func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	cfg.RatesTypes = make(map[string]bool, 5)
	cfg.RatesTypes[currenciespb.RateTypeOxr] = true
	cfg.RatesTypes[currenciespb.RateTypeCentralbanks] = true
	cfg.RatesTypes[currenciespb.RateTypePaysuper] = true
	cfg.RatesTypes[currenciespb.RateTypeStock] = true
	cfg.RatesTypes[currenciespb.RateTypeCardpay] = true

	cfg.Currencies = currency.CurrencyDefinitions
	cfg.SupportedCurrenciesParsed = make(map[string]bool, len(cfg.Currencies))
	cfg.CurrenciesPrecision = make(map[string]int32, len(cfg.Currencies))

	for code, properties := range cfg.Currencies {
		cfg.SupportedCurrencies = append(cfg.SupportedCurrencies, code)
		cfg.SupportedCurrenciesParsed[code] = true
		cfg.CurrenciesPrecision[code] = properties.Precision

		if properties.Price || properties.Vat || properties.Local {
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

	cfg.OxrRatesDirectPairs = make(map[string]bool)
	for _, from := range cfg.SettlementCurrencies {
		for _, to := range cfg.RatesRequestCurrencies {
			cfg.OxrRatesDirectPairs[from+to] = true
		}
	}

	return cfg, err
}
