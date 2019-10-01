# Paysuper currency rates
[![License: GNU 3.0](https://img.shields.io/badge/License-GNU3.0-green.svg)](https://opensource.org/licenses/GNU3.0)
[![Build Status](https://travis-ci.org/paysuper/paysuper-currencies.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-currencies) 
[![codecov](https://codecov.io/gh/paysuper/paysuper-currencies/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-currencies)
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-currencies)](https://goreportcard.com/report/github.com/paysuper/paysuper-currencies)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=paysuper_paysuper-currencies&metric=alert_status)](https://sonarcloud.io/dashboard?id=paysuper_paysuper-currencies)

This service designed for sync currencies rates and store it locally with history of changes

## Environment variables:

| Name                                 | Required | Default                  | Description                                                                         |
|:-------------------------------------|:--------:|:-------------------------|:------------------------------------------------------------------------------------|
| MONGO_DSN                            | true     | -                        | MongoBD DSN connection string                                                       |
| MONGO_DIAL_TIMEOUT                   | -        | 10                       | MongoBD dial timeout in seconds                                                     |
| BROKER_ADDRESS                       | -        | amqp://127.0.0.1:5672    | RabbitMQ broker address                                                             |
| BROKER_RETRY_TIMEOUT                 | -        | 60                       | RabbitMQ broker retry timeout                                                       |
| BROKER_MAX_RETRY                     | -        | 5                        | RabbitMQ broker max retry count                                                     |
| CENTRIFUGO_URL                       | -        | http://127.0.0.1:8000    | Centrifugo url                                                                      |
| CENTRIFUGO_SECRET                    | true     | -                        | Centrifugo secret key                                                               |
| CENTRIFUGO_CHANNEL                   | -        | paysuper:admin           | Centrifugo channel name to send alert notifications to admins                       |
| METRICS_PORT                         | -        | 80                       | Port for metrics and health check                                                   |
| OXR_APP_ID                           | true     | 1                        | API App id for openexchangerates.org                                                |
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
- `stock` - to calculate stock rates

Example: `$ paysuper-currencies.exe -source=oxr` runs rates requests from openexchangerates.org, and exit after it.

To run application as microservice simply don't pass any flags to command line :)  

### Example of data, store in DB:

```
{
  "_id": "5cc7030b68add4454016232d",
  "created_at": "2019-04-29T13:58:35.921Z",
  "create_date": "2019-04-29",
  "pair": "USDRUB",
  "rate": 64.679270801,
  "source": "OXR",
  "volume": 1
}
```
Where
* `_id` - record id
* `created_at` - datetime of save currency rate data to our DB
* `create_date` - date of save currency rate data to our DB, used for fast grouping rates to get first, last, min and max values by day
* `pair` - currency pair
* `rate` - currency pair rate
* `source` - code of rates source
* `volume` - volume of excanhges, made for this rate, optional, 0 by default

## Details about rate projection 

Right now rate projection methods based on Bollinger's bands was completely disabled in the project . Evaluation currency 
rate correction with algorithm based approach has critically low range to real stock rates. According previous test this 
range should be between 1.25% and 1.75% markups to low-volatile currencies in rates OXR vs actual stock rates and from 
1.66% to 2.11% for low-volatile currencies. For current algorithm this rates 0.79%-1.19% — the balance of operations in 
this forecast does not converge. We work at a loss - we need more real data to improve it.

In current realization currencies markup process in hand-driven: each Monday accountant should calculate median to currency
rates to each currency pair in system and create recommended markups for each merchant in the system according to average 
check price and processing vs settlement amounts.

All Bollinger's based core should be restored from git history after we pass 6 months after launch and check the real 
projection of rates. Also the code with RabbitMQ subscriber was removed because the actual business process behind rates
require a lot of hand based reports and we don't see the reasons to automate it right now. 


## Contributing
We feel that a welcoming community is important and we ask that you follow PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

PaySuper welcomes contributions from anyone and everyone. Please refer to each project's style and contribution guidelines for submitting patches and additions. In general, we follow the "fork-and-pull" Git workflow.

The master branch of this repository contains the latest stable release of this component.

 
