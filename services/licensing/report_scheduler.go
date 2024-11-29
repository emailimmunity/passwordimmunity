package licensing

import (
	"context"
	"sync"
	"time"
)

// ReportSchedule defines when reports should be generated
type ReportSchedule struct {
	OrganizationID string
	Period         string
	Format         ExportFormat
	Frequency      time.Duration
	LastRun        time.Time
	ExportPath     string
}

// ReportScheduler manages scheduled report generation
type ReportScheduler struct {
	mu        sync.RWMutex
	schedules map[string]*ReportSchedule
	service   *Service
	ctx       context.Context
	cancel    context.CancelFunc
	cleanupInterval time.Duration
	retentionPolicies map[string]ReportRetentionPolicy
}

// NewReportScheduler creates a new scheduler instance
func NewReportScheduler(service *Service) *ReportScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	scheduler := &ReportScheduler{
		schedules: make(map[string]*ReportSchedule),
		retentionPolicies: make(map[string]ReportRetentionPolicy),
		service:   service,
		ctx:       ctx,
		cancel:    cancel,
		cleanupInterval: 24 * time.Hour, // Clean up daily
	}
	go scheduler.runCleanup()
	return scheduler
}

// SetRetentionPolicy sets a custom retention policy for an organization
func (rs *ReportScheduler) SetRetentionPolicy(orgID string, policy ReportRetentionPolicy) error {
	if err := rs.service.SetCustomRetentionPolicy(orgID, policy); err != nil {
		return err
	}

	rs.mu.Lock()
	rs.retentionPolicies[orgID] = policy
	rs.mu.Unlock()
	return nil
}

// GetRetentionPolicy returns the current retention policy for an organization
func (rs *ReportScheduler) GetRetentionPolicy(orgID string) ReportRetentionPolicy {
    rs.mu.RLock()
    defer rs.mu.RUnlock()

    policy, exists := rs.retentionPolicies[orgID]
    if !exists {
        return DefaultRetentionPolicy
    }
    return policy
}

// RemoveRetentionPolicy removes a custom retention policy for an organization
func (rs *ReportScheduler) RemoveRetentionPolicy(orgID string) {
    rs.mu.Lock()
    delete(rs.retentionPolicies, orgID)
    rs.mu.Unlock()
}

// GetRetentionPolicy returns the current retention policy for an organization
func (rs *ReportScheduler) GetRetentionPolicy(orgID string) ReportRetentionPolicy {
    rs.mu.RLock()
    defer rs.mu.RUnlock()

    policy, exists := rs.retentionPolicies[orgID]
    if !exists {
        return DefaultRetentionPolicy
    }
    return policy
}

// ScheduleReport adds a new report schedule
func (rs *ReportScheduler) ScheduleReport(schedule *ReportSchedule) {
	rs.mu.Lock()
	rs.schedules[schedule.OrganizationID] = schedule
	rs.mu.Unlock()

	go rs.runSchedule(schedule)
}

// runSchedule handles the periodic report generation
func (rs *ReportScheduler) runSchedule(schedule *ReportSchedule) {
	ticker := time.NewTicker(schedule.Frequency)
	defer ticker.Stop()

	for {
		select {
		case <-rs.ctx.Done():
			return
		case <-ticker.C:
			if err := rs.generateScheduledReport(schedule); err != nil {
				// Log error or notify monitoring system
				continue
			}
			schedule.LastRun = time.Now()
		}
	}
}

// RemoveSchedule removes a report schedule and its associated retention policy
func (rs *ReportScheduler) RemoveSchedule(orgID string) {
    rs.mu.Lock()
    delete(rs.schedules, orgID)
    delete(rs.retentionPolicies, orgID)
    rs.mu.Unlock()
}

// generateScheduledReport creates and exports a report
func (rs *ReportScheduler) generateScheduledReport(schedule *ReportSchedule) error {
	output, err := rs.service.ExportUsageReport(
		schedule.OrganizationID,
		schedule.Period,
		schedule.Format,
	)
	if err != nil {
		return err
	}

	// Store the report using the storage system
	path, err := rs.service.StoreReport(schedule.OrganizationID, output, schedule.Format)
	if err != nil {
		return err
	}

	schedule.LastRun = time.Now()
	schedule.ExportPath = path
	return nil
}

// StopScheduler stops all scheduled reports
func (rs *ReportScheduler) StopScheduler() {
	rs.cancel()
}

// RemoveSchedule removes a report schedule
func (rs *ReportScheduler) RemoveSchedule(orgID string) {
	rs.mu.Lock()
	delete(rs.schedules, orgID)
	rs.mu.Unlock()
}

// GetSchedule retrieves a report schedule
func (rs *ReportScheduler) GetSchedule(orgID string) *ReportSchedule {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.schedules[orgID]
}

// runCleanup periodically removes old reports
func (rs *ReportScheduler) runCleanup() {
	ticker := time.NewTicker(rs.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rs.ctx.Done():
			return
		case <-ticker.C:
			rs.mu.RLock()
			for orgID := range rs.schedules {
				// Use custom policy if available, otherwise use default
				policy, exists := rs.retentionPolicies[orgID]
				if !exists {
					policy = DefaultRetentionPolicy
				}
				if err := rs.service.CleanupReportsWithPolicy(orgID, policy); err != nil {
					// Log error but continue cleanup
					continue
				}
			}
			rs.mu.RUnlock()
		}
	}
}
