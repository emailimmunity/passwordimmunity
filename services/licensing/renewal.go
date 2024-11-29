package licensing

import (
	"time"
)

// RenewalStatus represents the renewal status of a license
type RenewalStatus struct {
	NeedsRenewal     bool      `json:"needs_renewal"`
	DaysUntilExpiry  int       `json:"days_until_expiry"`
	ExpiresAt        time.Time `json:"expires_at"`
	InGracePeriod    bool      `json:"in_grace_period"`
	PaymentRequired  bool      `json:"payment_required"`
	RenewalAvailable bool      `json:"renewal_available"`
}

// GetRenewalStatus checks if a license needs renewal
func (s *Service) GetRenewalStatus(orgID string) *RenewalStatus {
	license, exists := s.licenses[orgID]
	if !exists {
		return &RenewalStatus{
			NeedsRenewal:     true,
			DaysUntilExpiry:  0,
			PaymentRequired:  true,
			RenewalAvailable: true,
		}
	}

	now := time.Now()
	daysUntilExpiry := int(license.ExpiresAt.Sub(now).Hours() / 24)

	status := &RenewalStatus{
		ExpiresAt:        license.ExpiresAt,
		DaysUntilExpiry:  daysUntilExpiry,
		RenewalAvailable: true,
	}

	// Check if renewal is needed (30 days or less until expiry)
	status.NeedsRenewal = daysUntilExpiry <= 30

	// Check if in grace period
	if now.After(license.ExpiresAt) {
		status.InGracePeriod = true
		status.NeedsRenewal = true
		status.PaymentRequired = true
		status.DaysUntilExpiry = 0
	}

	// Check if payment validation is current
	if err := ValidateLicensePayment(license); err != nil {
		status.PaymentRequired = true
		status.NeedsRenewal = true
	}

	return status
}
