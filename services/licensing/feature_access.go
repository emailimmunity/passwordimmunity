package licensing

import (
	"github.com/emailimmunity/passwordimmunity/config"
	"time"
)

// FeatureAccessStatus represents the access status of a feature
type FeatureAccessStatus struct {
	HasAccess    bool      `json:"has_access"`
	IsActive     bool      `json:"is_active"`
	ExpiresAt    time.Time `json:"expires_at"`
	InGracePeriod bool     `json:"in_grace_period"`
	PaymentValid  bool     `json:"payment_valid"`
}

// GetFeatureAccessStatus checks if a license has valid access to a specific feature
func (s *Service) GetFeatureAccessStatus(orgID string, featureID string) *FeatureAccessStatus {
	license, exists := s.licenses[orgID]
	if !exists {
		return &FeatureAccessStatus{}
	}

	status := &FeatureAccessStatus{
		ExpiresAt: license.ExpiresAt,
		IsActive:  license.Status == "active",
	}

	// Check if license is valid
	if !s.HasValidLicense(orgID) {
		status.InGracePeriod = s.IsInGracePeriod(orgID, featureID)
		return status
	}

	// Validate payment details
	status.PaymentValid = ValidateLicensePayment(license) == nil

	// Check direct feature access
	for _, f := range license.Features {
		if f == featureID {
			status.HasAccess = status.PaymentValid
			return status
		}
	}

	// Check bundle access
	for _, bundleID := range license.Bundles {
		features := config.GetFeaturesInBundle(bundleID)
		for _, f := range features {
			if f == featureID {
				status.HasAccess = status.PaymentValid
				return status
			}
		}
	}

	return status
}
