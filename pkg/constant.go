package pkg

const (
    // ServiceName - name of microservice
    ServiceName = "paysupercurrencies"

    // Version - version of service
    Version = "latest"

    RateTypeOxr          = "oxr"
    RateTypeCentralbanks = "centralbanks"
    RateTypePaysuper     = "paysuper"
    RateTypeStock        = "stock"
    RateTypeCardpay      = "cardpay"

    CardpayTopicRateData         = "cardpay-rate-data"
    CardpayTopicRateDataRetry    = "cardpay-rate-data-retry"
    CardpayTopicRateDataFinished = "cardpay-rate-data-finished"
)
