package licensing

import (
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/shopspring/decimal"
)

// FeatureTier represents the pricing tier of a feature
type FeatureTier struct {
	RequiresPayment bool            `json:"requires_payment"`
	MinimumAmount   decimal.Decimal `json:"minimum_amount"`
	BundleID        string          `json:"bundle_id,omitempty"`
	IsEnterprise    bool            `json:"is_enterprise"`
}

// GetFeatureTier returns the pricing tier information for a feature
func (s *Service) GetFeatureTier(featureID string) *FeatureTier {
	// Check if feature is part of enterprise features
	if config.IsEnterpriseFeature(featureID) {
		return &FeatureTier{
			RequiresPayment: true,
			MinimumAmount:   config.GetFeaturePrice(featureID),
			IsEnterprise:    true,
		}
	}

	// Check if feature is part of any bundle
	for _, bundleID := range config.GetAllBundles() {
		features := config.GetFeaturesInBundle(bundleID)
		for _, f := range features {
			if f == featureID {
				return &FeatureTier{
					RequiresPayment: true,
					MinimumAmount:   config.GetBundlePrice(bundleID),
					BundleID:        bundleID,
					IsEnterprise:    false,
				}
			}
		}
	}

	// Feature is free/open-source
	return &FeatureTier{
		RequiresPayment: false,
		MinimumAmount:   decimal.Zero,
		IsEnterprise:    false,
	}
}
