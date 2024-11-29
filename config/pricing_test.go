package config

import (
	"testing"
)

func TestGetFeaturePrice(t *testing.T) {
	tests := []struct {
		name      string
		featureID string
		want      float64
	}{
		{
			name:      "existing feature",
			featureID: "advanced_sso",
			want:      99.99,
		},
		{
			name:      "non-existent feature",
			featureID: "invalid_feature",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFeaturePrice(tt.featureID); got != tt.want {
				t.Errorf("GetFeaturePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBundlePrice(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     float64
	}{
		{
			name:     "existing bundle",
			bundleID: "security",
			want:     299.99,
		},
		{
			name:     "non-existent bundle",
			bundleID: "invalid_bundle",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBundlePrice(tt.bundleID); got != tt.want {
				t.Errorf("GetBundlePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFeatureGracePeriod(t *testing.T) {
	if got := GetFeatureGracePeriod("any_feature"); got != 14 {
		t.Errorf("GetFeatureGracePeriod() = %v, want 14", got)
	}
}

func TestGetBaseURL(t *testing.T) {
	if got := GetBaseURL(); got != "https://api.passwordimmunity.com" {
		t.Errorf("GetBaseURL() = %v, want https://api.passwordimmunity.com", got)
	}
}
