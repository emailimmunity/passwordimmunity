package licensing

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReportRetentionPolicy(t *testing.T) {
	// Create a temporary directory for test reports
	tmpDir, err := os.MkdirTemp("", "retention_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test service with custom storage location
	svc := &Service{
		storage: NewReportStorage(tmpDir),
	}

	// Create test files with different dates
	files := map[string]time.Time{
		"org1_daily_20240101.json":   time.Now().Add(-10 * 24 * time.Hour),
		"org1_weekly_20231201.json":  time.Now().Add(-40 * 24 * time.Hour),
		"org1_monthly_20230101.json": time.Now().Add(-400 * 24 * time.Hour),
	}

	for filename, modTime := range files {
		path := filepath.Join(tmpDir, "org1", filename)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("Failed to set file times: %v", err)
		}
	}

	// Test cleanup with custom policy
	policy := ReportRetentionPolicy{
		DailyReports:   7 * 24 * time.Hour,
		WeeklyReports:  30 * 24 * time.Hour,
		MonthlyReports: 365 * 24 * time.Hour,
	}

	if err := svc.CleanupReportsWithPolicy("org1", policy); err != nil {
		t.Fatalf("CleanupReportsWithPolicy failed: %v", err)
	}

	// Verify results
	remaining, err := svc.ListReports("org1")
	if err != nil {
		t.Fatalf("Failed to list reports: %v", err)
	}

	// Daily report should be deleted (>7 days old)
	// Weekly report should be deleted (>30 days old)
	// Monthly report should remain (<365 days old)
	expectedCount := 1
	if len(remaining) != expectedCount {
		t.Errorf("Expected %d remaining reports, got %d", expectedCount, len(remaining))
	}

	// Verify the monthly report is the one that remains
	for _, report := range remaining {
		if !strings.Contains(report, "monthly") {
			t.Errorf("Expected only monthly report to remain, found: %s", report)
		}
	}
}

func TestDefaultRetentionPolicy(t *testing.T) {
	policy := DefaultRetentionPolicy

	// Verify default retention periods
	expectedDaily := 7 * 24 * time.Hour
	if policy.DailyReports != expectedDaily {
		t.Errorf("Expected daily retention %v, got %v", expectedDaily, policy.DailyReports)
	}

	expectedWeekly := 30 * 24 * time.Hour
	if policy.WeeklyReports != expectedWeekly {
		t.Errorf("Expected weekly retention %v, got %v", expectedWeekly, policy.WeeklyReports)
	}

	expectedMonthly := 365 * 24 * time.Hour
	if policy.MonthlyReports != expectedMonthly {
		t.Errorf("Expected monthly retention %v, got %v", expectedMonthly, policy.MonthlyReports)
	}
}

func TestSetCustomRetentionPolicy(t *testing.T) {
	svc := &Service{
		storage: NewReportStorage(os.TempDir()),
	}

	t.Run("Valid Policy", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err != nil {
			t.Errorf("Expected no error for valid policy, got: %v", err)
		}
	})

	t.Run("Invalid Daily Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   12 * time.Hour, // Less than minimum
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid daily retention period")
		}
	})

	t.Run("Invalid Weekly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  5 * 24 * time.Hour, // Less than minimum
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid weekly retention period")
		}
	})

	t.Run("Invalid Monthly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 20 * 24 * time.Hour, // Less than minimum
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid monthly retention period")
		}
	})
}

func TestSchedulerRetentionPolicy(t *testing.T) {
    svc := &Service{
        storage: NewReportStorage(os.TempDir()),
    }
    scheduler := NewReportScheduler(svc)

    t.Run("Default Policy", func(t *testing.T) {
        policy := scheduler.GetRetentionPolicy("org1")
        if policy != DefaultRetentionPolicy {
            t.Error("Expected default policy for new organization")
        }
    })

    t.Run("Custom Policy", func(t *testing.T) {
        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }

        err := scheduler.SetRetentionPolicy("org1", customPolicy)
        if err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        policy := scheduler.GetRetentionPolicy("org1")
        if policy != customPolicy {
            t.Error("Expected custom policy after setting")
        }
    })

    t.Run("Remove Custom Policy", func(t *testing.T) {
        // First set a custom policy
        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }

        if err := scheduler.SetRetentionPolicy("org1", customPolicy); err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        // Remove the custom policy
        scheduler.RemoveRetentionPolicy("org1")

        // Should return to default policy
        policy := scheduler.GetRetentionPolicy("org1")
        if policy != DefaultRetentionPolicy {
            t.Error("Expected default policy after removal of custom policy")
        }
    })

    t.Run("Remove Schedule Cleanup", func(t *testing.T) {
        // Set up a schedule and custom policy
        schedule := &ReportSchedule{
            OrganizationID: "org2",
            Period:        "daily",
            Format:        JSONFormat,
            Frequency:     24 * time.Hour,
        }
        scheduler.ScheduleReport(schedule)

        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }
        if err := scheduler.SetRetentionPolicy("org2", customPolicy); err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        // Remove the schedule
        scheduler.RemoveSchedule("org2")

        // Verify both schedule and policy are removed
        if scheduler.GetSchedule("org2") != nil {
            t.Error("Schedule should be removed")
        }

        policy := scheduler.GetRetentionPolicy("org2")
        if policy != DefaultRetentionPolicy {
            t.Error("Custom policy should be removed with schedule")
        }
    })
}

