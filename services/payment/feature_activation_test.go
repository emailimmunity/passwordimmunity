package payment

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValidateFeatureActivation(t *testing.T) {
	now := time.Now()
	validActivation := &FeatureActivation{
		PaymentID:   "test_payment",
		Features:    []string{"sso"},
		TotalAmount: decimal.NewFromFloat(49.99),
		Currency:    "USD",
		ValidFrom:   now,
		ValidUntil:  now.Add(24 * time.Hour),
	}

	tests := []struct {
		name      string
		activation *FeatureActivation
		wantErr   error
	}{
		{
			name:      "valid activation",
			activation: validActivation,
			wantErr:   nil,
		},
		{
			name: "empty payment ID",
			activation: &FeatureActivation{
				Features:    []string{"sso"},
				TotalAmount: decimal.NewFromFloat(49.99),
				ValidFrom:   now,
				ValidUntil:  now.Add(24 * time.Hour),
			},
			wantErr: ErrInvalidPaymentID,
		},
		{
			name: "no features selected",
			activation: &FeatureActivation{
				PaymentID:   "test_payment",
				TotalAmount: decimal.NewFromFloat(49.99),
				ValidFrom:   now,
				ValidUntil:  now.Add(24 * time.Hour),
			},
			wantErr: ErrNoFeaturesSelected,
		},
		{
			name: "invalid amount",
			activation: &FeatureActivation{
				PaymentID:   "test_payment",
				Features:    []string{"sso"},
				TotalAmount: decimal.Zero,
				ValidFrom:   now,
				ValidUntil:  now.Add(24 * time.Hour),
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "invalid date range",
			activation: &FeatureActivation{
				PaymentID:   "test_payment",
				Features:    []string{"sso"},
				TotalAmount: decimal.NewFromFloat(49.99),
				ValidFrom:   now,
				ValidUntil:  now.Add(-24 * time.Hour),
			},
			wantErr: ErrInvalidDateRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFeatureActivation(tt.activation)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
