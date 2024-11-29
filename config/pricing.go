package config

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// Currency conversion rates
var currencyRates = map[string]decimal.Decimal{
	"EUR": decimal.NewFromFloat(1.00),
	"USD": decimal.NewFromFloat(1.10),
	"GBP": decimal.NewFromFloat(0.85),
}

// Feature pricing configuration
var featurePrices = map[string]decimal.Decimal{
	"advanced_sso":       decimal.NewFromFloat(99.99),
	"custom_roles":       decimal.NewFromFloat(49.99),
	"advanced_reporting": decimal.NewFromFloat(79.99),
	"multi_tenant":       decimal.NewFromFloat(149.99),
}

// Bundle pricing configuration
var bundleConfig = map[string]struct {
	Price    decimal.Decimal
	Features []string
}{
	"security": {
		Price:    decimal.NewFromFloat(299.99),
		Features: []string{"advanced_sso", "custom_roles"},
	},
	"enterprise": {
		Price:    decimal.NewFromFloat(499.99),
		Features: []string{"advanced_sso", "custom_roles", "advanced_reporting", "multi_tenant"},
	},
	"professional": {
		Price:    decimal.NewFromFloat(199.99),
		Features: []string{"custom_roles", "advanced_reporting"},
	},
}

// GetFeaturePrice returns the price for a specific feature
func GetFeaturePrice(featureID string, currency string) (decimal.Decimal, error) {
	basePrice, ok := featurePrices[featureID]
	if !ok {
		return decimal.Zero, fmt.Errorf("invalid feature: %s", featureID)
	}

	rate, ok := currencyRates[currency]
	if !ok {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", currency)
	}

	return basePrice.Mul(rate).Round(2), nil
}

// GetBundlePrice returns the price for a specific bundle
func GetBundlePrice(bundleID string, currency string) (decimal.Decimal, []string, error) {
	bundle, ok := bundleConfig[bundleID]
	if !ok {
		return decimal.Zero, nil, fmt.Errorf("invalid bundle: %s", bundleID)
	}

	rate, ok := currencyRates[currency]
	if !ok {
		return decimal.Zero, nil, fmt.Errorf("unsupported currency: %s", currency)
	}

	return bundle.Price.Mul(rate).Round(2), bundle.Features, nil
}

// GetFeatureGracePeriod returns the grace period in days for a feature
func GetFeatureGracePeriod(featureID string) int {
	// Default grace period of 14 days
	return 14
}

// GetBaseURL returns the base URL for the application
func GetBaseURL() string {
	// TODO: Make this configurable
	return "https://api.passwordimmunity.com"
}
