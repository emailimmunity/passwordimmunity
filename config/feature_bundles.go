package config

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// Bundle defines a collection of features with pricing
type Bundle struct {
	Name        string
	Description string
	Features    []string
	Price       map[string]decimal.Decimal // Currency -> Price
}

// FeatureBundles defines available enterprise feature bundles
var FeatureBundles = map[string]Bundle{
	"enterprise_starter": {
		Name:        "Enterprise Starter",
		Description: "Essential enterprise features for growing organizations",
		Features: []string{
			"basic_reporting",
			"basic_roles",
			"basic_auth",
			"custom_roles",
		},
		Price: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(199.99),
			"EUR": decimal.NewFromFloat(179.99),
		},
	},
	"enterprise_pro": {
		Name:        "Enterprise Professional",
		Description: "Advanced features for large organizations",
		Features: []string{
			"advanced_reporting",
			"custom_roles",
			"advanced_sso",
			"basic_reporting",
			"basic_roles",
			"basic_auth",
		},
		Price: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(399.99),
			"EUR": decimal.NewFromFloat(359.99),
		},
	},
	"enterprise_ultimate": {
		Name:        "Enterprise Ultimate",
		Description: "Complete enterprise solution with all features",
		Features: []string{
			"advanced_reporting",
			"custom_roles",
			"advanced_sso",
			"multi_tenant",
			"basic_reporting",
			"basic_roles",
			"basic_auth",
		},
		Price: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(799.99),
			"EUR": decimal.NewFromFloat(719.99),
		},
	},
}

// GetBundlePrice returns the price and features for a bundle in the specified currency
func GetBundlePrice(bundleID string, currency string) (decimal.Decimal, []string, error) {
	bundle, exists := FeatureBundles[bundleID]
	if !exists {
		return decimal.Zero, nil, fmt.Errorf("unknown bundle: %s", bundleID)
	}

	price, exists := bundle.Price[currency]
	if !exists {
		return decimal.Zero, nil, fmt.Errorf("price not available in currency: %s", currency)
	}

	return price, bundle.Features, nil
}

// GetBundle returns the bundle configuration for the specified bundle ID
func GetBundle(bundleID string) (*Bundle, error) {
	bundle, exists := FeatureBundles[bundleID]
	if !exists {
		return nil, fmt.Errorf("unknown bundle: %s", bundleID)
	}
	return &bundle, nil
}
