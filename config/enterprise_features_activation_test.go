package config

import (
	"testing"
)

func TestIsFeatureActive(t *testing.T) {
	tests := []struct {
		name           string
		featureID      string
		organizationID string
		want           bool
	}{
		{
			name:           "inactive feature",
			featureID:      "advanced_sso",
			organizationID: "org1",
			want:           false,
		},
		{
			name:           "nonexistent feature",
			featureID:      "nonexistent",
			organizationID: "org1",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFeatureActive(tt.featureID, tt.organizationID)
			if got != tt.want {
				t.Errorf("IsFeatureActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetActiveFeatures(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		want           []string
	}{
		{
			name:           "no active features",
			organizationID: "org1",
			want:          []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetActiveFeatures(tt.organizationID)
			if len(got) != len(tt.want) {
				t.Errorf("GetActiveFeatures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetActiveBundles(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		want           []string
	}{
		{
			name:           "no active bundles",
			organizationID: "org1",
			want:          []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetActiveBundles(tt.organizationID)
			if len(got) != len(tt.want) {
				t.Errorf("GetActiveBundles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequiresEnterpriseLicense(t *testing.T) {
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
			name:      "nonexistent feature",
			featureID: "nonexistent",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiresEnterpriseLicense(tt.featureID)
			if got != tt.want {
				t.Errorf("RequiresEnterpriseLicense() = %v, want %v", got, tt.want)
			}
		})
	}
}
