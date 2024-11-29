package config

import (
	"testing"
)

func TestCalculateTotalPrice(t *testing.T) {
	tests := []struct {
		name     string
		features []string
		bundles  []string
		currency string
		want     float64
	}{
		{
			name:     "single bundle EUR",
			features: []string{},
			bundles:  []string{"management"},
			currency: "EUR",
			want:     129.99,
		},
		{
			name:     "bundle and feature USD",
			features: []string{"advanced_sso"},
			bundles:  []string{"management"},
			currency: "USD",
			want:     194.98, // 139.99 + 54.99
		},
		{
			name:     "multiple bundles GBP",
			features: []string{},
			bundles:  []string{"management", "compliance"},
			currency: "GBP",
			want:     179.98, // 119.99 + 59.99
		},
		{
			name:     "feature already in bundle EUR",
			features: []string{"multi_tenant"},
			bundles:  []string{"management"},
			currency: "EUR",
			want:     129.99, // No extra charge for feature in bundle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalPrice(tt.features, tt.bundles, tt.currency)
			if got != tt.want {
				t.Errorf("CalculateTotalPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFeatureAndBundleCombination(t *testing.T) {
	tests := []struct {
		name     string
		features []string
		bundles  []string
		want     bool
	}{
		{
			name:     "valid combination",
			features: []string{"advanced_sso"},
			bundles:  []string{"management"},
			want:     true,
		},
		{
			name:     "invalid feature",
			features: []string{"nonexistent"},
			bundles:  []string{"management"},
			want:     false,
		},
		{
			name:     "invalid bundle",
			features: []string{"advanced_sso"},
			bundles:  []string{"nonexistent"},
			want:     false,
		},
		{
			name:     "missing dependency",
			features: []string{"multi_tenant"},
			bundles:  []string{},
			want:     false,
		},
		{
			name:     "dependency satisfied by bundle",
			features: []string{"multi_tenant"},
			bundles:  []string{"management"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFeatureAndBundleCombination(tt.features, tt.bundles)
			if got != tt.want {
				t.Errorf("ValidateFeatureAndBundleCombination() = %v, want %v", got, tt.want)
			}
		})
	}
}
