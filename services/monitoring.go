package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type MonitoringService interface {
	CheckSystemHealth(ctx context.Context) (*models.HealthStatus, error)
	RecordSystemMetric(ctx context.Context, metricName string, value float64)
	GetSystemMetrics(ctx context.Context, startTime, endTime time.Time) ([]models.SystemMetric, error)
	ConfigureAlert(ctx context.Context, alertConfig models.AlertConfig) error
	GetActiveAlerts(ctx context.Context) ([]models.Alert, error)
}

type monitoringService struct {
	repo        repository.Repository
	notification NotificationService
	mu          sync.RWMutex
	metrics     map[string][]models.SystemMetric
}

func NewMonitoringService(repo repository.Repository, notification NotificationService) MonitoringService {
	return &monitoringService{
		repo:         repo,
		notification: notification,
		metrics:      make(map[string][]models.SystemMetric),
	}
}

func (s *monitoringService) CheckSystemHealth(ctx context.Context) (*models.HealthStatus, error) {
	status := &models.HealthStatus{
		Timestamp: time.Now(),
		Status:    string(StatusHealthy),
		Components: map[string]string{
			"database":    "healthy",
			"cache":      "healthy",
			"encryption": "healthy",
		},
	}

	// Check database health
	if err := s.repo.Ping(ctx); err != nil {
		status.Status = string(StatusUnhealthy)
		status.Components["database"] = "unhealthy"
	}

	// Check cache health
	if err := s.checkCacheHealth(ctx); err != nil {
		status.Status = string(StatusDegraded)
		status.Components["cache"] = "unhealthy"
	}

	// Check encryption service health
	if err := s.checkEncryptionHealth(ctx); err != nil {
		status.Status = string(StatusUnhealthy)
		status.Components["encryption"] = "unhealthy"
	}

	return status, nil
}

func (s *monitoringService) RecordSystemMetric(ctx context.Context, metricName string, value float64) {
	metric := models.SystemMetric{
		ID:        uuid.New(),
		Name:      metricName,
		Value:     value,
		Timestamp: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics[metricName] = append(s.metrics[metricName], metric)
	s.repo.CreateSystemMetric(ctx, &metric)

	// Check alert conditions
	s.checkAlertConditions(ctx, metricName, value)
}

func (s *monitoringService) GetSystemMetrics(ctx context.Context, startTime, endTime time.Time) ([]models.SystemMetric, error) {
	return s.repo.GetSystemMetrics(ctx, startTime, endTime)
}

func (s *monitoringService) ConfigureAlert(ctx context.Context, alertConfig models.AlertConfig) error {
	if err := s.validateAlertConfig(alertConfig); err != nil {
		return err
	}

	return s.repo.CreateAlertConfig(ctx, &alertConfig)
}

func (s *monitoringService) GetActiveAlerts(ctx context.Context) ([]models.Alert, error) {
	return s.repo.GetActiveAlerts(ctx)
}

func (s *monitoringService) checkAlertConditions(ctx context.Context, metricName string, value float64) {
	configs, err := s.repo.GetAlertConfigsForMetric(ctx, metricName)
	if err != nil {
		return
	}

	for _, config := range configs {
		if s.shouldTriggerAlert(config, value) {
			alert := &models.Alert{
				ID:        uuid.New(),
				ConfigID:  config.ID,
				Value:     value,
				Message:   config.Message,
				Status:    "active",
				CreatedAt: time.Now(),
			}

			s.repo.CreateAlert(ctx, alert)

			// Send notification
			s.notification.SendSystemNotification(ctx,
				uuid.Nil,
				NotificationTypeWarning,
				config.Message,
			)
		}
	}
}

func (s *monitoringService) validateAlertConfig(config models.AlertConfig) error {
	if config.Threshold <= 0 {
		return errors.New("threshold must be positive")
	}
	if config.Message == "" {
		return errors.New("alert message is required")
	}
	return nil
}

func (s *monitoringService) checkCacheHealth(ctx context.Context) error {
	// Implement cache health check
	return nil
}

func (s *monitoringService) checkEncryptionHealth(ctx context.Context) error {
	// Implement encryption service health check
	return nil
}
