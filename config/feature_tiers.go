package config

import (
	"github.com/shopspring/decimal"
)

// Tier represents a pricing tier with associated features
type Tier struct {
	Name         string
	Description  string
	Features     []string
	MonthlyPrice map[string]decimal.Decimal
	YearlyPrice  map[string]decimal.Decimal
}

// Bundle represents a group of features that can be purchased together
type Bundle struct {
	Name         string
	Description  string
	Features     []string
	MonthlyPrice map[string]decimal.Decimal
	YearlyPrice  map[string]decimal.Decimal
}

// Feature represents a single feature with its metadata
type Feature struct {
	Name         string
	Description  string
	IsEnterprise bool
}

// FeatureTiers defines the available pricing tiers
var FeatureTiers = map[string]Tier{
	"free": {
		Name:        "Free",
		Description: "Basic password management features",
		Features: []string{
			"basic_auth",
			"basic_roles",
			"basic_reporting",
		},
		MonthlyPrice: map[string]decimal.Decimal{
			"USD": decimal.Zero,
			"EUR": decimal.Zero,
		},
		YearlyPrice: map[string]decimal.Decimal{
			"USD": decimal.Zero,
			"EUR": decimal.Zero,
		},
	},
	"premium": {
		Name:        "Premium",
		Description: "Advanced features for individuals and small teams",
		Features: []string{
			"basic_auth",
			"basic_roles",
			"basic_reporting",
			"custom_roles",
			"advanced_reporting",
		},
		MonthlyPrice: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(9.99),
			"EUR": decimal.NewFromFloat(8.99),
		},
		YearlyPrice: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(99.99),
			"EUR": decimal.NewFromFloat(89.99),
		},
	},
	"enterprise": {
		Name:        "Enterprise",
		Description: "Complete solution for organizations",
		Features: []string{
			"basic_auth",
			"basic_roles",
			"basic_reporting",
			"custom_roles",
			"advanced_reporting",
			"advanced_sso",
			"multi_tenant",
			"advanced_audit",
			"directory_sync",
			"emergency_access",
			"custom_policies",
		},
		MonthlyPrice: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(199.99),
			"EUR": decimal.NewFromFloat(17.99),
		},
		YearlyPrice: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(199.99),
			"EUR": decimal.NewFromFloat(179.99),
		},
	},
}

// GetTierFeatures returns the list of features included in a tier
func GetTierFeatures(tierID string) ([]string, error) {
	tier, exists := FeatureTiers[tierID]
	if !exists {
		return nil, fmt.Errorf("unknown tier: %s", tierID)
	}
	return tier.Features, nil
}

// GetTierPrice returns the price for a tier based on billing period
func GetTierPrice(tierID, currency string, yearly bool) (decimal.Decimal, error) {
	tier, exists := FeatureTiers[tierID]
	if !exists {
		return decimal.Zero, fmt.Errorf("unknown tier: %s", tierID)
	}

	var prices map[string]decimal.Decimal
	if yearly {
		prices = tier.YearlyPrice
	} else {
		prices = tier.MonthlyPrice
	}

	price, exists := prices[currency]
	if !exists {
		return decimal.Zero, fmt.Errorf("price not available in currency: %s", currency)
	}

	return price, nil
}

// IsFeatureInTier checks if a feature is included in a specific tier
func IsFeatureInTier(tierID string, featureID string) bool {
	tier, exists := FeatureTiers[tierID]
	if !exists {
		return false
	}

	for _, f := range tier.Features {
		if f == featureID {
			return true
		}
	}
	return false
}

// GetBundleFeatures returns the list of features included in a bundle
func GetBundleFeatures(bundleID string) ([]string, error) {
	bundle, exists := FeatureBundles[bundleID]
	if !exists {
		return nil, fmt.Errorf("unknown bundle: %s", bundleID)
	}
	return bundle.Features, nil
}

// GetBundlePrice returns the price for a bundle based on billing period
func GetBundlePrice(bundleID, currency string, yearly bool) (decimal.Decimal, error) {
	bundle, exists := FeatureBundles[bundleID]
	if !exists {
		return decimal.Zero, fmt.Errorf("unknown bundle: %s", bundleID)
	}

	var prices map[string]decimal.Decimal
	if yearly {
		prices = bundle.YearlyPrice
	} else {
		prices = bundle.MonthlyPrice
	}

	price, exists := prices[currency]
	if !exists {
		return decimal.Zero, fmt.Errorf("price not available in currency: %s", currency)
	}

	return price, nil
}

// IsEnterpriseFeature checks if a feature is an enterprise feature
func IsEnterpriseFeature(featureID string) bool {
	feature, exists := Features[featureID]
	if !exists {
		return false
	}
	return feature.IsEnterprise
}
