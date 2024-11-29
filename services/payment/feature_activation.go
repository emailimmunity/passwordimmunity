package payment

import (
	"github.com/shopspring/decimal"
	"time"
)

// FeatureActivation represents the metadata for feature activation
type FeatureActivation struct {
	PaymentID   string          `json:"payment_id"`
	Features    []string        `json:"features"`
	Bundles     []string        `json:"bundles"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Currency    string         `json:"currency"`
	ValidFrom   time.Time      `json:"valid_from"`
	ValidUntil  time.Time      `json:"valid_until"`
}

// FeatureActivationRequest represents a request to activate features
type FeatureActivationRequest struct {
	Features []string `json:"features"`
	Bundles  []string `json:"bundles"`
	Currency string   `json:"currency"`
}

// ValidateFeatureActivation checks if the feature activation is valid
func ValidateFeatureActivation(activation *FeatureActivation) error {
	if activation.PaymentID == "" {
		return ErrInvalidPaymentID
	}
	if len(activation.Features) == 0 && len(activation.Bundles) == 0 {
		return ErrNoFeaturesSelected
	}
	if activation.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if activation.ValidUntil.Before(activation.ValidFrom) {
		return ErrInvalidDateRange
	}
	return nil
}
