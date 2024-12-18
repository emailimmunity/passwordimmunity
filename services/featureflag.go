package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type FeatureFlagService interface {
	IsEnabled(ctx context.Context, flag string, orgID uuid.UUID) (bool, error)
	SetFlag(ctx context.Context, flag string, enabled bool, options models.FeatureFlagOptions) error
	DeleteFlag(ctx context.Context, flag string) error
	ListFlags(ctx context.Context) ([]models.FeatureFlag, error)
	GetFlag(ctx context.Context, flag string) (*models.FeatureFlag, error)
}

type featureFlagService struct {
	repo      repository.Repository
	cache     CacheService
	licensing LicensingService
	mu        sync.RWMutex
}

func NewFeatureFlagService(
	repo repository.Repository,
	cache CacheService,
	licensing LicensingService,
) FeatureFlagService {
	return &featureFlagService{
		repo:      repo,
		cache:     cache,
		licensing: licensing,
	}
}

func (s *featureFlagService) IsEnabled(ctx context.Context, flag string, orgID uuid.UUID) (bool, error) {
	// Check cache first
	cacheKey := getFeatureFlagCacheKey(flag, orgID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(bool), nil
	}

	// Get flag configuration
	featureFlag, err := s.GetFlag(ctx, flag)
	if err != nil {
		return false, err
	}
	if featureFlag == nil {
		return false, nil
	}

	// Check if feature requires license
	if featureFlag.RequiresLicense {
		enabled, err := s.licensing.IsFeatureEnabled(ctx, orgID, flag)
		if err != nil || !enabled {
			return false, err
		}
	}

	// Apply organization rules
	enabled := s.evaluateFlag(featureFlag, orgID)

	// Cache result
	s.cache.Set(ctx, cacheKey, enabled, time.Hour)

	return enabled, nil
}

func (s *featureFlagService) SetFlag(ctx context.Context, flag string, enabled bool, options models.FeatureFlagOptions) error {
	featureFlag := &models.FeatureFlag{
		Name:            flag,
		Enabled:         enabled,
		Options:         options,
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := s.repo.SetFeatureFlag(ctx, featureFlag); err != nil {
		return err
	}

	// Clear cache
	s.clearFlagCache(flag)

	return nil
}

func (s *featureFlagService) DeleteFlag(ctx context.Context, flag string) error {
	if err := s.repo.DeleteFeatureFlag(ctx, flag); err != nil {
		return err
	}

	// Clear cache
	s.clearFlagCache(flag)

	return nil
}

func (s *featureFlagService) ListFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	return s.repo.ListFeatureFlags(ctx)
}

func (s *featureFlagService) GetFlag(ctx context.Context, flag string) (*models.FeatureFlag, error) {
	return s.repo.GetFeatureFlag(ctx, flag)
}

func (s *featureFlagService) evaluateFlag(flag *models.FeatureFlag, orgID uuid.UUID) bool {
	if !flag.Enabled {
		return false
	}

	// Check organization allowlist
	if len(flag.Options.AllowedOrgs) > 0 {
		allowed := false
		for _, allowedOrg := range flag.Options.AllowedOrgs {
			if allowedOrg == orgID {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Check percentage rollout
	if flag.Options.RolloutPercentage > 0 {
		hash := fnv.New32a()
		hash.Write([]byte(orgID.String()))
		hashValue := float64(hash.Sum32() % 100)
		if hashValue >= flag.Options.RolloutPercentage {
			return false
		}
	}

	return true
}

func (s *featureFlagService) clearFlagCache(flag string) {
	pattern := fmt.Sprintf("featureflag:%s:*", flag)
	s.cache.DeletePattern(context.Background(), pattern)
}

func getFeatureFlagCacheKey(flag string, orgID uuid.UUID) string {
	return fmt.Sprintf("featureflag:%s:%s", flag, orgID)
}
