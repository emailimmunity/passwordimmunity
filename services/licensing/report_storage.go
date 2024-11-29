package licensing

import (
	"fmt"
	"path/filepath"
	"time"
)

// ReportStorage handles storing and retrieving generated reports
type ReportStorage struct {
	baseDir string
}

// NewReportStorage creates a new report storage instance
func NewReportStorage(baseDir string) *ReportStorage {
	return &ReportStorage{
		baseDir: baseDir,
	}
}

// StoreReport saves a report to the filesystem
func (rs *ReportStorage) StoreReport(orgID string, report string, format ExportFormat) (string, error) {
	timestamp := time.Now().Format("2006-01-02-150405")
	filename := fmt.Sprintf("report-%s.%s", timestamp, format)
	path := filepath.Join(rs.baseDir, orgID, filename)

	// Create organization directory if it doesn't exist
	orgDir := filepath.Join(rs.baseDir, orgID)
	if err := ensureDir(orgDir); err != nil {
		return "", fmt.Errorf("failed to create org directory: %w", err)
	}

	// Write report to file
	if err := writeFile(path, []byte(report)); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return path, nil
}

// GetReportPath generates the path for a report
func (rs *ReportStorage) GetReportPath(orgID string, timestamp time.Time, format ExportFormat) string {
	filename := fmt.Sprintf("report-%s.%s", timestamp.Format("2006-01-02-150405"), format)
	return filepath.Join(rs.baseDir, orgID, filename)
}

// ListReports returns all reports for an organization
func (rs *ReportStorage) ListReports(orgID string) ([]string, error) {
	orgDir := filepath.Join(rs.baseDir, orgID)
	if !dirExists(orgDir) {
		return nil, nil
	}

	pattern := filepath.Join(orgDir, "report-*.???")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	return matches, nil
}

// DeleteReport removes a report file
func (rs *ReportStorage) DeleteReport(path string) error {
	if err := deleteFile(path); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}
	return nil
}

// CleanupOldReports removes reports older than the specified duration
func (rs *ReportStorage) CleanupOldReports(orgID string, age time.Duration) error {
	reports, err := rs.ListReports(orgID)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-age)
	for _, report := range reports {
		info, err := getFileInfo(report)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			rs.DeleteReport(report)
		}
	}

	return nil
}

// Helper functions that would be implemented in a separate file
func ensureDir(path string) error {
	// Implementation would create directory if it doesn't exist
	return nil
}

func writeFile(path string, data []byte) error {
	// Implementation would write data to file
	return nil
}

func dirExists(path string) bool {
	// Implementation would check if directory exists
	return true
}

func deleteFile(path string) error {
	// Implementation would delete file
	return nil
}

func getFileInfo(path string) (FileInfo, error) {
	// Implementation would return file info
	return FileInfo{}, nil
}

// FileInfo interface for testing
type FileInfo interface {
	ModTime() time.Time
}
