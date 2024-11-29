package licensing

import (
	"time"
	"github.com/shopspring/decimal"
)

// UsageReport represents a comprehensive report of feature usage
type UsageReport struct {
	OrganizationID string                  `json:"organization_id"`
	GeneratedAt    time.Time               `json:"generated_at"`
	Period         string                  `json:"period"` // daily, weekly, monthly
	Features       map[string]*FeatureReport `json:"features"`
	TotalCost      decimal.Decimal         `json:"total_cost"`
}

// FeatureReport contains detailed usage statistics for a feature
type FeatureReport struct {
	FeatureID        string          `json:"feature_id"`
	TotalUsage       int64           `json:"total_usage"`
	UniqueUsers      int             `json:"unique_users"`
	AverageUsage     float64         `json:"average_usage"`
	PeakUsage        int             `json:"peak_usage"`
	CostPerUse       decimal.Decimal `json:"cost_per_use"`
	TotalCost        decimal.Decimal `json:"total_cost"`
	LastUsed         time.Time       `json:"last_used"`
	UsageTrend       []int64         `json:"usage_trend"`
	ActiveSessions   int             `json:"active_sessions"`
	LicenseStatus    string          `json:"license_status"`
	ExpirationStatus string          `json:"expiration_status"`
}

// GenerateUsageReport creates a detailed usage report for an organization
func (s *Service) GenerateUsageReport(orgID string, period string) (*UsageReport, error) {
	report := &UsageReport{
		OrganizationID: orgID,
		GeneratedAt:    time.Now(),
		Period:         period,
		Features:       make(map[string]*FeatureReport),
	}

	// Get license information
	license, exists := s.licenses[orgID]
	if !exists {
		return nil, ErrLicenseNotFound
	}

	// Get usage statistics
	stats := s.usageTracker.GetUsageStats(orgID)

	// Calculate total cost and generate feature reports
	totalCost := decimal.Zero
	for _, featureID := range license.Features {
		if stat, exists := stats[featureID]; exists {
			featureReport := &FeatureReport{
				FeatureID:      featureID,
				TotalUsage:     stat.UsageCount,
				LastUsed:       stat.LastUsed,
				ActiveSessions: stat.ActiveSessions,
				LicenseStatus:  license.Status,
			}

			// Calculate cost per use based on feature pricing
			pricing, _ := s.CalculatePricing([]string{featureID}, nil, license.Currency)
			if stat.UsageCount > 0 {
				featureReport.CostPerUse = pricing.TotalAmount.Div(decimal.NewFromInt(stat.UsageCount))
			}
			featureReport.TotalCost = pricing.TotalAmount

			// Set expiration status
			if time.Now().After(license.ExpiresAt) {
				featureReport.ExpirationStatus = "expired"
			} else if time.Now().Add(30 * 24 * time.Hour).After(license.ExpiresAt) {
				featureReport.ExpirationStatus = "expiring_soon"
			} else {
				featureReport.ExpirationStatus = "active"
			}

			report.Features[featureID] = featureReport
			totalCost = totalCost.Add(featureReport.TotalCost)
		}
	}

	report.TotalCost = totalCost
	return report, nil
}
