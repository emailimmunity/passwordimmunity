package licensing

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ExportFormat represents supported export formats
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportUsageReport exports a usage report in the specified format
func (s *Service) ExportUsageReport(orgID string, period string, format ExportFormat) (string, error) {
	report, err := s.GenerateUsageReport(orgID, period)
	if err != nil {
		return "", err
	}

	switch format {
	case FormatJSON:
		return exportJSON(report)
	case FormatCSV:
		return exportCSV(report)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func exportJSON(report *UsageReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

func exportCSV(report *UsageReport) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	header := []string{
		"Feature ID",
		"Total Usage",
		"Active Sessions",
		"Last Used",
		"Cost Per Use",
		"Total Cost",
		"License Status",
		"Expiration Status",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write feature data
	for _, feature := range report.Features {
		row := []string{
			feature.FeatureID,
			fmt.Sprintf("%d", feature.TotalUsage),
			fmt.Sprintf("%d", feature.ActiveSessions),
			feature.LastUsed.Format(time.RFC3339),
			feature.CostPerUse.String(),
			feature.TotalCost.String(),
			feature.LicenseStatus,
			feature.ExpirationStatus,
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Write summary row
	summary := []string{
		"TOTAL",
		"",
		"",
		"",
		"",
		report.TotalCost.String(),
		"",
		"",
	}
	if err := writer.Write(summary); err != nil {
		return "", fmt.Errorf("failed to write CSV summary: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return sb.String(), nil
}
