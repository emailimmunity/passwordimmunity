package licensing

import (
	"testing"
	"time"
)

func TestUsageTracker(t *testing.T) {
	tracker := NewUsageTracker()
	orgID := "test_org"
	featureID := "advanced_sso"

	t.Run("Track new feature usage", func(t *testing.T) {
		tracker.TrackFeatureUsage(orgID, featureID)

		stats := tracker.GetUsageStats(orgID)
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
			if time.Since(stat.LastUsed) > time.Second {
				t.Error("LastUsed time not recently updated")
			}
		} else {
			t.Error("Feature stats not found")
		}
	})

	t.Run("Track multiple usages", func(t *testing.T) {
		tracker.TrackFeatureUsage(orgID, featureID)
		tracker.TrackFeatureUsage(orgID, featureID)

		stats := tracker.GetUsageStats(orgID)
		if stat := stats[featureID]; stat.UsageCount != 3 {
			t.Errorf("Expected usage count 3, got %d", stat.UsageCount)
		}
	})

	t.Run("End feature usage", func(t *testing.T) {
		tracker.EndFeatureUsage(orgID, featureID)

		stats := tracker.GetUsageStats(orgID)
		if stat := stats[featureID]; stat.ActiveSessions != 2 {
			t.Errorf("Expected active sessions 2, got %d", stat.ActiveSessions)
		}
	})

	t.Run("Multiple organizations", func(t *testing.T) {
		otherOrgID := "other_org"
		tracker.TrackFeatureUsage(otherOrgID, featureID)

		stats1 := tracker.GetUsageStats(orgID)
		stats2 := tracker.GetUsageStats(otherOrgID)

		if len(stats1) != 1 || len(stats2) != 1 {
			t.Error("Expected stats for both organizations")
		}

		if stats1[featureID].UsageCount == stats2[featureID].UsageCount {
			t.Error("Organizations should have independent usage counts")
		}
	})

	t.Run("Nonexistent organization", func(t *testing.T) {
		stats := tracker.GetUsageStats("nonexistent")
		if len(stats) != 0 {
			t.Error("Expected empty stats for nonexistent organization")
		}
	})

	t.Run("End usage for nonexistent feature", func(t *testing.T) {
		tracker.EndFeatureUsage(orgID, "nonexistent")
		// Should not panic or affect other stats
		stats := tracker.GetUsageStats(orgID)
		if stat := stats[featureID]; stat.ActiveSessions != 2 {
			t.Errorf("Expected active sessions to remain 2, got %d", stat.ActiveSessions)
		}
	})
}
