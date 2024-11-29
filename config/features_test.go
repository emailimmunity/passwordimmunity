package config

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetFeaturePrice(t *testing.T) {
	tests := []struct {
		name      string
		featureID string
		currency  string
		want      decimal.Decimal
		wantErr   bool
	}{
		{
			name:      "existing feature USD",
			featureID: "advanced_sso",
			currency:  "USD",
			want:      decimal.NewFromFloat(49.99),
			wantErr:   false,
		},
		{
			name:      "non-existent feature",
			featureID: "invalid_feature",
			currency:  "USD",
			want:      decimal.Zero,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFeaturePrice(tt.featureID, tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}
}

func TestGetBundlePrice(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		currency string
		want     decimal.Decimal
		wantErr  bool
	}{
		{
			name:     "existing bundle USD",
			bundleID: "security",
			currency: "USD",
			want:     decimal.NewFromFloat(69.99),
			wantErr:  false,
		},
		{
			name:     "non-existent bundle",
			bundleID: "invalid_bundle",
			currency: "USD",
			want:     decimal.Zero,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBundlePrice(tt.bundleID, tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}
}

func TestGetFeaturesInBundle(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     []string
	}{
		{
			name:     "security bundle",
			bundleID: "security",
			want:     []string{"advanced_sso", "advanced_audit"},
		},
		{
			name:     "business bundle",
			bundleID: "business",
			want:     []string{"multi_tenant", "advanced_audit"},
		},
		{
			name:     "non-existent bundle",
			bundleID: "invalid_bundle",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFeaturesInBundle(tt.bundleID)
			if tt.want == nil && got != nil {
				t.Errorf("GetFeaturesInBundle() = %v, want nil", got)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetFeaturesInBundle() len = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, feature := range tt.want {
				if got[i] != feature {
					t.Errorf("GetFeaturesInBundle()[%d] = %v, want %v", i, got[i], feature)
				}
			}
		})
	}
}

func TestGetAllBundles(t *testing.T) {
	bundles := GetAllBundles()
	if len(bundles) != len(defaultBundles) {
		t.Errorf("GetAllBundles() returned %d bundles, want %d", len(bundles), len(defaultBundles))
	}

	// Verify specific bundle exists
	found := false
	for _, bundle := range bundles {
		if bundle.Name == "Security Bundle" {
			found = true
			expectedPrice := decimal.NewFromFloat(69.99)
			if !bundle.Price.Equal(expectedPrice) {
				t.Errorf("Security bundle price = %v, want %v", bundle.Price, expectedPrice)
			}
			break
		}
	}
	if !found {
		t.Error("Security bundle not found in GetAllBundles()")
	}
}

func TestGetAllFeatures(t *testing.T) {
	features := GetAllFeatures()
	if len(features) != len(defaultEnterpriseFeatures) {
		t.Errorf("GetAllFeatures() returned %d features, want %d", len(features), len(defaultEnterpriseFeatures))
	}

	// Verify specific feature exists
	found := false
	for _, feature := range features {
		if feature.Name == "Advanced SSO Integration" {
			found = true
			if feature.Price != 49.99 {
				t.Errorf("Advanced SSO price = %v, want 49.99", feature.Price)
			}
			break
		}
	}
	if !found {
		t.Error("Advanced SSO feature not found in GetAllFeatures()")
	}
}
