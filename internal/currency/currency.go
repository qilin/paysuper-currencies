package currency

// CurrencyProperties - set of flags of currency using allowance
type CurrencyProperties struct {
	Settlement bool
	Price      bool
	Vat        bool
	Accounting bool
}

// CurrencyDefinitions - list of currencies with properties
var CurrencyDefinitions = map[string]CurrencyProperties{
	"AED": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ALL": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"AMD": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"ARS": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"AUD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"BHD": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"BRL": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"BYN": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"CAD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"CHF": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"CLP": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"CNY": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"COP": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"DKK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"EGP": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"EUR": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"GBP": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"GHS": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"HKD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"IDR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ILS": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"INR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ISK": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"JPY": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"KES": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"KRW": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"KWD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"KZT": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"MXN": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"MYR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"NOK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"NZD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"PEN": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"PHP": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"PLN": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"QAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"RSD": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"RUB": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"SAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"SEK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"SGD": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"THB": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"TRY": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"TWD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"TZS": {Price: false, Settlement: false, Vat: true, Accounting: false},
	"UAH": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"USD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"UYU": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"VND": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ZAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
}
