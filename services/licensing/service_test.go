package licensing

import (
	"context"
	"testing"
	"time"
)

func TestHasValidLicense(t *testing.T) {
	svc := GetService()

	// Setup test license
	_, err := svc.ActivateLicense(context.Background(), "org1",
		[]string{"advanced_sso"},
		[]string{"security"},
		30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	tests := []struct {
		name    string
		orgID   string
		want    bool
	}{
		{
			name:  "valid license",
			orgID: "org1",
			want:  true,
		},
		{
			name:  "no license",
			orgID: "org2",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.HasValidLicense(tt.orgID); got != tt.want {
				t.Errorf("HasValidLicense() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFeatureAccess(t *testing.T) {
	svc := GetService()

	// Setup test licenses
	_, err := svc.ActivateLicense(context.Background(), "org1",
		[]string{"advanced_sso"},
		[]string{"security"},
		30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	tests := []struct {
		name      string
		orgID     string
		featureID string
		want      bool
	}{
		{
			name:      "direct feature access",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      true,
		},
		{
			name:      "bundle feature access",
			orgID:     "org1",
			featureID: "advanced_policy",
			want:      true,
		},
		{
			name:      "no access",
			orgID:     "org1",
			featureID: "invalid_feature",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.HasFeatureAccess(tt.orgID, tt.featureID); got != tt.want {
				t.Errorf("HasFeatureAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInGracePeriod(t *testing.T) {
	svc := GetService()

	// Setup expired license in grace period
	license, err := svc.ActivateLicense(context.Background(), "org1",
		[]string{"advanced_sso"},
		[]string{"security"},
		-1*time.Hour) // Expired 1 hour ago
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}
	license.Status = "grace_period"

	tests := []struct {
		name      string
		orgID     string
		featureID string
		want      bool
	}{
		{
			name:      "in grace period",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      true,
		},
		{
			name:      "no grace period",
			orgID:     "org2",
			featureID: "advanced_sso",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.IsInGracePeriod(tt.orgID, tt.featureID); got != tt.want {
				t.Errorf("IsInGracePeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseActivationDeactivation(t *testing.T) {
	svc := GetService()
	ctx := context.Background()

	// Test activation
	license, err := svc.ActivateLicense(ctx, "org1",
		[]string{"advanced_sso"},
		[]string{"security"},
		30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	if !svc.HasValidLicense("org1") {
		t.Error("Expected valid license after activation")
	}

	// Test deactivation
	err = svc.DeactivateLicense(ctx, "org1")
	if err != nil {
		t.Fatalf("Failed to deactivate license: %v", err)
	}

	if svc.HasValidLicense("org1") {
		t.Error("Expected invalid license after deactivation")
	}
}

func TestBundleOperations(t *testing.T) {
	svc := GetService()
	ctx := context.Background()

	// Setup test license
	orgID := "test_org_bundles"
	license, err := svc.ActivateLicense(ctx, orgID, nil, nil, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	// Test bundle activation
	bundleID := "security"
	if err := svc.ActivateBundle(ctx, orgID, bundleID); err != nil {
		t.Errorf("ActivateBundle() error = %v", err)
	}

	if !svc.HasBundleAccess(orgID, bundleID) {
		t.Error("Expected bundle access after activation")
	}

	// Test GetAvailableBundles
	bundles := svc.GetAvailableBundles(orgID)
	if len(bundles) != 1 || bundles[0] != bundleID {
		t.Errorf("GetAvailableBundles() = %v, want [%s]", bundles, bundleID)
	}

	// Test bundle deactivation
	if err := svc.DeactivateBundle(ctx, orgID, bundleID); err != nil {
		t.Errorf("DeactivateBundle() error = %v", err)
	}

	if svc.HasBundleAccess(orgID, bundleID) {
		t.Error("Expected no bundle access after deactivation")
	}

	// Test error cases
	if err := svc.ActivateBundle(ctx, "invalid_org", bundleID); err != ErrLicenseNotFound {
		t.Errorf("ActivateBundle() error = %v, want %v", err, ErrLicenseNotFound)
	}

	if bundles := svc.GetAvailableBundles("invalid_org"); bundles != nil {
		t.Errorf("GetAvailableBundles() = %v, want nil", bundles)
	}
}
