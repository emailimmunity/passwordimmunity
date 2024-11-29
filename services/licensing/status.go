package licensing

import (
	"time"
	"github.com/shopspring/decimal"
)

// LicenseStatus represents the complete status of an organization's license
type LicenseStatus struct {
	IsActive       bool                     `json:"is_active"`
	ExpiresAt      time.Time                `json:"expires_at"`
	ActiveFeatures []string                 `json:"active_features"`
	ActiveBundles  []string                 `json:"active_bundles"`
	PaymentStatus  PaymentStatus            `json:"payment_status"`
	PricingDetails *PricingDetails          `json:"pricing_details"`
	FeatureAccess  map[string]AccessDetails `json:"feature_access"`
}

// PaymentStatus contains payment-related information
type PaymentStatus struct {
	LastPaymentID string          `json:"last_payment_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	ValidUntil    time.Time       `json:"valid_until"`
}

// AccessDetails contains access information for a specific feature
type AccessDetails struct {
	HasAccess     bool      `json:"has_access"`
	InGracePeriod bool      `json:"in_grace_period"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// GetLicenseStatus returns the complete status of an organization's license
func (s *Service) GetLicenseStatus(orgID string) *LicenseStatus {
	license, exists := s.licenses[orgID]
	if !exists {
		return &LicenseStatus{
			IsActive:      false,
			ActiveFeatures: []string{},
			ActiveBundles:  []string{},
			FeatureAccess:  make(map[string]AccessDetails),
		}
	}

	status := &LicenseStatus{
		IsActive:       license.Status == "active",
		ExpiresAt:      license.ExpiresAt,
		ActiveFeatures: license.Features,
		ActiveBundles:  license.Bundles,
		PaymentStatus: PaymentStatus{
			LastPaymentID: license.PaymentID,
			Amount:        license.Amount,
			Currency:      license.Currency,
			ValidUntil:    license.ExpiresAt,
		},
		FeatureAccess: make(map[string]AccessDetails),
	}

	// Calculate current pricing
	pricing, _ := s.CalculatePricing(license.Features, license.Bundles, license.Currency)
	status.PricingDetails = pricing

	// Get access details for each feature
	for _, featureID := range license.Features {
		accessStatus := s.GetFeatureAccessStatus(orgID, featureID)
		status.FeatureAccess[featureID] = AccessDetails{
			HasAccess:     accessStatus.HasAccess,
			InGracePeriod: accessStatus.InGracePeriod,
			ExpiresAt:     accessStatus.ExpiresAt,
		}
	}

	return status
}
