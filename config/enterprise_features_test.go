package config

import (
	"context"
	"testing"
)

func TestGetFeaturePrice(t *testing.T) {
	tests := []struct {
		name      string
		featureID string
		currency  string
		want      float64
	}{
		{
			name:      "advanced_sso EUR",
			featureID: "advanced_sso",
			currency:  "EUR",
			want:      49.99,
		},
		{
			name:      "advanced_sso USD",
			featureID: "advanced_sso",
			currency:  "USD",
			want:      54.99,
		},
		{
			name:      "advanced_sso GBP",
			featureID: "advanced_sso",
			currency:  "GBP",
			want:      44.99,
		},
		{
			name:      "invalid feature",
			featureID: "nonexistent",
			currency:  "EUR",
			want:      0,
		},
		{
			name:      "invalid currency",
			featureID: "advanced_sso",
			currency:  "JPY",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := GetFeaturePrice(tt.featureID, tt.currency)
			if price != tt.want {
				t.Errorf("GetFeaturePrice() = %v, want %v", price, tt.want)
			}
		})
	}
}

func TestGetFeatureByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    bool
		wantTier string
	}{
		{
			name:     "existing feature",
			id:       "advanced_sso",
			want:     true,
			wantTier: "enterprise",
		},
		{
			name: "non-existent feature",
			id:   "invalid_feature",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feature, exists := GetFeatureByID(tt.id)
			if exists != tt.want {
				t.Errorf("GetFeatureByID() exists = %v, want %v", exists, tt.want)
			}
			if exists && feature.Tier != tt.wantTier {
				t.Errorf("GetFeatureByID() tier = %v, want %v", feature.Tier, tt.wantTier)
			}
		})
	}
}

func TestGetBundleByID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		want          bool
		wantFeatures  int
	}{
		{
			name:         "existing bundle",
			id:           "security",
			want:         true,
			wantFeatures: 2,
		},
		{
			name: "non-existent bundle",
			id:   "invalid_bundle",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, exists := GetBundleByID(tt.id)
			if exists != tt.want {
				t.Errorf("GetBundleByID() exists = %v, want %v", exists, tt.want)
			}
			if exists && len(bundle.Features) != tt.wantFeatures {
				t.Errorf("GetBundleByID() features count = %v, want %v", len(bundle.Features), tt.wantFeatures)
			}
		})
	}
}

func TestGetBundlePrice(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		currency string
		want     float64
	}{
		{
			name:     "security bundle EUR",
			bundleID: "security",
			currency: "EUR",
			want:     89.99,
		},
		{
			name:     "security bundle USD",
			bundleID: "security",
			currency: "USD",
			want:     99.99,
		},
		{
			name:     "security bundle GBP",
			bundleID: "security",
			currency: "GBP",
			want:     79.99,
		},
		{
			name:     "invalid bundle",
			bundleID: "invalid_bundle",
			currency: "EUR",
			want:     0,
		},
		{
			name:     "invalid currency",
			bundleID: "security",
			currency: "JPY",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := GetBundlePrice(tt.bundleID, tt.currency)
			if price != tt.want {
				t.Errorf("GetBundlePrice() = %v, want %v", price, tt.want)
			}
		})
	}
}

func TestFeatureDependencies(t *testing.T) {
	feature, exists := GetFeatureByID("multi_tenant")
	if !exists {
		t.Fatal("multi_tenant feature not found")
	}

	if len(feature.Dependencies) == 0 {
		t.Error("expected dependencies for multi_tenant feature")
	}

	// Check if advanced_policy is a dependency
	hasPolicy := false
	for _, dep := range feature.Dependencies {
		if dep == "advanced_policy" {
			hasPolicy = true
			break
		}
	}
	if !hasPolicy {
		t.Error("expected advanced_policy as dependency for multi_tenant feature")
	}
}

func TestGetFeaturesInBundle(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     int
	}{
		{
			name:     "security bundle",
			bundleID: "security",
			want:     2,
		},
		{
			name:     "management bundle",
			bundleID: "management",
			want:     2,
		},
		{
			name:     "invalid bundle",
			bundleID: "invalid_bundle",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := GetFeaturesInBundle(tt.bundleID)
			if len(features) != tt.want {
				t.Errorf("GetFeaturesInBundle() count = %v, want %v", len(features), tt.want)
			}
		})
	}
}

func TestGetActiveFeatures(t *testing.T) {
	mockService := &mockFeatureActivationService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":    true,
			"advanced_audit:test_org":  true,
			"multi_tenant:test_org":    false,
			"advanced_policy:test_org": false,
		},
	}
	SetEnterpriseService(mockService)

	activeFeatures := GetActiveFeatures("test_org")
	expectedCount := 2 // advanced_sso and advanced_audit

	if len(activeFeatures) != expectedCount {
		t.Errorf("GetActiveFeatures() returned %d features, want %d", len(activeFeatures), expectedCount)
	}

	// Verify specific features
	hasFeature := func(id string) bool {
		for _, f := range activeFeatures {
			if f == id {
				return true
			}
		}
		return false
	}

	if !hasFeature("advanced_sso") {
		t.Error("GetActiveFeatures() missing advanced_sso")
	}
	if !hasFeature("advanced_audit") {
		t.Error("GetActiveFeatures() missing advanced_audit")
	}
}

func TestGetActiveBundles(t *testing.T) {
	mockService := &mockFeatureActivationService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":    true,
			"advanced_policy:test_org": true,  // Makes security bundle active
			"advanced_audit:test_org":  true,  // Part of compliance bundle
			"multi_tenant:test_org":    true,  // Part of compliance bundle
		},
	}
	SetEnterpriseService(mockService)

	activeBundles := GetActiveBundles("test_org")
	expectedBundles := []string{"security", "compliance"}

	if len(activeBundles) != len(expectedBundles) {
		t.Errorf("GetActiveBundles() returned %d bundles, want %d", len(activeBundles), len(expectedBundles))
	}

	// Verify specific bundles
	hasBundle := func(id string) bool {
		for _, b := range activeBundles {
			if b == id {
				return true
			}
		}
		return false
	}

	for _, expected := range expectedBundles {
		if !hasBundle(expected) {
			t.Errorf("GetActiveBundles() missing bundle %s", expected)
		}
	}
}

type mockFeatureActivationService struct {
	activeFeatures map[string]bool
}

func (m *mockFeatureActivationService) IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error) {
	key := featureID + ":" + organizationID
	return m.activeFeatures[key], nil
}

func TestIsFeatureActive(t *testing.T) {
	// Setup mock service
	mockService := &mockFeatureActivationService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":    true,
			"multi_tenant:test_org":    false,
			"advanced_audit:test_org":  true,
			"advanced_policy:test_org": false,
		},
	}

	// Set mock service
	SetEnterpriseService(mockService)

	tests := []struct {
		name           string
		featureID      string
		organizationID string
		want           bool
	}{
		{
			name:           "active feature",
			featureID:      "advanced_sso",
			organizationID: "test_org",
			want:           true,
		},
		{
			name:           "inactive feature",
			featureID:      "multi_tenant",
			organizationID: "test_org",
			want:           false,
		},
		{
			name:           "nonexistent feature",
			featureID:      "nonexistent",
			organizationID: "test_org",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFeatureActive(tt.featureID, tt.organizationID); got != tt.want {
				t.Errorf("IsFeatureActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
