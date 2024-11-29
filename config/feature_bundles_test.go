package config

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetBundlePrice(t *testing.T) {
	tests := []struct {
		name          string
		bundleID      string
		currency      string
		expectedPrice decimal.Decimal
		expectError   bool
	}{
		{
			name:          "enterprise starter USD",
			bundleID:      "enterprise_starter",
			currency:      "USD",
			expectedPrice: decimal.NewFromFloat(199.99),
			expectError:   false,
		},
		{
			name:          "enterprise pro EUR",
			bundleID:      "enterprise_pro",
			currency:      "EUR",
			expectedPrice: decimal.NewFromFloat(359.99),
			expectError:   false,
		},
		{
			name:        "invalid bundle",
			bundleID:    "nonexistent_bundle",
			currency:    "USD",
			expectError: true,
		},
		{
			name:        "invalid currency",
			bundleID:    "enterprise_starter",
			currency:    "GBP",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, features, err := GetBundlePrice(tt.bundleID, tt.currency)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, tt.expectedPrice.Equal(price))
			assert.NotEmpty(t, features)
		})
	}
}

func TestGetBundle(t *testing.T) {
	tests := []struct {
		name           string
		bundleID       string
		expectError    bool
		validateBundle func(*testing.T, *Bundle)
	}{
		{
			name:        "enterprise ultimate",
			bundleID:    "enterprise_ultimate",
			expectError: false,
			validateBundle: func(t *testing.T, b *Bundle) {
				assert.Equal(t, "Enterprise Ultimate", b.Name)
				assert.Contains(t, b.Features, "multi_tenant")
				assert.Contains(t, b.Features, "advanced_sso")
				assert.NotEmpty(t, b.Description)
				assert.Contains(t, b.Price, "USD")
				assert.Contains(t, b.Price, "EUR")
			},
		},
		{
			name:        "invalid bundle",
			bundleID:    "nonexistent_bundle",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := GetBundle(tt.bundleID)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, bundle)
			if tt.validateBundle != nil {
				tt.validateBundle(t, bundle)
			}
		})
	}
}
