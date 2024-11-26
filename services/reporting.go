package services

import (
	"context"
	"time"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type ReportType string

const (
	ReportTypeUsage     ReportType = "usage"
	ReportTypeSecurity  ReportType = "security"
	ReportTypeAudit     ReportType = "audit"
	ReportTypeCompliance ReportType = "compliance"
)

type ReportingService interface {
	GenerateReport(ctx context.Context, orgID uuid.UUID, reportType ReportType, startTime, endTime time.Time) (*models.Report, error)
	ScheduleReport(ctx context.Context, orgID uuid.UUID, schedule models.ReportSchedule) error
	ListReports(ctx context.Context, orgID uuid.UUID) ([]models.Report, error)
	GetReport(ctx context.Context, reportID uuid.UUID) (*models.Report, error)
	DeleteReport(ctx context.Context, reportID uuid.UUID) error
}

type reportingService struct {
	repo        repository.Repository
	metrics     MetricsService
	audit       AuditService
}

func NewReportingService(repo repository.Repository, metrics MetricsService, audit AuditService) ReportingService {
	return &reportingService{
		repo:    repo,
		metrics: metrics,
		audit:   audit,
	}
}

func (s *reportingService) GenerateReport(ctx context.Context, orgID uuid.UUID, reportType ReportType, startTime, endTime time.Time) (*models.Report, error) {
	var data interface{}
	var err error

	switch reportType {
	case ReportTypeUsage:
		data, err = s.generateUsageReport(ctx, orgID, startTime, endTime)
	case ReportTypeSecurity:
		data, err = s.generateSecurityReport(ctx, orgID, startTime, endTime)
	case ReportTypeAudit:
		data, err = s.generateAuditReport(ctx, orgID, startTime, endTime)
	case ReportTypeCompliance:
		data, err = s.generateComplianceReport(ctx, orgID, startTime, endTime)
	default:
		return nil, errors.New("unsupported report type")
	}

	if err != nil {
		return nil, err
	}

	reportData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	report := &models.Report{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           string(reportType),
		StartTime:      startTime,
		EndTime:        endTime,
		Data:           reportData,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	// Create audit log
	metadata := createBasicMetadata("report_generated", "Report generated")
	metadata["report_type"] = string(reportType)
	if err := s.createAuditLog(ctx, "report.generated", uuid.Nil, orgID, metadata); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *reportingService) ScheduleReport(ctx context.Context, orgID uuid.UUID, schedule models.ReportSchedule) error {
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("report_scheduled", "Report schedule created")
	metadata["report_type"] = schedule.ReportType
	metadata["frequency"] = schedule.Frequency
	if err := s.createAuditLog(ctx, "report.scheduled", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.CreateReportSchedule(ctx, &schedule)
}

func (s *reportingService) ListReports(ctx context.Context, orgID uuid.UUID) ([]models.Report, error) {
	return s.repo.ListReports(ctx, orgID)
}

func (s *reportingService) GetReport(ctx context.Context, reportID uuid.UUID) (*models.Report, error) {
	return s.repo.GetReport(ctx, reportID)
}

func (s *reportingService) DeleteReport(ctx context.Context, reportID uuid.UUID) error {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("report_deleted", "Report deleted")
	metadata["report_type"] = report.Type
	if err := s.createAuditLog(ctx, "report.deleted", uuid.Nil, report.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.DeleteReport(ctx, reportID)
}

func (s *reportingService) generateUsageReport(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (interface{}, error) {
	return s.metrics.GetMetrics(ctx, orgID, startTime, endTime)
}

func (s *reportingService) generateSecurityReport(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (interface{}, error) {
	// Implement security report generation
	return nil, nil
}

func (s *reportingService) generateAuditReport(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (interface{}, error) {
	return s.audit.GetAuditLogs(ctx, orgID, startTime, endTime)
}


func (s *reportingService) generateComplianceReport(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (interface{}, error) {
	// Implement compliance report generation
	return nil, nil
}

func (s *reportingService) validateSchedule(schedule models.ReportSchedule) error {
	if schedule.Frequency == "" {
		return errors.New("schedule frequency is required")
	}
	if schedule.ReportType == "" {
		return errors.New("report type is required")
	}
	return nil
}
