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

    OxrSupportedCurrencies []string `envconfig:"OXR_SUPPORTED_CURRENCIES" default:"USD,EUR,RUB,CAD,AUD,GBP,JPY,SGD,KRW,TRY,BRL,UAH,MXN,NZD,NOK,PLN,CNY,INR,CLP,PEN,COP,ZAR,HKD,TWD,THB,VND,SAR,AED,ARS,ILS,KZT,KWD,QAR,UYU,IDR,MYR,PHP"`
    OxrBaseCurrencies      []string `envconfig:"OXR_BASE_CURRENCIES" default:"EUR,USD"`
    OxrAppId               string   `envconfig:"OXR_APP_ID" required:"true"`

    CbrfBaseCurrencies []string `envconfig:"CBRF_BASE_CURRENCIES" default:"EUR,USD"`
    CbeuBaseCurrencies []string `envconfig:"CBEU_BASE_CURRENCIES" default:"USD"`

    BollingerDays   int `envconfig:"BOLLINGER_DAYS" default:"7"`
    BollingerPeriod int `envconfig:"BOLLINGER_PERIOD" default:"21"`
}

func NewConfig() (*Config, error) {
    cfg := &Config{}
    err := envconfig.Process("", cfg)

    return cfg, err
}
