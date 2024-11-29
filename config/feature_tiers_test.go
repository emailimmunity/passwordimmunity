package config

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetTierFeatures(t *testing.T) {
	tests := []struct {
		name      string
		tierID    string
		wantErr   bool
		features  []string
	}{
		{
			name:    "free tier features",
			tierID:  "free",
			wantErr: false,
			features: []string{
				"basic_auth",
				"basic_roles",
				"basic_reporting",
			},
		},
		{
			name:    "enterprise tier features",
			tierID:  "enterprise",
			wantErr: false,
			features: []string{
				"basic_auth",
				"basic_roles",
				"basic_reporting",
				"custom_roles",
				"advanced_reporting",
				"advanced_sso",
				"multi_tenant",
				"advanced_audit",
				"directory_sync",
				"emergency_access",
				"custom_policies",
			},
		},
		{
			name:    "invalid tier",
			tierID:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features, err := GetTierFeatures(tt.tierID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.features, features)
		})
	}
}

func TestGetTierPrice(t *testing.T) {
	tests := []struct {
		name      string
		tierID    string
		currency  string
		yearly    bool
		expected  decimal.Decimal
		wantErr   bool
	}{
		{
			name:     "enterprise monthly USD",
			tierID:   "enterprise",
			currency: "USD",
			yearly:   false,
			expected: decimal.NewFromFloat(199.99),
		},
		{
			name:     "enterprise yearly EUR",
			tierID:   "enterprise",
			currency: "EUR",
			yearly:   true,
			expected: decimal.NewFromFloat(1799.99),
		},
		{
			name:     "free tier",
			tierID:   "free",
			currency: "USD",
			yearly:   false,
			expected: decimal.Zero,
		},
		{
			name:     "invalid tier",
			tierID:   "nonexistent",
			currency: "USD",
			yearly:   false,
			wantErr:  true,
		},
		{
			name:     "invalid currency",
			tierID:   "enterprise",
			currency: "GBP",
			yearly:   false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := GetTierPrice(tt.tierID, tt.currency, tt.yearly)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, tt.expected.Equal(price),
				"Expected %s but got %s", tt.expected, price)
		})
	}
}

func TestGetBundleFeatures(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     []string
		wantErr  bool
	}{
		{
			name:     "valid security bundle",
			bundleID: "security",
			want:     []string{"advanced_audit", "custom_policies", "emergency_access"},
			wantErr:  false,
		},
		{
			name:     "invalid bundle",
			bundleID: "nonexistent",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBundleFeatures(tt.bundleID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetBundlePrice(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		currency string
		yearly   bool
		want     decimal.Decimal
		wantErr  bool
	}{
		{
			name:     "valid monthly USD price",
			bundleID: "security",
			currency: "USD",
			yearly:   false,
			want:     decimal.NewFromFloat(49.99),
			wantErr:  false,
		},
		{
			name:     "valid yearly EUR price",
			bundleID: "security",
			currency: "EUR",
			yearly:   true,
			want:     decimal.NewFromFloat(449.99),
			wantErr:  false,
		},
		{
			name:     "invalid bundle",
			bundleID: "nonexistent",
			currency: "USD",
			yearly:   false,
			want:     decimal.Zero,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBundlePrice(tt.bundleID, tt.currency, tt.yearly)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.want.Equal(got))
			}
		})
	}
}

func TestIsEnterpriseFeature(t *testing.T) {
	tests := []struct {
		name      string
		featureID string
		want      bool
	}{
		{
			name:      "enterprise feature",
			featureID: "advanced_sso",
			want:      true,
		},
		{
			name:      "basic feature",
			featureID: "basic_auth",
			want:      false,
		},
		{
			name:      "nonexistent feature",
			featureID: "nonexistent",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnterpriseFeature(tt.featureID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsFeatureInTier(t *testing.T) {
	tests := []struct {
		name      string
		tierID    string
		featureID string
		want      bool
	}{
		{
			name:      "feature in enterprise tier",
			tierID:    "enterprise",
			featureID: "advanced_sso",
			want:      true,
		},
		{
			name:      "feature not in free tier",
			tierID:    "free",
			featureID: "advanced_sso",
			want:      false,
		},
		{
			name:      "invalid tier",
			tierID:    "nonexistent",
			featureID: "basic_auth",
			want:      false,
		},
		{
			name:      "invalid feature",
			tierID:    "enterprise",
			featureID: "nonexistent",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFeatureInTier(tt.tierID, tt.featureID)
			assert.Equal(t, tt.want, got)
		})
	}
}
