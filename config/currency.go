package config

import "strings"

var supportedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
	"JPY": true,
	"AUD": true,
	"CAD": true,
	"CHF": true,
	"CNY": true,
	"SEK": true,
	"NZD": true,
}

// IsSupportedCurrency checks if the given currency code is supported
func IsSupportedCurrency(currency string) bool {
	return supportedCurrencies[strings.ToUpper(currency)]
}

// GetSupportedCurrencies returns a list of all supported currency codes
func GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(supportedCurrencies))
	for currency := range supportedCurrencies {
		currencies = append(currencies, currency)
	}
	return currencies
}
