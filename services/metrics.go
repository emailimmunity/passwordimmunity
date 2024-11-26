package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsService interface {
	RecordLogin(ctx context.Context, userID uuid.UUID, success bool)
	RecordAPIRequest(ctx context.Context, endpoint string, statusCode int, duration time.Duration)
	RecordVaultAccess(ctx context.Context, userID uuid.UUID, itemType string)
	RecordAuthFailure(ctx context.Context, reason string)
	GetMetrics(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (*models.MetricsReport, error)
}

type metricsService struct {
	repo repository.Repository
	mu   sync.RWMutex

	// Prometheus metrics
	loginAttempts   *prometheus.CounterVec
	apiRequests     *prometheus.HistogramVec
	vaultAccesses   *prometheus.CounterVec
	authFailures    *prometheus.CounterVec
}

func NewMetricsService(repo repository.Repository) MetricsService {
	s := &metricsService{
		repo: repo,
		loginAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "login_attempts_total",
				Help: "Total number of login attempts",
			},
			[]string{"success"},
		),
		apiRequests: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "api_request_duration_seconds",
				Help: "API request duration in seconds",
			},
			[]string{"endpoint", "status"},
		),
		vaultAccesses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vault_accesses_total",
				Help: "Total number of vault accesses",
			},
			[]string{"item_type"},
		),
		authFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_failures_total",
				Help: "Total number of authentication failures",
			},
			[]string{"reason"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(s.loginAttempts)
	prometheus.MustRegister(s.apiRequests)
	prometheus.MustRegister(s.vaultAccesses)
	prometheus.MustRegister(s.authFailures)

	return s
}

func (s *metricsService) RecordLogin(ctx context.Context, userID uuid.UUID, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	s.loginAttempts.WithLabelValues(successStr).Inc()

	// Store in database for historical analysis
	metric := &models.Metric{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "login",
		Success:   success,
		Timestamp: time.Now(),
	}
	s.repo.CreateMetric(ctx, metric)
}

func (s *metricsService) RecordAPIRequest(ctx context.Context, endpoint string, statusCode int, duration time.Duration) {
	s.apiRequests.WithLabelValues(
		endpoint,
		string(statusCode),
	).Observe(duration.Seconds())

	// Store in database for historical analysis
	metric := &models.Metric{
		ID:        uuid.New(),
		Type:      "api_request",
		Metadata:  map[string]interface{}{
			"endpoint":    endpoint,
			"status_code": statusCode,
			"duration":    duration.Seconds(),
		},
		Timestamp: time.Now(),
	}
	s.repo.CreateMetric(ctx, metric)
}

func (s *metricsService) RecordVaultAccess(ctx context.Context, userID uuid.UUID, itemType string) {
	s.vaultAccesses.WithLabelValues(itemType).Inc()

	// Store in database for historical analysis
	metric := &models.Metric{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "vault_access",
		Metadata:  map[string]interface{}{
			"item_type": itemType,
		},
		Timestamp: time.Now(),
	}
	s.repo.CreateMetric(ctx, metric)
}

func (s *metricsService) RecordAuthFailure(ctx context.Context, reason string) {
	s.authFailures.WithLabelValues(reason).Inc()

	// Store in database for historical analysis
	metric := &models.Metric{
		ID:        uuid.New(),
		Type:      "auth_failure",
		Metadata:  map[string]interface{}{
			"reason": reason,
		},
		Timestamp: time.Now(),
	}
	s.repo.CreateMetric(ctx, metric)
}

func (s *metricsService) GetMetrics(ctx context.Context, orgID uuid.UUID, startTime, endTime time.Time) (*models.MetricsReport, error) {
	metrics, err := s.repo.GetMetrics(ctx, orgID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Aggregate metrics into a report
	report := &models.MetricsReport{
		OrganizationID: orgID,
		StartTime:      startTime,
		EndTime:        endTime,
		LoginAttempts:  make(map[bool]int),
		APIRequests:    make(map[string]int),
		VaultAccesses:  make(map[string]int),
		AuthFailures:   make(map[string]int),
	}

	for _, metric := range metrics {
		switch metric.Type {
		case "login":
			report.LoginAttempts[metric.Success]++
		case "api_request":
			if endpoint, ok := metric.Metadata["endpoint"].(string); ok {
				report.APIRequests[endpoint]++
			}
		case "vault_access":
			if itemType, ok := metric.Metadata["item_type"].(string); ok {
				report.VaultAccesses[itemType]++
			}
		case "auth_failure":
			if reason, ok := metric.Metadata["reason"].(string); ok {
				report.AuthFailures[reason]++
			}
		}
	}

	return report, nil
}
