package licensing

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestGetRenewalNotifications(t *testing.T) {
	svc := GetService()
	now := time.Now()

	// Setup test licenses
	testLicenses := map[string]*License{
		"org_expired": {
			OrganizationID: "org_expired",
			ExpiresAt:      now.Add(-24 * time.Hour),
			Status:         "expired",
			Amount:         decimal.NewFromFloat(100),
		},
		"org_critical": {
			OrganizationID: "org_critical",
			ExpiresAt:      now.Add(5 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(100),
		},
		"org_medium": {
			OrganizationID: "org_medium",
			ExpiresAt:      now.Add(10 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(100),
		},
		"org_low": {
			OrganizationID: "org_low",
			ExpiresAt:      now.Add(25 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(100),
		},
		"org_valid": {
			OrganizationID: "org_valid",
			ExpiresAt:      now.Add(60 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(100),
		},
	}

	for _, license := range testLicenses {
		svc.licenses[license.OrganizationID] = license
	}

	notifications := svc.GetRenewalNotifications()

	// Map to store found priorities
	found := make(map[string]bool)

	for _, notification := range notifications {
		switch notification.Priority {
		case "critical":
			if notification.OrganizationID != "org_expired" {
				t.Errorf("Expected org_expired to have critical priority, got %s", notification.OrganizationID)
			}
			found["critical"] = true
		case "high":
			if notification.OrganizationID != "org_critical" {
				t.Errorf("Expected org_critical to have high priority, got %s", notification.OrganizationID)
			}
			found["high"] = true
		case "medium":
			if notification.OrganizationID != "org_medium" {
				t.Errorf("Expected org_medium to have medium priority, got %s", notification.OrganizationID)
			}
			found["medium"] = true
		case "low":
			if notification.OrganizationID != "org_low" {
				t.Errorf("Expected org_low to have low priority, got %s", notification.OrganizationID)
			}
			found["low"] = true
		}
	}

	// Verify all expected priorities were found
	expectedPriorities := []string{"critical", "high", "medium", "low"}
	for _, priority := range expectedPriorities {
		if !found[priority] {
			t.Errorf("Expected to find notification with priority %s", priority)
		}
	}

	// Verify no notification for valid license
	for _, notification := range notifications {
		if notification.OrganizationID == "org_valid" {
			t.Error("Should not have notification for valid license")
		}
	}
}
