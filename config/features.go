package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"github.com/shopspring/decimal"
)

// FeatureConfig defines an enterprise feature configuration
type FeatureConfig struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	Tier        string          `json:"tier"`
	GracePeriod int            `json:"grace_period"`
}

// BundleConfig defines a bundle of enterprise features
type BundleConfig struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Features    []string        `json:"features"`
	Price       decimal.Decimal `json:"price"`
	Tier        string          `json:"tier"`
}

// Default enterprise features configuration
var defaultEnterpriseFeatures = map[string]FeatureConfig{
	"advanced_sso": {
		Name:        "Advanced SSO Integration",
		Description: "Enterprise-grade Single Sign-On with SAML and OIDC support",
		Price:       decimal.NewFromFloat(49.99),
		Tier:        "enterprise",
		GracePeriod: 14,
	},
	"multi_tenant": {
		Name:        "Multi-Tenant Management",
		Description: "Advanced organization and user management capabilities",
		Price:       decimal.NewFromFloat(79.99),
		Tier:        "enterprise",
		GracePeriod: 14,
	},
	"advanced_audit": {
		Name:        "Advanced Audit Logs",
		Description: "Detailed audit logging and reporting capabilities",
		Price:       decimal.NewFromFloat(29.99),
		Tier:        "business",
		GracePeriod: 7,
	},
}

// Default feature bundles configuration
var defaultBundles = map[string]BundleConfig{
	"security": {
		Name:        "Security Bundle",
		Description: "Complete security feature set including SSO and audit",
		Features:    []string{"advanced_sso", "advanced_audit"},
		Price:       decimal.NewFromFloat(69.99),
		Tier:        "enterprise",
	},
	"business": {
		Name:        "Business Bundle",
		Description: "Complete business feature set",
		Features:    []string{"multi_tenant", "advanced_audit"},
		Price:       decimal.NewFromFloat(99.99),
		Tier:        "business",
	},
}

// loadEnterpriseFeatures loads enterprise feature configuration from environment or defaults
func loadEnterpriseFeatures() map[string]FeatureConfig {
	features := make(map[string]FeatureConfig)

	// Try to load from environment variable
	if featuresJSON := os.Getenv("ENTERPRISE_FEATURES"); featuresJSON != "" {
		if err := json.Unmarshal([]byte(featuresJSON), &features); err == nil {
			return features
		}
	}

	// Fall back to default features
	return defaultEnterpriseFeatures
}

// getEnvInt gets an integer from environment with fallback
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetFeaturePrice returns the price for a specific feature in a specific currency
func GetFeaturePrice(featureID string, currency string) (decimal.Decimal, error) {
	if feature, exists := defaultEnterpriseFeatures[featureID]; exists {
		basePrice := decimal.NewFromFloat(feature.Price)
		return convertCurrency(basePrice, "USD", currency)
	}
	return decimal.Zero, fmt.Errorf("feature %s not found", featureID)
}

// IsEnterpriseFeature checks if a feature is an enterprise feature
func IsEnterpriseFeature(featureID string) bool {
	_, exists := defaultEnterpriseFeatures[featureID]
	return exists
}

// GetFeatureTier returns the tier required for a specific feature
func GetFeatureTier(featureID string) string {
	if feature, exists := defaultEnterpriseFeatures[featureID]; exists {
		return feature.Tier
	}
	return ""
}

// GetFeatureGracePeriod returns the grace period for a specific feature
func GetFeatureGracePeriod(featureID string) int {
	if feature, exists := defaultEnterpriseFeatures[featureID]; exists {
		return feature.GracePeriod
	}
	return 0
}

// GetBundlePrice returns the price for a specific bundle
func GetBundlePrice(bundleID string) float64 {
	if bundle, exists := defaultBundles[bundleID]; exists {
		return bundle.Price
	}
	return 0
}

// GetFeaturesInBundle returns the features included in a bundle
func GetFeaturesInBundle(bundleID string) []string {
	if bundle, exists := defaultBundles[bundleID]; exists {
		return bundle.Features
	}
	return nil
}

// GetAllBundles returns all available feature bundles
func GetAllBundles() []BundleConfig {
	bundles := make([]BundleConfig, 0, len(defaultBundles))
	for _, bundle := range defaultBundles {
		bundles = append(bundles, bundle)
	}
	return bundles
}

// GetAllFeatures returns all available enterprise features
func GetAllFeatures() []FeatureConfig {
	features := make([]FeatureConfig, 0, len(defaultEnterpriseFeatures))
	for _, feature := range defaultEnterpriseFeatures {
		features = append(features, feature)
	}
	return features
}

// List of supported currencies
var supportedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
}

// convertCurrency converts an amount from one currency to another
func convertCurrency(amount decimal.Decimal, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	if !supportedCurrencies[fromCurrency] {
		return decimal.Zero, fmt.Errorf("unsupported source currency: %s", fromCurrency)
	}

	if !supportedCurrencies[toCurrency] {
		return decimal.Zero, fmt.Errorf("unsupported target currency: %s", toCurrency)
	}

	// TODO: Implement real currency conversion using exchange rates
	// For now, assume 1:1 conversion rate for supported currencies
	return amount, nil
}
