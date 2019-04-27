package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
    MongoHost     string `envconfig:"MONGO_HOST" required:"true"`
    MongoDatabase string `envconfig:"MONGO_DB" required:"true"`
    MongoUser     string `envconfig:"MONGO_USER" default:""`
    MongoPassword string `envconfig:"MONGO_PASSWORD" default:""`

    MetricsPort int `envconfig:"METRICS_PORT" required:"false" default:"80"`

    CentrifugoSecret string `envconfig:"CENTRIFUGO_SECRET" required:"true"`
    CentrifugoURL    string `envconfig:"CENTRIFUGO_URL" required:"false" default:"http://127.0.0.1:8000"`

    MicroRegistry string `envconfig:"MICRO_REGISTRY" required:"false"`

    XeSupportedCurrencies []string           `envconfig:"XE_SUPPORTED_CURRENCIES" default:"USD,EUR,RUB,CAD,AUD,GBP,JPY,SGD,KRW,TRY,BRL,UAH,MXN,NZD,NOK,PLN,CNY,INR,CLP,PEN,COP,ZAR,HKD,TWD,THB,VND,SAR,AED,ARS,ILS,KZT,KWD,QAR,UYU,IDR,MYR,PHP"`
    XeBaseCurrencies      []string           `envconfig:"XE_BASE_CURRENCIES" default:"EUR,USD"`
    XeCommonCorrection    float64            `envconfig:"XE_COMMON_CORRECTION" default:"1"`
    XePairsCorrections    map[string]float64 `envconfig:"XE_PAIRS_CORRECTIONS" default:""`
    XeAuthCredentials     string             `envconfig:"XE_AUTH_CREDENTIALS" required:"true"`
    XeRatesRequestPeriod  int64              `envconfig:"XE_RATES_REQUEST_PERIOD" default:"24"` // in hours
}

func NewConfig() (*Config, error) {
    cfg := &Config{}
    err := envconfig.Process("", cfg)

    return cfg, err
}
