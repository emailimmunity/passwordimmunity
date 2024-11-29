package config

// EnterpriseFeature represents a feature that requires enterprise licensing
type EnterpriseFeature struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Prices      map[string]float64 `json:"prices"` // Currency -> Price mapping
	Tier        string             `json:"tier"`
	GracePeriod int               `json:"grace_period_days"`
	Bundle      string             `json:"bundle"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

// FeatureBundle represents a collection of features offered as a package
type FeatureBundle struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Prices      map[string]float64 `json:"prices"` // Currency -> Price mapping
	Features    []string           `json:"features"`
}

// DefaultEnterpriseFeatures defines the default enterprise features and their configurations
var DefaultEnterpriseFeatures = []EnterpriseFeature{
	{
		ID:          "advanced_sso",
		Name:        "Advanced SSO Integration",
		Description: "Enterprise-grade Single Sign-On with SAML and OIDC support",
		Prices: map[string]float64{
			"EUR": 49.99,
			"USD": 54.99,
			"GBP": 44.99,
		},
		Tier:        "enterprise",
		GracePeriod: 14,
		Bundle:      "security",
	},
	{
		ID:          "multi_tenant",
		Name:        "Multi-Tenant Management",
		Description: "Advanced organization and tenant management capabilities",
		Prices: map[string]float64{
			"EUR": 79.99,
			"USD": 89.99,
			"GBP": 69.99,
		},
		Tier:        "enterprise",
		GracePeriod: 14,
		Bundle:      "management",
		Dependencies: []string{"advanced_policy"},
	},
	{
		ID:          "advanced_audit",
		Name:        "Advanced Audit Logging",
		Description: "Comprehensive audit logging and reporting",
		Prices: map[string]float64{
			"EUR": 39.99,
			"USD": 44.99,
			"GBP": 34.99,
		},
		Tier:        "enterprise",
		GracePeriod: 14,
		Bundle:      "compliance",
	},
	{
		ID:          "advanced_policy",
		Name:        "Advanced Policy Management",
		Description: "Enterprise policy controls and enforcement",
		Prices: map[string]float64{
			"EUR": 59.99,
			"USD": 64.99,
			"GBP": 54.99,
		},
		Tier:        "enterprise",
		GracePeriod: 14,
		Bundle:      "management",
	},
}

// DefaultFeatureBundles defines the available feature bundles
var DefaultFeatureBundles = []FeatureBundle{
	{
		ID:          "security",
		Name:        "Security Suite",
		Description: "Complete security and authentication features",
		Prices: map[string]float64{
			"EUR": 89.99,
			"USD": 99.99,
			"GBP": 79.99,
		},
		Features:    []string{"advanced_sso", "advanced_policy"},
	},
	{
		ID:          "management",
		Name:        "Management Suite",
		Description: "Advanced organization and policy management",
		Prices: map[string]float64{
			"EUR": 129.99,
			"USD": 139.99,
			"GBP": 119.99,
		},
		Features:    []string{"multi_tenant", "advanced_policy"},
	},
	{
		ID:          "compliance",
		Name:        "Compliance Suite",
		Description: "Comprehensive audit and compliance features",
		Price:       69.99,
		Features:    []string{"advanced_audit"},
	},
}

// GetFeatureByID returns an enterprise feature by its ID
func GetFeatureByID(featureID string) (EnterpriseFeature, bool) {
	for _, feature := range DefaultEnterpriseFeatures {
		if feature.ID == featureID {
			return feature, true
		}
	}
	return EnterpriseFeature{}, false
}

// IsEnterpriseFeature checks if a feature ID corresponds to an enterprise feature
func IsEnterpriseFeature(featureID string) bool {
	_, exists := GetFeatureByID(featureID)
	return exists
}

// GetFeaturePrice returns the price for a specific feature in the specified currency
func GetFeaturePrice(featureID string, currency string) float64 {
	if feature, exists := GetFeatureByID(featureID); exists {
		if price, ok := feature.Prices[currency]; ok {
			return price
		}
	}
	return 0
}

// GetFeatureGracePeriod returns the grace period for a specific feature
func GetFeatureGracePeriod(featureID string) int {
	if feature, exists := GetFeatureByID(featureID); exists {
		return feature.GracePeriod
	}
	return 0
}

// GetBundleByID returns a feature bundle by its ID
func GetBundleByID(bundleID string) (FeatureBundle, bool) {
	for _, bundle := range DefaultFeatureBundles {
		if bundle.ID == bundleID {
			return bundle, true
		}
	}
	return FeatureBundle{}, false
}

// GetBundlePrice returns the price for a specific bundle in the specified currency
func GetBundlePrice(bundleID string, currency string) float64 {
	if bundle, exists := GetBundleByID(bundleID); exists {
		if price, ok := bundle.Prices[currency]; ok {
			return price
		}
	}
	return 0
}

// GetFeaturesInBundle returns all features included in a bundle
func GetFeaturesInBundle(bundleID string) []string {
	if bundle, exists := GetBundleByID(bundleID); exists {
		return bundle.Features
	}
	return []string{}
}

// IsFeatureActive checks if a feature is active for an organization
func IsFeatureActive(featureID string, organizationID string) bool {
	if feature, exists := GetFeatureByID(featureID); exists {
		if service := GetEnterpriseService(); service != nil {
			if active, err := service.IsFeatureActive(context.Background(), featureID, organizationID); err == nil {
				return active
			}
		}
	}
	return false
}

// GetActiveFeatures returns a list of active feature IDs for an organization
func GetActiveFeatures(organizationID string) []string {
	var activeFeatures []string
	service := GetEnterpriseService()
	if service == nil {
		return activeFeatures
	}

	for featureID := range Features {
		if active, err := service.IsFeatureActive(context.Background(), featureID, organizationID); err == nil && active {
			activeFeatures = append(activeFeatures, featureID)
		}
	}
	return activeFeatures
}

// GetActiveBundles returns a list of active bundle IDs for an organization
func GetActiveBundles(organizationID string) []string {
	activeFeatures := GetActiveFeatures(organizationID)
	activeBundles := make(map[string]bool)

	// Check each bundle to see if all its features are active
	for bundleID, bundle := range Bundles {
		allActive := true
		for _, featureID := range bundle.Features {
			isActive := false
			for _, activeFeature := range activeFeatures {
				if activeFeature == featureID {
					isActive = true
					break
				}
			}
			if !isActive {
				allActive = false
				break
			}
		}
		if allActive {
			activeBundles[bundleID] = true
		}
	}

	// Convert map to slice
	var result []string
	for bundleID := range activeBundles {
		result = append(result, bundleID)
	}
	return result
}

// GetSupportedCurrencies returns the list of supported currencies
func GetSupportedCurrencies() []string {
	return []string{"EUR", "USD", "GBP"}
}

// GetDefaultCurrency returns the default currency for pricing
func GetDefaultCurrency() string {
	return "EUR"
}

// RequiresEnterpriseLicense checks if a feature requires enterprise licensing
func RequiresEnterpriseLicense(featureID string) bool {
	feature, exists := GetFeatureByID(featureID)
	return exists && feature.Tier == "enterprise"
}
