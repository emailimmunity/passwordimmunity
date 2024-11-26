package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type RateLimitService interface {
	Allow(ctx context.Context, key string) (bool, error)
	AllowN(ctx context.Context, key string, n int) (bool, error)
	Reset(ctx context.Context, key string) error
	GetLimit(ctx context.Context, key string) (*models.RateLimit, error)
	SetLimit(ctx context.Context, key string, limit models.RateLimit) error
}

type rateLimitService struct {
	repo     repository.Repository
	cache    CacheService
	limiters sync.Map
	config   models.RateLimitConfig
}

func NewRateLimitService(
	repo repository.Repository,
	cache CacheService,
	config models.RateLimitConfig,
) RateLimitService {
	return &rateLimitService{
		repo:     repo,
		cache:    cache,
		limiters: sync.Map{},
		config:   config,
	}
}

func (s *rateLimitService) Allow(ctx context.Context, key string) (bool, error) {
	return s.AllowN(ctx, key, 1)
}

func (s *rateLimitService) AllowN(ctx context.Context, key string, n int) (bool, error) {
	// Get or create limiter
	limiter, err := s.getLimiter(ctx, key)
	if err != nil {
		return false, err
	}

	// Check if request is allowed
	allowed := limiter.AllowN(time.Now(), n)

	// Update rate limit info in cache
	if err := s.updateRateLimit(ctx, key, limiter); err != nil {
		return false, err
	}

	return allowed, nil
}

func (s *rateLimitService) Reset(ctx context.Context, key string) error {
	// Delete from cache
	if err := s.cache.Delete(ctx, getRateLimitCacheKey(key)); err != nil {
		return err
	}

	// Remove from local limiters
	s.limiters.Delete(key)

	return nil
}

func (s *rateLimitService) GetLimit(ctx context.Context, key string) (*models.RateLimit, error) {
	// Try cache first
	limit, err := s.cache.Get(ctx, getRateLimitCacheKey(key))
	if err != nil {
		return nil, err
	}
	if limit != nil {
		return limit.(*models.RateLimit), nil
	}

	// Get from database
	return s.repo.GetRateLimit(ctx, key)
}

func (s *rateLimitService) SetLimit(ctx context.Context, key string, limit models.RateLimit) error {
	// Save to database
	if err := s.repo.SetRateLimit(ctx, key, &limit); err != nil {
		return err
	}

	// Update cache
	if err := s.cache.Set(ctx, getRateLimitCacheKey(key), &limit, time.Hour); err != nil {
		return err
	}

	// Reset existing limiter
	return s.Reset(ctx, key)
}

func (s *rateLimitService) getLimiter(ctx context.Context, key string) (*rate.Limiter, error) {
	// Check local cache
	if limiter, ok := s.limiters.Load(key); ok {
		return limiter.(*rate.Limiter), nil
	}

	// Get rate limit configuration
	limit, err := s.GetLimit(ctx, key)
	if err != nil {
		return nil, err
	}

	// Use default limits if none configured
	if limit == nil {
		limit = &models.RateLimit{
			Rate:     s.config.DefaultRate,
			Burst:    s.config.DefaultBurst,
			Duration: s.config.DefaultDuration,
		}
	}

	// Create new limiter
	limiter := rate.NewLimiter(rate.Limit(limit.Rate), limit.Burst)

	// Store in local cache
	s.limiters.Store(key, limiter)

	return limiter, nil
}

func (s *rateLimitService) updateRateLimit(ctx context.Context, key string, limiter *rate.Limiter) error {
	limit := &models.RateLimit{
		Key:       key,
		Rate:      float64(limiter.Limit()),
		Burst:     limiter.Burst(),
		Remaining: int(limiter.Tokens()),
		ResetAt:   time.Now().Add(time.Second),
		UpdatedAt: time.Now(),
	}

	return s.cache.Set(ctx, getRateLimitCacheKey(key), limit, time.Hour)
}

func getRateLimitCacheKey(key string) string {
	return fmt.Sprintf("ratelimit:%s", key)
}

func generateAPIRateLimitKey(userID uuid.UUID) string {
	return fmt.Sprintf("api:%s", userID.String())
}

func generateAuthRateLimitKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:%s", userID.String())
}

func generateIPRateLimitKey(ip string) string {
	return fmt.Sprintf("ip:%s", ip)
}
