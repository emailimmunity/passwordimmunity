package licensing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/shopspring/decimal"
)

// License represents an enterprise license
type License struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	Features       []string        `json:"features"`
	Bundles        []string        `json:"bundles"`
	ExpiresAt      time.Time       `json:"expires_at"`
	IssuedAt       time.Time       `json:"issued_at"`
	Status         string          `json:"status"`
	PaymentID      string          `json:"payment_id"`
	Currency       string          `json:"currency"`
	Amount         decimal.Decimal `json:"amount"`
}

// Service handles license management and verification
type Service struct {
	licenses     map[string]*License // organizationID -> License
	usageTracker *UsageTracker
	scheduler    *ReportScheduler
	storage      *ReportStorage
}

var instance *Service

// GetService returns the singleton licensing service instance
func GetService() *Service {
	if instance == nil {
		svc := &Service{
			licenses:     make(map[string]*License),
			usageTracker: NewUsageTracker(),
			storage:      NewReportStorage(filepath.Join(os.TempDir(), "passwordimmunity", "reports")),
		}
		svc.scheduler = NewReportScheduler(svc)
		instance = svc
	}
	return instance
}

// HasValidLicense checks if an organization has any valid enterprise license
func (s *Service) HasValidLicense(orgID string) bool {
	license, exists := s.licenses[orgID]
	if !exists {
		return false
	}
	return license.Status == "active" && time.Now().Before(license.ExpiresAt)
}

// HasFeatureAccess checks if an organization has access to a specific feature
func (s *Service) HasFeatureAccess(orgID string, featureID string) bool {
	status := s.GetFeatureAccessStatus(orgID, featureID)
	hasAccess := status.HasAccess && status.PaymentValid && (status.IsActive || status.InGracePeriod)
	if hasAccess {
		s.usageTracker.TrackFeatureUsage(orgID, featureID)
	}
	return hasAccess
}

// IsInGracePeriod checks if a feature is in its grace period
func (s *Service) IsInGracePeriod(orgID string, featureID string) bool {
	license, exists := s.licenses[orgID]
	if !exists {
		return false
	}

	if license.Status != "grace_period" {
		return false
	}

	gracePeriod := config.GetFeatureGracePeriod(featureID)
	if gracePeriod == 0 {
		return false
	}

	gracePeriodEnd := license.ExpiresAt.Add(time.Duration(gracePeriod) * 24 * time.Hour)
	return time.Now().Before(gracePeriodEnd)
}

// ActivateLicense activates a new license for an organization
func (s *Service) ActivateLicense(ctx context.Context, orgID string, features []string, bundles []string, duration time.Duration, paymentID string, currency string, amount decimal.Decimal) (*License, error) {
	// Validate inputs
	if orgID == "" {
		return nil, ErrInvalidOrganizationID
	}
	if paymentID == "" {
		return nil, ErrInvalidPaymentID
	}
	if len(features) == 0 && len(bundles) == 0 {
		return nil, ErrNoFeaturesOrBundles
	}
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}
	if !config.IsSupportedCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	// Calculate required pricing
	pricing, err := s.CalculatePricing(features, bundles, currency)
	if err != nil {
		return nil, err
	}

	if amount.LessThan(pricing.TotalAmount) {
		return nil, ErrInsufficientPayment
	}

	license := &License{
		ID:             generateLicenseID(),
		OrganizationID: orgID,
		Features:       features,
		Bundles:        bundles,
		IssuedAt:       time.Now(),
		ExpiresAt:      time.Now().Add(duration),
		Status:         "active",
		PaymentID:      paymentID,
		Currency:       currency,
		Amount:         amount,
	}

	// Validate payment details
	if err := ValidateLicensePayment(license); err != nil {
		return nil, err
	}

	s.licenses[orgID] = license
	return license, nil
}

// DeactivateLicense deactivates an organization's license
func (s *Service) DeactivateLicense(ctx context.Context, orgID string) error {
	if license, exists := s.licenses[orgID]; exists {
		license.Status = "inactive"
		license.ExpiresAt = time.Now()
	}
	return nil
}

// generateLicenseID generates a unique license ID
func generateLicenseID() string {
	return "lic_" + time.Now().Format("20060102150405")
}

// HasBundleAccess checks if an organization has access to a specific bundle
func (s *Service) HasBundleAccess(orgID string, bundleID string) bool {
	license, exists := s.licenses[orgID]
	if !exists {
		return false
	}

	if !s.HasValidLicense(orgID) {
		return false
	}

	for _, b := range license.Bundles {
		if b == bundleID {
			return true
		}
	}

	return false
}

// GetAvailableBundles returns all bundles available to an organization
func (s *Service) GetAvailableBundles(orgID string) []string {
	license, exists := s.licenses[orgID]
	if !exists {
		return nil
	}

	if !s.HasValidLicense(orgID) {
		return nil
	}

	return license.Bundles
}

