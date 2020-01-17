package pkg

const (
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

	ExchangeDirectionSell = "sell"
	ExchangeDirectionBuy  = "buy"

	ErrorDatabaseQueryFailed          = "Query to database collection failed"
	ErrorDatabaseFieldCollection      = "collection"
	ErrorDatabaseFieldDocumentId      = "document_id"
	ErrorDatabaseFieldQuery           = "query"
	ErrorDatabaseFieldSet             = "set"
	ErrorDatabaseFieldSorts           = "sorts"
	ErrorDatabaseFieldLimit           = "limit"
	ErrorDatabaseFieldOffset          = "offset"
	ErrorDatabaseFieldOperation       = "operation"
	ErrorDatabaseFieldOperationInsert = "insert"
	ErrorDatabaseFieldOperationUpdate = "update"
	ErrorDatabaseFieldOperationUpsert = "upsert"
	ErrorDatabaseFieldDocument        = "document"
)

var (
	SupportedExchangeDirections = map[string]bool{
		ExchangeDirectionSell: true,
		ExchangeDirectionBuy:  true,
	}
)
