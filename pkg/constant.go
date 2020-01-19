package pkg

import (
	"github.com/paysuper/paysuper-proto/go/currenciespb"
)

const (
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
		currenciespb.ExchangeDirectionSell: true,
		currenciespb.ExchangeDirectionBuy:  true,
	}
)
