package licensing

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockFileInfo struct {
	modTime time.Time
}

func (m mockFileInfo) ModTime() time.Time {
	return m.modTime
}

func TestReportStorage(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "report_test")
	storage := NewReportStorage(tmpDir)

	t.Run("Store Report", func(t *testing.T) {
		orgID := "test_org"
		report := "Test report content"
		path, err := storage.StoreReport(orgID, report, FormatJSON)
		if err != nil {
			t.Fatalf("Failed to store report: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path")
		}
		if !strings.Contains(path, orgID) {
			t.Error("Path should contain organization ID")
		}
	})

	t.Run("List Reports", func(t *testing.T) {
		orgID := "list_org"
		reports := []string{"report1", "report2"}

		for _, r := range reports {
			_, err := storage.StoreReport(orgID, r, FormatJSON)
			if err != nil {
				t.Fatalf("Failed to store report: %v", err)
			}
		}

		listed, err := storage.ListReports(orgID)
		if err != nil {
			t.Fatalf("Failed to list reports: %v", err)
		}
		if len(listed) != len(reports) {
			t.Errorf("Expected %d reports, got %d", len(reports), len(listed))
		}
	})

	t.Run("Cleanup Old Reports", func(t *testing.T) {
		orgID := "cleanup_org"
		report := "Old report"
		path, err := storage.StoreReport(orgID, report, FormatJSON)
		if err != nil {
			t.Fatalf("Failed to store report: %v", err)
		}

		// Set file modification time to past
		oldTime := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("Failed to change file time: %v", err)
		}

		err = storage.CleanupOldReports(orgID, 24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to cleanup reports: %v", err)
		}

		// Verify report was cleaned up
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("Old report should have been deleted")
		}
	})

	t.Run("Delete Report", func(t *testing.T) {
		orgID := "delete_org"
		report := "Report to delete"
		path, err := storage.StoreReport(orgID, report, FormatJSON)
		if err != nil {
			t.Fatalf("Failed to store report: %v", err)
		}

		err = storage.DeleteReport(path)
		if err != nil {
			t.Fatalf("Failed to delete report: %v", err)
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("Report should have been deleted")
		}
	})

	// Cleanup
	os.RemoveAll(tmpDir)
}