func TestSetCustomRetentionPolicy(t *testing.T) {
	svc := &Service{
		storage: NewReportStorage(os.TempDir()),
	}

	t.Run("Valid Policy", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err != nil {
			t.Errorf("Expected no error for valid policy, got: %v", err)
		}
	})

	t.Run("Invalid Daily Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   12 * time.Hour, // Less than minimum
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid daily retention period")
		}
	})

	t.Run("Invalid Weekly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  5 * 24 * time.Hour, // Less than minimum
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid weekly retention period")
		}
	})

	t.Run("Invalid Monthly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 20 * 24 * time.Hour, // Less than minimum
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid monthly retention period")
		}
	})
}

func TestSchedulerRetentionPolicy(t *testing.T) {
    svc := &Service{
        storage: NewReportStorage(os.TempDir()),
    }
    scheduler := NewReportScheduler(svc)

    t.Run("Default Policy", func(t *testing.T) {
        policy := scheduler.GetRetentionPolicy("org1")
        if policy != DefaultRetentionPolicy {
            t.Error("Expected default policy for new organization")
        }
    })

    t.Run("Custom Policy", func(t *testing.T) {
        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }

        err := scheduler.SetRetentionPolicy("org1", customPolicy)
        if err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        policy := scheduler.GetRetentionPolicy("org1")
        if policy != customPolicy {
            t.Error("Expected custom policy after setting")
        }
    })

    t.Run("Remove Custom Policy", func(t *testing.T) {
        // First set a custom policy
        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }

        if err := scheduler.SetRetentionPolicy("org1", customPolicy); err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        // Remove the custom policy
        scheduler.RemoveRetentionPolicy("org1")

        // Should return to default policy
        policy := scheduler.GetRetentionPolicy("org1")
        if policy != DefaultRetentionPolicy {
            t.Error("Expected default policy after removal of custom policy")
        }
    })
}

func TestSchedulerRetentionPolicy(t *testing.T) {
    svc := &Service{
        storage: NewReportStorage(os.TempDir()),
    }
    scheduler := NewReportScheduler(svc)

    t.Run("Default Policy", func(t *testing.T) {
        policy := scheduler.GetRetentionPolicy("org1")
        if policy != DefaultRetentionPolicy {
            t.Error("Expected default policy for new organization")
        }
    })

    t.Run("Custom Policy", func(t *testing.T) {
        customPolicy := ReportRetentionPolicy{
            DailyReports:   2 * 24 * time.Hour,
            WeeklyReports:  14 * 24 * time.Hour,
            MonthlyReports: 60 * 24 * time.Hour,
        }

        err := scheduler.SetRetentionPolicy("org1", customPolicy)
        if err != nil {
            t.Fatalf("Failed to set custom policy: %v", err)
        }

        policy := scheduler.GetRetentionPolicy("org1")
        if policy != customPolicy {
            t.Error("Expected custom policy after setting")
        }
    })
}

func TestSetCustomRetentionPolicy(t *testing.T) {
	svc := &Service{
		storage: NewReportStorage(os.TempDir()),
	}

	t.Run("Valid Policy", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err != nil {
			t.Errorf("Expected no error for valid policy, got: %v", err)
		}
	})

	t.Run("Invalid Daily Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   12 * time.Hour, // Less than minimum
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid daily retention period")
		}
	})

	t.Run("Invalid Weekly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  5 * 24 * time.Hour, // Less than minimum
			MonthlyReports: 60 * 24 * time.Hour,
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid weekly retention period")
		}
	})

	t.Run("Invalid Monthly Retention", func(t *testing.T) {
		policy := ReportRetentionPolicy{
			DailyReports:   2 * 24 * time.Hour,
			WeeklyReports:  14 * 24 * time.Hour,
			MonthlyReports: 20 * 24 * time.Hour, // Less than minimum
		}
		err := svc.SetCustomRetentionPolicy("org1", policy)
		if err == nil {
			t.Error("Expected error for invalid monthly retention period")
		}
	})
}
