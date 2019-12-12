# Paysuper currency rates
[![License: GNU 3.0](https://img.shields.io/badge/License-GNU3.0-green.svg)](https://opensource.org/licenses/GNU3.0)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/paysuper/paysuper-currencies/issues)
[![Build Status](https://travis-ci.org/paysuper/paysuper-currencies.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-currencies)
[![codecov](https://codecov.io/gh/paysuper/paysuper-currencies/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-currencies)
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-currencies)](https://goreportcard.com/report/github.com/paysuper/paysuper-currencies)

This service is designed to synchronize currencies rates and to store it locally with a history of changes.

***

## Features

* Importing "OXR" and central banks currency rates.
* Calculating stock rates.
* Storing a rates' history of changes.

## Table of Contents

- [Developing](#developing)
    - [Branches](#branches)
    - [Start the application](#start-the-application)
    - [Environment variables](#environment-variables)
    - [Correction rules](#storing)
    - [Exchange directions](#storing)
    - [Storing](#storing)
- [Contributing](#contributing)
- [License](#license)

## Developing

### Branches

We use the [GitFlow](https://nvie.com/posts/a-successful-git-branching-model) as a branching model for Git.

### Start the application

Application can be started in two modes:

* **microservice** - to maintain rates requests from other components of system. This mode does not request any rates. To run application as microservice don't pass any flags to a command line.
* **console mode** - to retrieve new rates from a source that has been passed as a command line argument. The console mode can be used with a cron schedule.

To start an application in a console mode you need to set a `-source` flag in a command line with one of following values:

* `oxr` - to get the rates from openexchangerates.org and recalculate the PaySuper prediction rates.
* `paysuper` - to recalculate only the PaySuper prediction rates.
* `centralbanks` - to get the rates from central banks (currently from cbr.ru and ecb.europa.eu).
* `stock` - to calculate the stock rates.

This is an example of command that runs rates requests from openexchangerates.org and at the end exits the application:

```bash
paysuper-currencies.exe -source=oxr
```

### Environment variables

| Name                                 | Required | Default                  | Description                                                                         |
|:-------------------------------------|:--------:|:-------------------------|:------------------------------------------------------------------------------------|
| OXR_APP_ID                           | true     | 1                        | API App id for openexchangerates.org                                                |
| MONGO_DSN                            | true     | -                        | MongoBD DSN connection string                                                       |
| MONGO_DIAL_TIMEOUT                   | -        | 10                       | MongoBD dial timeout in seconds                                                     |
| CENTRIFUGO_URL                       | -        | http://127.0.0.1:8000    | Centrifugo url                                                                      |
| CENTRIFUGO_SECRET                    | true     | -                        | Centrifugo secret key                                                               |
| CENTRIFUGO_CHANNEL                   | -        | paysuper:admin           | Centrifugo channel name to send alert notifications to admins                       |
| METRICS_PORT                         | -        | 80                       | Port for metrics and health check                                                   |

## Correction rules

There are correction rules for a rate. The correction rules can be applied at the moment of a rate or exchange request processing.

The correction rules can be defined for: 
* a combination of RateType, ExchangeDirection, Merchant,
* optionally, for some currencies' pair (or for all pairs by default).

The system correction can be defined for:
* a combination of RateType, ExchangeDirection,
* optionally, some currencies' pair (or for all pairs by default).

## Exchange directions

There are two directions for exchange and rates requests: `buy` and `sell`. The direction affects the application of the correction rules for rates and exchages.

* Exchange direction `buy` increases an exchange rate for a percent determined in the corresponding correction rule, and decreases a result amount.
* Exchange direction `sell` decreases an exchange rate for a percent determined in the corresponding correction rule, and increases a result amount.

### Storing

Example of a currency rate stored in the PaySuper database:

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

Attribute|Description
---|---
`_id`|The ID of a record.
`created_at`|The datetime of saving a currency rate to the PaySuper database.
`create_date`|The date of save currency rate to the PaySuper database. It's used for a fast grouping rates to get the first, the last, min and max values by days.
`pair`|The currency's pair.
`rate`|The currency's pair rate.
`source`|The code of a rates source.
`volume`|The volume of exchanges that has been made for this rate. Optional. Default value equals to 0.


## Contributing

If you like this project then you can put a ⭐️ on it.

We welcome contributions to PaySuper of any kind including documentation, suggestions, bug reports, pull requests etc. We would love to hear from you. In general, we follow the "fork-and-pull" Git workflow.

We feel that a welcoming community is important and we ask that you follow the PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

## License

The project is available as open source under the terms of the [GPL v3 License](https://www.gnu.org/licenses/gpl-3.0).