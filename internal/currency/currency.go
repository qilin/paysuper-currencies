package currency

type CurrencyProperties struct {
	Settlement bool
	Price      bool
	Vat        bool
	Accounting bool
}

var CurrencyDefinitions = map[string]CurrencyProperties{
	"USD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"EUR": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"RUB": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"CAD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"AUD": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"GBP": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"JPY": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"SGD": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"KRW": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"TRY": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"BRL": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"UAH": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"MXN": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"NZD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"NOK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"SEK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"DKK": {Price: true, Settlement: true, Vat: false, Accounting: true},
	"PLN": {Price: true, Settlement: true, Vat: true, Accounting: true},
	"CNY": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"INR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"CLP": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"PEN": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"COP": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ZAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"HKD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"TWD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"THB": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"VND": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"SAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"AED": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ARS": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"ILS": {Price: true, Settlement: false, Vat: true, Accounting: false},
	"KZT": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"KWD": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"QAR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"UYU": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"IDR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"MYR": {Price: true, Settlement: false, Vat: false, Accounting: false},
	"PHP": {Price: true, Settlement: false, Vat: false, Accounting: false},
}
