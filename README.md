# Paysuper currency rates
[![License: GNU 3.0](https://img.shields.io/badge/License-GNU3.0-green.svg)](https://opensource.org/licenses/GNU3.0)
[![Build Status](https://travis-ci.org/paysuper/paysuper-currencies-rates.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-currencies-rates) 
[![codecov](https://codecov.io/gh/paysuper/paysuper-currencies-rates/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-currencies-rates)
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-currencies-rates)](https://goreportcard.com/report/github.com/paysuper/paysuper-currencies-rates)

This service designed for sync currencies rates and store it locally with history of changes

## Environment variables:

| Name                                 | Required | Default                  | Description                                                                         |
|:-------------------------------------|:--------:|:-------------------------|:------------------------------------------------------------------------------------|
| MONGO_HOST                           | true     | -                        | MongoDb host address                                                                |
| MONGO_DB                             | true     | -                        | MongoDb database name                                                               |
| MONGO_USER                           | -        | -                        | MongoDb user                                                                        |
| MONGO_PASSWORD                       | -        | -                        | MongoDb password                                                                    |
| CENTRIFUGO_URL                       | -        | http://127.0.0.1:8000    | Centrifugo url                                                                      |
| CENTRIFUGO_KEY                       | true     | -                        | Centrifugo secret key                                                               |
| METRICS_PORT                         | -        | 80                       | Port for metrics and health check                                                   |
| MICRO_REGISTRY                       | -        | -                        | Microservices registry                                                              |
| OXR_BASE_CURRENCIES                  | -        | EUR,USD                  | Base currencies to get rates from/to on openexchangerates.org                       |  
| OXR_SUPPORTED_CURRENCIES             | -        | USD, EUR, RUB, CAD, AUD, GBP, JPY, SGD, KRW, TRY, BRL, UAH, MXN, NZD, NOK, PLN, CNY, INR, CLP, PEN, COP, ZAR, HKD, TWD, THB, VND, SAR, AED, ARS, ILS, KZT, KWD, QAR, UYU, IDR, MYR, PHP | Currencies to get rates to/from base currencies on openexchangerates.org |
| OXR_APP_ID                           | true     | 1                        | API App id for openexchangerates.org                                                |


### Example of data, store in DB:

```
{
  "_id": "5cc7030b68add4454016232d",
  "created_at": "2019-04-29T13:58:35.921Z",
  "pair": "USDRUB",
  "rate": 64.679270801,
  "source": "XE"
}
```
Where
* `_id` - record id
* `created_at` - datetime of save currency rate data to our DB
* `pair` - currency pair
* `rate` - currency pair rate
* `correction` - correction percent for pair, if set
* `corrected_rate` - currency pair rate corrected with persent above. If not set is equal to `rate` field
* `is_cb_rate` - flag for rates, given from local central banks (reserved for future features)
* `source` - code of rates source. Now use only one - XE (xe.com)
