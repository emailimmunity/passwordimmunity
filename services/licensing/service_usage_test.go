package licensing

import (
	"context"
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestServiceUsageTracking(t *testing.T) {
	svc := GetService()
	ctx := context.Background()
	orgID := "test_org"
	featureID := "advanced_sso"

	// Setup test license
	license, err := svc.ActivateLicense(
		ctx,
		orgID,
		[]string{featureID},
		[]string{"business"},
		365*24*time.Hour,
		"pay_test",
		"USD",
		decimal.NewFromFloat(199.99),
	)
	if err != nil {
		t.Fatalf("Failed to activate license: %v", err)
	}

	t.Run("Usage tracking on feature access", func(t *testing.T) {
		// First access should create usage stats
		if !svc.HasFeatureAccess(orgID, featureID) {
			t.Fatal("Expected feature access to be granted")
		}

		stats := svc.GetFeatureUsageStats(orgID)
		if len(stats) != 1 {
			t.Errorf("Expected 1 feature stat, got %d", len(stats))
		}

		if stat, exists := stats[featureID]; exists {
			if stat.UsageCount != 1 {
				t.Errorf("Expected usage count 1, got %d", stat.UsageCount)
			}
			if stat.ActiveSessions != 1 {
				t.Errorf("Expected active sessions 1, got %d", stat.ActiveSessions)
			}
		} else {
			t.Error("Feature stats not found")
		}

		// Multiple accesses should increment counters
		svc.HasFeatureAccess(orgID, featureID)
		stats = svc.GetFeatureUsageStats(orgID)
		if stat := stats[featureID]; stat.UsageCount != 2 {
			t.Errorf("Expected usage count 2, got %d", stat.UsageCount)
		}
	})

	t.Run("End feature usage", func(t *testing.T) {
		svc.EndFeatureUsage(orgID, featureID)

		stats := svc.GetFeatureUsageStats(orgID)
		if stat := stats[featureID]; stat.ActiveSessions != 1 {
			t.Errorf("Expected active sessions 1, got %d", stat.ActiveSessions)
		}
	})

	t.Run("No tracking for unauthorized access", func(t *testing.T) {
		unauthorizedFeature := "unauthorized_feature"
		svc.HasFeatureAccess(orgID, unauthorizedFeature)

		stats := svc.GetFeatureUsageStats(orgID)
		if _, exists := stats[unauthorizedFeature]; exists {
			t.Error("Should not track usage for unauthorized feature")
		}
	})

	t.Run("No tracking for expired license", func(t *testing.T) {
		// Deactivate license
		svc.DeactivateLicense(ctx, orgID)

		// Try to access feature
		svc.HasFeatureAccess(orgID, featureID)

		// Get current stats
		stats := svc.GetFeatureUsageStats(orgID)
		if stat := stats[featureID]; stat.UsageCount > 2 {
			t.Error("Should not track usage for expired license")
		}
	})
}