// ActivateBundle adds a bundle to an organization's license
func (s *Service) ActivateBundle(ctx context.Context, orgID string, bundleID string) error {
	license, exists := s.licenses[orgID]
	if !exists {
		return ErrLicenseNotFound
	}

	// Check if bundle already activated
	for _, b := range license.Bundles {
		if b == bundleID {
			return nil
		}
	}

	license.Bundles = append(license.Bundles, bundleID)
	return nil
}

// DeactivateBundle removes a bundle from an organization's license
func (s *Service) DeactivateBundle(ctx context.Context, orgID string, bundleID string) error {
	license, exists := s.licenses[orgID]
	if !exists {
		return ErrLicenseNotFound
	}

	// Remove bundle from license
	newBundles := make([]string, 0)
	for _, b := range license.Bundles {
		if b != bundleID {
			newBundles = append(newBundles, b)
		}
	}
	license.Bundles = newBundles
	return nil
}

// GetFeatureUsageStats returns usage statistics for an organization
func (s *Service) GetFeatureUsageStats(orgID string) map[string]*FeatureUsageStats {
	return s.usageTracker.GetUsageStats(orgID)
}

// EndFeatureUsage marks the end of a feature usage session
func (s *Service) EndFeatureUsage(orgID string, featureID string) {
	s.usageTracker.EndFeatureUsage(orgID, featureID)
}

// ErrLicenseNotFound is returned when a license operation is attempted on a non-existent license
var ErrLicenseNotFound = errors.New("license not found")

// ScheduleReport schedules periodic report generation for an organization
func (s *Service) ScheduleReport(schedule *ReportSchedule) {
	s.scheduler.ScheduleReport(schedule)
}

// GetReportSchedule retrieves a report schedule for an organization
func (s *Service) GetReportSchedule(orgID string) *ReportSchedule {
	return s.scheduler.GetSchedule(orgID)
}

// RemoveReportSchedule removes a report schedule for an organization
func (s *Service) RemoveReportSchedule(orgID string) {
	s.scheduler.RemoveSchedule(orgID)
}

// StoreReport saves a generated report
func (s *Service) StoreReport(orgID string, report string, format ExportFormat) (string, error) {
	return s.storage.StoreReport(orgID, report, format)
}

// ListReports returns all reports for an organization
func (s *Service) ListReports(orgID string) ([]string, error) {
	return s.storage.ListReports(orgID)
}

// CleanupOldReports removes reports older than the specified duration
func (s *Service) CleanupOldReports(orgID string, age time.Duration) error {
	return s.storage.CleanupOldReports(orgID, age)
}

// SetCustomRetentionPolicy allows an organization to set their own retention policy
func (s *Service) SetCustomRetentionPolicy(orgID string, policy ReportRetentionPolicy) error {
	if policy.DailyReports < 24*time.Hour ||
	   policy.WeeklyReports < 7*24*time.Hour ||
	   policy.MonthlyReports < 30*24*time.Hour {
		return errors.New("retention periods must meet minimum requirements")
	}

	// Store the custom policy in the organization's configuration
	// For now, we'll use it directly in cleanup operations
	return s.CleanupReportsWithPolicy(orgID, policy)
}

// Shutdown cleans up resources and stops scheduled reports
func (s *Service) Shutdown() {
	if s.scheduler != nil {
		s.scheduler.StopScheduler()
	}
}

// ReportRetentionPolicy defines how long reports should be kept
type ReportRetentionPolicy struct {
	DailyReports   time.Duration
	WeeklyReports  time.Duration
	MonthlyReports time.Duration
}

// DefaultRetentionPolicy provides standard retention periods
var DefaultRetentionPolicy = ReportRetentionPolicy{
	DailyReports:   7 * 24 * time.Hour,   // Keep daily reports for 1 week
	WeeklyReports:  30 * 24 * time.Hour,  // Keep weekly reports for 1 month
	MonthlyReports: 365 * 24 * time.Hour, // Keep monthly reports for 1 year
}

// CleanupReportsWithPolicy removes old reports based on their type and retention policy
func (s *Service) CleanupReportsWithPolicy(orgID string, policy ReportRetentionPolicy) error {
	reports, err := s.ListReports(orgID)
	if err != nil {
		return err
	}

	for _, report := range reports {
		// Parse report period from filename
		if strings.Contains(report, "daily") {
			if err := s.cleanupOldReport(report, policy.DailyReports); err != nil {
				continue // Log error but continue cleanup
			}
		} else if strings.Contains(report, "weekly") {
			if err := s.cleanupOldReport(report, policy.WeeklyReports); err != nil {
				continue
			}
		} else if strings.Contains(report, "monthly") {
			if err := s.cleanupOldReport(report, policy.MonthlyReports); err != nil {
				continue
			}
		}
	}
	return nil
}

// cleanupOldReport removes a single report if it's older than the retention period
func (s *Service) cleanupOldReport(path string, retention time.Duration) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if time.Since(info.ModTime()) > retention {
		return s.storage.DeleteReport(path)
	}
	return nil
}
