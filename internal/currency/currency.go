package currency

// CurrencyProperties - set of flags of currency using allowance
type CurrencyProperties struct {
	Settlement bool
	Price      bool
	Vat        bool
	Local      bool
	Accounting bool
}

// CurrencyDefinitions - list of currencies with properties
var CurrencyDefinitions = map[string]CurrencyProperties{
	"AED": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ALL": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"AMD": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"ARS": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"AUD": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BGN": {Price: false, Settlement: false, Vat: false, Local: true, Accounting: false},
	"BHD": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BRL": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BYN": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CAD": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CHF": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CLP": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"CNY": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"COP": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"CZK": {Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"DKK": {Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"EGP": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"EUR": {Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"GBP": {Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"GHS": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"HKD": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"HRK": {Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"HUF": {Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"IDR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ILS": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"INR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ISK": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"JPY": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KES": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KRW": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KWD": {Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"KZT": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"MXN": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"MYR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"NOK": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"NZD": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PEN": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PHP": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PLN": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"QAR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"RON": {Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"RSD": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"RUB": {Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"SAR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"SEK": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"SGD": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"THB": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	// "TRY": {Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"TWD": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"TZS": {Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"UAH": {Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"USD": {Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"UYU": {Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"VND": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ZAR": {Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
}
