package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
)

type HealthService interface {
	CheckHealth(ctx context.Context) (*models.HealthStatus, error)
	CheckComponent(ctx context.Context, component string) (*models.ComponentHealth, error)
	RegisterCheck(name string, check HealthCheck)
	UnregisterCheck(name string)
}

type HealthCheck func(ctx context.Context) error

type healthService struct {
	checks    map[string]HealthCheck
	mu        sync.RWMutex
	cache     CacheService
	db        repository.Repository
	redis     *redis.Client
}

func NewHealthService(
	cache CacheService,
	db repository.Repository,
	redis *redis.Client,
) HealthService {
	s := &healthService{
		checks: make(map[string]HealthCheck),
		cache:  cache,
		db:     db,
		redis:  redis,
	}

	// Register default health checks
	s.registerDefaultChecks()

	return s
}

func (s *healthService) CheckHealth(ctx context.Context) (*models.HealthStatus, error) {
	status := &models.HealthStatus{
		Status:     "healthy",
		Components: make(map[string]*models.ComponentHealth),
		Timestamp:  time.Now(),
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Run all health checks concurrently
	var wg sync.WaitGroup
	results := make(chan *models.ComponentHealth, len(s.checks))

	for name, check := range s.checks {
		wg.Add(1)
		go func(name string, check HealthCheck) {
			defer wg.Done()
			componentHealth, _ := s.CheckComponent(ctx, name)
			results <- componentHealth
		}(name, check)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		status.Components[result.Name] = result
		if result.Status != "healthy" {
			status.Status = "unhealthy"
		}
	}

	return status, nil
}

func (s *healthService) CheckComponent(ctx context.Context, component string) (*models.ComponentHealth, error) {
	s.mu.RLock()
	check, exists := s.checks[component]
	s.mu.RUnlock()

	health := &models.ComponentHealth{
		Name:      component,
		Status:    "unknown",
		Timestamp: time.Now(),
	}

	if !exists {
		health.Error = "no health check registered for component"
		return health, nil
	}

	// Run health check with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := check(ctx)
	if err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
	} else {
		health.Status = "healthy"
	}

	return health, nil
}

func (s *healthService) RegisterCheck(name string, check HealthCheck) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checks[name] = check
}

func (s *healthService) UnregisterCheck(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.checks, name)
}

func (s *healthService) registerDefaultChecks() {
	// Database health check
	s.RegisterCheck("database", func(ctx context.Context) error {
		return s.db.Ping(ctx)
	})

	// Redis health check
	s.RegisterCheck("redis", func(ctx context.Context) error {
		return s.redis.Ping(ctx).Err()
	})

	// Cache health check
	s.RegisterCheck("cache", func(ctx context.Context) error {
		key := "health_check"
		value := time.Now().String()

		if err := s.cache.Set(ctx, key, value, time.Minute); err != nil {
			return err
		}

		_, err := s.cache.Get(ctx, key)
		return err
	})
}
