package licensing

import (
	"context"
	"time"
	"github.com/shopspring/decimal"
)

// RenewalRequest represents a request to renew a license
type RenewalRequest struct {
	OrganizationID string          `json:"organization_id"`
	PaymentID      string          `json:"payment_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Duration       time.Duration   `json:"duration"`
}

// RenewLicense handles license renewal requests
func (s *Service) RenewLicense(ctx context.Context, req *RenewalRequest) (*License, error) {
	// Get existing license
	license, exists := s.licenses[req.OrganizationID]
	if !exists {
		return nil, ErrLicenseNotFound
	}

	// Calculate required pricing for renewal
	pricing, err := s.CalculatePricing(license.Features, license.Bundles, req.Currency)
	if err != nil {
		return nil, err
	}

	// Validate payment amount
	if req.Amount.LessThan(pricing.TotalAmount) {
		return nil, ErrInsufficientPayment
	}

	// Create renewed license
	renewedLicense := &License{
		ID:             GenerateLicenseID(),
		OrganizationID: req.OrganizationID,
		Features:       license.Features,
		Bundles:        license.Bundles,
		IssuedAt:       time.Now(),
		ExpiresAt:      time.Now().Add(req.Duration),
		Status:         "active",
		PaymentID:      req.PaymentID,
		Currency:       req.Currency,
		Amount:         req.Amount,
	}

	// Validate payment details
	if err := ValidateLicensePayment(renewedLicense); err != nil {
		return nil, err
	}

	// Update license in storage
	s.licenses[req.OrganizationID] = renewedLicense
	return renewedLicense, nil
}
