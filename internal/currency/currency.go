package currency

// CurrencyProperties - set of flags of currency using allowance
type CurrencyProperties struct {
	Settlement bool
	Price      bool
	Vat        bool
	Local      bool
	Accounting bool
	Precision  int32
}

// CurrencyDefinitions - list of currencies with properties
var CurrencyDefinitions = map[string]CurrencyProperties{
	"AED": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ALL": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"AMD": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"ARS": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"AUD": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BGN": {Precision: 2, Price: false, Settlement: false, Vat: false, Local: true, Accounting: false},
	"BHD": {Precision: 3, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BRL": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"BYN": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CAD": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CHF": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"CLP": {Precision: 0, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"CNY": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"COP": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"CZK": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"DKK": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"EGP": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"EUR": {Precision: 2, Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"GBP": {Precision: 2, Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"GHS": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"HKD": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"HRK": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"HUF": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"IDR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ILS": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"INR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ISK": {Precision: 0, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"JPY": {Precision: 0, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KES": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KRW": {Precision: 0, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"KWD": {Precision: 3, Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"KZT": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"MXN": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"MYR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"NOK": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"NZD": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PEN": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PHP": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"PLN": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"QAR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"RON": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: true, Accounting: false},
	"RSD": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"RUB": {Precision: 2, Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"SAR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"SEK": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"SGD": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"THB": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	// "TRY": {Precision: 2, Price: true, Settlement: false, Vat: true, Local: true, Accounting: false},
	"TWD": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"TZS": {Precision: 2, Price: false, Settlement: false, Vat: true, Local: true, Accounting: false},
	"UAH": {Precision: 2, Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"USD": {Precision: 2, Price: true, Settlement: true, Vat: true, Local: true, Accounting: true},
	"UYU": {Precision: 2, Price: false, Settlement: false, Vat: false, Local: false, Accounting: false},
	"VND": {Precision: 0, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
	"ZAR": {Precision: 2, Price: true, Settlement: false, Vat: false, Local: false, Accounting: false},
}
