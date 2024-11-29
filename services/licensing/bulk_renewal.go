package licensing

import (
	"context"
	"time"
	"github.com/shopspring/decimal"
)

// BulkRenewalRequest represents a request to renew multiple licenses
type BulkRenewalRequest struct {
	OrganizationIDs []string        `json:"organization_ids"`
	PaymentID       string          `json:"payment_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	Duration        time.Duration   `json:"duration"`
}

// BulkRenewalResult represents the result of a bulk renewal operation
type BulkRenewalResult struct {
	SuccessfulRenewals map[string]*License `json:"successful_renewals"`
	FailedRenewals     map[string]error    `json:"failed_renewals"`
	TotalAmount        decimal.Decimal     `json:"total_amount"`
}

// RenewLicensesBulk handles bulk license renewal requests
func (s *Service) RenewLicensesBulk(ctx context.Context, req *BulkRenewalRequest) (*BulkRenewalResult, error) {
	result := &BulkRenewalResult{
		SuccessfulRenewals: make(map[string]*License),
		FailedRenewals:     make(map[string]error),
		TotalAmount:        decimal.Zero,
	}

	// Calculate total required amount
	for _, orgID := range req.OrganizationIDs {
		license, exists := s.licenses[orgID]
		if !exists {
			result.FailedRenewals[orgID] = ErrLicenseNotFound
			continue
		}

		pricing, err := s.CalculatePricing(license.Features, license.Bundles, req.Currency)
		if err != nil {
			result.FailedRenewals[orgID] = err
			continue
		}

		result.TotalAmount = result.TotalAmount.Add(pricing.TotalAmount)
	}

	// Validate total payment amount
	if req.Amount.LessThan(result.TotalAmount) {
		return nil, ErrInsufficientPayment
	}

	// Process renewals
	for _, orgID := range req.OrganizationIDs {
		if _, failed := result.FailedRenewals[orgID]; failed {
			continue
		}

		renewalReq := &RenewalRequest{
			OrganizationID: orgID,
			PaymentID:      req.PaymentID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Duration:       req.Duration,
		}

		license, err := s.RenewLicense(ctx, renewalReq)
		if err != nil {
			result.FailedRenewals[orgID] = err
		} else {
			result.SuccessfulRenewals[orgID] = license
		}
	}

	return result, nil
}
