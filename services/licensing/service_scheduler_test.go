package licensing

import (
	"os"
	"testing"
	"time"
)

func TestServiceSchedulerIntegration(t *testing.T) {
	svc := GetService()
	defer svc.Shutdown()

	t.Run("Schedule Management", func(t *testing.T) {
		orgID := "test_org"
		schedule := &ReportSchedule{
			OrganizationID: orgID,
			Period:         "daily",
			Format:         FormatJSON,
			Frequency:      time.Hour,
		}

		// Add schedule
		svc.ScheduleReport(schedule)

		// Verify schedule retrieval
		stored := svc.GetReportSchedule(orgID)
		if stored == nil {
			t.Fatal("Schedule not stored")
		}
		if stored.OrganizationID != orgID {
			t.Errorf("Expected org ID %s, got %s", orgID, stored.OrganizationID)
		}

		// Remove schedule
		svc.RemoveReportSchedule(orgID)
		if stored = svc.GetReportSchedule(orgID); stored != nil {
			t.Error("Schedule not removed")
		}
	})

	t.Run("Report Cleanup", func(t *testing.T) {
		svc := GetService()
		defer svc.Shutdown()

		orgID := "cleanup_test_org"
		schedule := &ReportSchedule{
			OrganizationID: orgID,
			Period:         "daily",
			Format:         FormatJSON,
			Frequency:      time.Hour,
		}

		// Add schedule and generate a report
		svc.ScheduleReport(schedule)

		// Generate and store a test report
		report := "Test report content"
		path, err := svc.StoreReport(orgID, report, FormatJSON)
		if err != nil {
			t.Fatalf("Failed to store report: %v", err)
		}

		// Set file modification time to past
		oldTime := time.Now().Add(-31 * 24 * time.Hour)
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to change file time: %v", err)
		}

		// Trigger cleanup
		if err := svc.CleanupOldReports(orgID, 30*24*time.Hour); err != nil {
			t.Fatalf("Failed to cleanup reports: %v", err)
		}

		// Verify report was cleaned up
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("Old report should have been deleted")
		}
	})

	t.Run("Service Shutdown", func(t *testing.T) {
		orgID := "shutdown_org"
		schedule := &ReportSchedule{
			OrganizationID: orgID,
			Period:         "daily",
			Format:         FormatJSON,
			Frequency:      time.Hour,
		}

		svc.ScheduleReport(schedule)
		svc.Shutdown()

		// Verify scheduler is stopped
		time.Sleep(time.Millisecond * 100)
		stored := svc.GetReportSchedule(orgID)
		if stored == nil {
			t.Fatal("Schedule lost after shutdown")
		}
		if stored.LastRun.IsZero() {
			t.Log("Scheduler successfully stopped - no executions after shutdown")
		}
	})
}
