package licensing

import (
	"context"
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestGenerateUsageReport(t *testing.T) {
	svc := GetService()
	ctx := context.Background()
	orgID := "test_org"
	features := []string{"advanced_sso", "custom_roles"}

	// Setup test license
	license, err := svc.ActivateLicense(
		ctx,
		orgID,
		features,
		[]string{"enterprise"},
		365*24*time.Hour,
		"pay_test",
		"USD",
		decimal.NewFromFloat(499.99),
	)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	// Simulate feature usage
	for i := 0; i < 5; i++ {
		svc.HasFeatureAccess(orgID, features[0])
	}
	for i := 0; i < 3; i++ {
		svc.HasFeatureAccess(orgID, features[1])
	}

	t.Run("Generate monthly report", func(t *testing.T) {
		report, err := svc.GenerateUsageReport(orgID, "monthly")
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		// Verify report basics
		if report.OrganizationID != orgID {
			t.Errorf("Expected org ID %s, got %s", orgID, report.OrganizationID)
		}
		if time.Since(report.GeneratedAt) > time.Minute {
			t.Error("Report generation time not recent")
		}
		if report.Period != "monthly" {
			t.Errorf("Expected monthly period, got %s", report.Period)
		}

		// Verify feature reports
		if len(report.Features) != len(features) {
			t.Errorf("Expected %d features, got %d", len(features), len(report.Features))
		}

		// Check SSO feature usage
		if ssoReport := report.Features["advanced_sso"]; ssoReport != nil {
			if ssoReport.TotalUsage != 5 {
				t.Errorf("Expected 5 SSO usages, got %d", ssoReport.TotalUsage)
			}
			if ssoReport.ExpirationStatus != "active" {
				t.Errorf("Expected active status, got %s", ssoReport.ExpirationStatus)
			}
		} else {
			t.Error("SSO feature report missing")
		}

		// Check roles feature usage
		if rolesReport := report.Features["custom_roles"]; rolesReport != nil {
			if rolesReport.TotalUsage != 3 {
				t.Errorf("Expected 3 roles usages, got %d", rolesReport.TotalUsage)
			}
		} else {
			t.Error("Roles feature report missing")
		}

		// Verify costs are calculated
		if report.TotalCost.IsZero() {
			t.Error("Expected non-zero total cost")
		}
	})

	t.Run("Report for expired license", func(t *testing.T) {
		// Deactivate license
		svc.DeactivateLicense(ctx, orgID)

		report, err := svc.GenerateUsageReport(orgID, "monthly")
		if err != nil {
			t.Fatalf("Failed to generate report: %v", err)
		}

		for _, feature := range report.Features {
			if feature.ExpirationStatus != "expired" {
				t.Errorf("Expected expired status, got %s", feature.ExpirationStatus)
			}
		}
	})

	t.Run("Report for nonexistent organization", func(t *testing.T) {
		_, err := svc.GenerateUsageReport("nonexistent", "monthly")
		if err != ErrLicenseNotFound {
			t.Errorf("Expected ErrLicenseNotFound, got %v", err)
		}
	})
}
