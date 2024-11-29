package licensing

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestExportUsageReport(t *testing.T) {
	svc := GetService()
	ctx := context.Background()
	orgID := "test_org"
	features := []string{"advanced_sso", "custom_roles"}

	// Setup test license and generate usage
	_, err := svc.ActivateLicense(
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
	svc.HasFeatureAccess(orgID, features[0])
	svc.HasFeatureAccess(orgID, features[1])

	t.Run("Export JSON format", func(t *testing.T) {
		output, err := svc.ExportUsageReport(orgID, "monthly", FormatJSON)
		if err != nil {
			t.Fatalf("Failed to export JSON report: %v", err)
		}

		// Verify JSON structure
		var report UsageReport
		if err := json.Unmarshal([]byte(output), &report); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		if report.OrganizationID != orgID {
			t.Errorf("Expected org ID %s, got %s", orgID, report.OrganizationID)
		}

		if len(report.Features) != len(features) {
			t.Errorf("Expected %d features, got %d", len(features), len(report.Features))
		}
	})

	t.Run("Export CSV format", func(t *testing.T) {
		output, err := svc.ExportUsageReport(orgID, "monthly", FormatCSV)
		if err != nil {
			t.Fatalf("Failed to export CSV report: %v", err)
		}

		// Verify CSV structure
		lines := strings.Split(output, "\n")
		if len(lines) < 4 { // Header + 2 features + summary + empty line
			t.Errorf("Expected at least 4 lines in CSV, got %d", len(lines))
		}

		// Check header
		header := strings.Split(lines[0], ",")
		if len(header) != 8 {
			t.Errorf("Expected 8 columns in CSV header, got %d", len(header))
		}

		// Check feature rows
		for _, feature := range features {
			found := false
			for _, line := range lines[1:] {
				if strings.Contains(line, feature) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Feature %s not found in CSV output", feature)
			}
		}

		// Check summary row
		if !strings.Contains(lines[len(lines)-2], "TOTAL") {
			t.Error("Summary row not found in CSV output")
		}
	})

	t.Run("Invalid export format", func(t *testing.T) {
		_, err := svc.ExportUsageReport(orgID, "monthly", "invalid")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("Export for nonexistent organization", func(t *testing.T) {
		_, err := svc.ExportUsageReport("nonexistent", "monthly", FormatJSON)
		if err != ErrLicenseNotFound {
			t.Errorf("Expected ErrLicenseNotFound, got %v", err)
		}
	})
}
