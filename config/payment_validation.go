package config

// MinimumAmounts defines the minimum payment amount required for each supported currency
var MinimumAmounts = map[string]float64{
	"EUR": 10.00,
	"USD": 10.00,
	"GBP": 10.00,
}

// SupportedCurrencies defines the list of supported currencies for payments
var SupportedCurrencies = []string{"EUR", "USD", "GBP"}

// ValidatePaymentAmount checks if the amount meets minimum requirements for the currency
func ValidatePaymentAmount(amount float64, currency string) bool {
	if minAmount, ok := MinimumAmounts[currency]; ok {
		return amount >= minAmount
	}
	return false
}

// IsSupportedCurrency checks if the provided currency is supported
func IsSupportedCurrency(currency string) bool {
	for _, supported := range SupportedCurrencies {
		if currency == supported {
			return true
		}
	}
	return false
}

// GetMinimumAmount returns the minimum amount required for a currency
func GetMinimumAmount(currency string) float64 {
	if amount, ok := MinimumAmounts[currency]; ok {
		return amount
	}
	return 0
}
