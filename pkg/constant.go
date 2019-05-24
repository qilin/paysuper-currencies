package pkg

const (
	// ServiceName - name of microservice
	ServiceName = "paysupercurrencies"

	// Version - version of service
	Version = "latest"

	// RateTypeOxr - rate type value for Oxr rates
	RateTypeOxr = "oxr"

	// RateTypeCentralbanks - rate type value for central banks rates
	RateTypeCentralbanks = "centralbanks"

	// RateTypePaysuper - rate type value for Paysuper rates
	RateTypePaysuper = "paysuper"

	// RateTypeStock - rate type value for Stock rates
	RateTypeStock = "stock"

	// RateTypeCardpay - rate type value for Cardpay rates
	RateTypeCardpay = "cardpay"

	// CardpayTopicRateData - rabbitMq topic name for Cardpay rate data
	CardpayTopicRateData = "cardpay-rate-data"

	// CardpayTopicRateData - rabbitMq topic name for Cardpay rate data retry
	CardpayTopicRateDataRetry = "cardpay-rate-data-retry"

	// CardpayTopicRateData - rabbitMq topic name for Cardpay rates finished
	CardpayTopicRateDataFinished = "cardpay-rate-data-finished"
)
