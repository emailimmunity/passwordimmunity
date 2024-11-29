package licensing

import (
	"testing"
	"time"
)

func TestReportScheduler(t *testing.T) {
	svc := GetService()
	scheduler := NewReportScheduler(svc)
	defer scheduler.StopScheduler()

	t.Run("Schedule Creation", func(t *testing.T) {
		schedule := &ReportSchedule{
			OrganizationID: "test_org",
			Period:         "daily",
			Format:         FormatJSON,
			Frequency:      time.Hour,
			ExportPath:     "/tmp/reports",
		}

		scheduler.ScheduleReport(schedule)

		// Verify schedule was created
		stored := scheduler.GetSchedule("test_org")
		if stored == nil {
			t.Fatal("Schedule not stored")
		}
		if stored.OrganizationID != "test_org" {
			t.Errorf("Expected org ID test_org, got %s", stored.OrganizationID)
		}
	})

	t.Run("Schedule Removal", func(t *testing.T) {
		orgID := "remove_org"
		schedule := &ReportSchedule{
			OrganizationID: orgID,
			Period:         "weekly",
			Format:         FormatCSV,
			Frequency:      time.Hour * 24,
		}

		scheduler.ScheduleReport(schedule)
		scheduler.RemoveSchedule(orgID)

		if stored := scheduler.GetSchedule(orgID); stored != nil {
			t.Error("Schedule not removed")
		}
	})

	t.Run("Multiple Schedules", func(t *testing.T) {
		orgs := []string{"org1", "org2", "org3"}
		for _, org := range orgs {
			schedule := &ReportSchedule{
				OrganizationID: org,
				Period:         "monthly",
				Format:         FormatJSON,
				Frequency:      time.Hour * 24 * 30,
			}
			scheduler.ScheduleReport(schedule)
		}

		for _, org := range orgs {
			if stored := scheduler.GetSchedule(org); stored == nil {
				t.Errorf("Schedule for %s not found", org)
			}
		}
	})

	t.Run("Schedule Execution", func(t *testing.T) {
		schedule := &ReportSchedule{
			OrganizationID: "exec_org",
			Period:         "daily",
			Format:         FormatJSON,
			Frequency:      time.Millisecond * 100,
		}

		scheduler.ScheduleReport(schedule)
		time.Sleep(time.Millisecond * 150)

		stored := scheduler.GetSchedule("exec_org")
		if stored == nil {
			t.Fatal("Schedule not found")
		}
		if stored.LastRun.IsZero() {
			t.Error("Schedule not executed")
		}
	})
}
