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
| CENTRIFUGO_CHANNEL                   | -        | paysuper:admin                               | Centrifugo channel name to send alert notifications to admins   |
| METRICS_PORT                         | -        | 80                       | Port for metrics and health check                                                   |
| MICRO_REGISTRY                       | -        | -                        | Microservices registry                                                              |
| OXR_BASE_CURRENCIES                  | -        | EUR,USD                  | Base currencies to get rates from/to on openexchangerates.org                       |  
| OXR_SUPPORTED_CURRENCIES             | -        | USD, EUR, RUB, CAD, AUD, GBP, JPY, SGD, KRW, TRY, BRL, UAH, MXN, NZD, NOK, PLN, CNY, INR, CLP, PEN, COP, ZAR, HKD, TWD, THB, VND, SAR, AED, ARS, ILS, KZT, KWD, QAR, UYU, IDR, MYR, PHP | Currencies to get rates to/from base currencies on openexchangerates.org |
| OXR_APP_ID                           | true     | 1                        | API App id for openexchangerates.org                                                |
| CBRF_BASE_CURRENCIES                 | -        | EUR,USD                  | Base currencies to get rates from/to on cbr.ru (Central bank of Russia)             |
| CBEU_BASE_CURRENCIES                 | -        | USD                      | Base currencies to get rates from/to on ecb.europa.eu (Central bank of Europe)      |
| CBCA_BASE_CURRENCIES                 | -        | EUR,USD                  | Base currencies to get rates from/to on bankofcanada.ca (Central bank of Canada)    |
| CBAU_BASE_CURRENCIES                 | -        | EUR,USD                  | Base currencies to get rates from/to on rba.gov.au (Central bank of Australia)      |
| CBPL_BASE_CURRENCIES                 | -        | EUR,USD                  | Base currencies to get rates from/to on nbp.pl (Central bank of Poland)             |
| BOLLINGER_DAYS                       | -        | 7                        | Number of days for plot Bollinger functions to calculate Paysuper Prediction Rates  |
| BOLLINGER_PERIOD                     | -        | 21                       | Number of days in period for each Bollinger function                                |



## Starting the app

This application can be started in 2 modes:

* as microservice, to maintain rates requests from other components of system. This mode does not requests any rates
* as console app, to retrieve new rates from source, that passed as command line argument

Console mode can be used with cron schedule.

To start app in console mode you must set `-source` flag in command line to one of these values:

- `oxr` - to get rates from openexchangerates.org and recalculate paysuper prediction rates
- `paysuper` - to recalculate paysuper prediction rates only
- `centralbanks` - to get rates from central banks (currently from cbr.ru and ecb.europa.eu)
- `stock` - to get stock rates (currently not implemented)
- `cardpay` - to get cardpay rates (currently not implemented)

Example: `$ paysuper-currencies-rates.exe -source=oxr` runs rates requests from openexchangerates.org, and exit after it.

To run application as microservice simply don't pass any flags to command line :)  

### Example of data, store in DB:

```
{
  "_id": "5cc7030b68add4454016232d",
  "created_at": "2019-04-29T13:58:35.921Z",
  "create_date": "2019-04-29",
  "pair": "USDRUB",
  "rate": 64.679270801,
  "source": "OXR"
}
```
Where
* `_id` - record id
* `created_at` - datetime of save currency rate data to our DB
* `create_date` - date of save currency rate data to our DB, used for fast grouping rates to get first, last, min and max values by day
* `pair` - currency pair
* `rate` - currency pair rate
* `source` - code of rates source. Now use only one - XE (xe.com)
