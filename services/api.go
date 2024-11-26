package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type APIService interface {
	CreateAPIKey(ctx context.Context, orgID uuid.UUID, name string, permissions []string) (*models.APIKey, error)
	RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error
	ValidateAPIKey(ctx context.Context, key string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error)
	CheckRateLimit(ctx context.Context, keyID uuid.UUID) error
}

type apiService struct {
	repo        repository.Repository
	audit       AuditService
	licensing   LicensingService
	rateLimiters map[uuid.UUID]*rate.Limiter
	mu           sync.RWMutex
}

func NewAPIService(repo repository.Repository, audit AuditService, licensing LicensingService) APIService {
	return &apiService{
		repo:         repo,
		audit:        audit,
		licensing:    licensing,
		rateLimiters: make(map[uuid.UUID]*rate.Limiter),
	}
}

func (s *apiService) CreateAPIKey(ctx context.Context, orgID uuid.UUID, name string, permissions []string) (*models.APIKey, error) {
	// Check if organization has API access
	hasAccess, err := s.licensing.CheckFeatureAccess(ctx, orgID, "api_access")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("api access not available in current license")
	}

	key := &models.APIKey{
		ID:            uuid.New(),
		OrganizationID: orgID,
		Name:          name,
		Key:           generateAPIKey(),
		Permissions:   permissions,
		CreatedAt:     time.Now(),
		LastUsedAt:    nil,
	}

	if err := s.repo.CreateAPIKey(ctx, key); err != nil {
		return nil, err
	}

	// Create rate limiter for the new key
	s.mu.Lock()
	s.rateLimiters[key.ID] = rate.NewLimiter(rate.Limit(100), 100) // 100 requests per second burst
	s.mu.Unlock()

	// Create audit log
	metadata := createBasicMetadata("api_key_created", "API key created")
	metadata["key_name"] = name
	if err := s.createAuditLog(ctx, "api.key.created", uuid.Nil, orgID, metadata); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *apiService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	key, err := s.repo.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	// Remove rate limiter
	s.mu.Lock()
	delete(s.rateLimiters, keyID)
	s.mu.Unlock()

	// Create audit log
	metadata := createBasicMetadata("api_key_revoked", "API key revoked")
	metadata["key_name"] = key.Name
	if err := s.createAuditLog(ctx, "api.key.revoked", uuid.Nil, key.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.DeleteAPIKey(ctx, keyID)
}

func (s *apiService) ValidateAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	apiKey, err := s.repo.GetAPIKeyByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp
	apiKey.LastUsedAt = &time.Time{}
	*apiKey.LastUsedAt = time.Now()
	if err := s.repo.UpdateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *apiService) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.ListAPIKeys(ctx, orgID)
}

func (s *apiService) CheckRateLimit(ctx context.Context, keyID uuid.UUID) error {
	s.mu.RLock()
	limiter, exists := s.rateLimiters[keyID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("rate limiter not found")
	}

	if !limiter.Allow() {
		return errors.New("rate limit exceeded")
	}

	return nil
}

func generateAPIKey() string {
	// Implement secure API key generation
	return uuid.New().String()
}
