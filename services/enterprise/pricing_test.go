package enterprise

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPricingManager_ConvertPrice(t *testing.T) {
	pm := NewPricingManager("USD")

	tests := []struct {
		name         string
		amount       decimal.Decimal
		fromCurrency string
		toCurrency   string
		expected     decimal.Decimal
		expectError  bool
	}{
		{
			name:         "USD to EUR",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "USD",
			toCurrency:   "EUR",
			expected:     decimal.NewFromFloat(90),
		},
		{
			name:         "EUR to USD",
			amount:       decimal.NewFromFloat(90),
			fromCurrency: "EUR",
			toCurrency:   "USD",
			expected:     decimal.NewFromFloat(100),
		},
		{
			name:         "same currency",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "USD",
			toCurrency:   "USD",
			expected:     decimal.NewFromFloat(100),
		},
		{
			name:         "unsupported currency",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "GBP",
			toCurrency:   "USD",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pm.ConvertPrice(tt.amount, tt.fromCurrency, tt.toCurrency)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, tt.expected.Equal(result),
				"Expected %s but got %s", tt.expected, result)
		})
	}
}

func TestPricingManager_ValidatePayment(t *testing.T) {
	pm := NewPricingManager("USD")

	tests := []struct {
		name       string
		amount     decimal.Decimal
		currency   string
		bundleID   string
		expectError bool
	}{
		{
			name:       "valid enterprise starter payment USD",
			amount:     decimal.NewFromFloat(199.99),
			currency:   "USD",
			bundleID:   "enterprise_starter",
			expectError: false,
		},
		{
			name:       "invalid amount",
			amount:     decimal.NewFromFloat(150),
			currency:   "USD",
			bundleID:   "enterprise_starter",
			expectError: true,
		},
		{
			name:       "invalid bundle",
			amount:     decimal.NewFromFloat(199.99),
			currency:   "USD",
			bundleID:   "nonexistent_bundle",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidatePayment(tt.amount, tt.currency, tt.bundleID)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
