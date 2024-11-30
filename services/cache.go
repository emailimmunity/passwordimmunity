package services

import (
	"context"
	"time"
	"sync"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
)

type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), expiration time.Duration) (interface{}, error)
}

type cacheService struct {
	redis  *redis.Client
	local  sync.Map
	config models.CacheConfig
}

func NewCacheService(redisClient *redis.Client, config models.CacheConfig) CacheService {
	return &cacheService{
		redis:  redisClient,
		local:  sync.Map{},
		config: config,
	}
}

func (s *cacheService) Get(ctx context.Context, key string) (interface{}, error) {
	// Try local cache first
	if value, ok := s.local.Load(key); ok {
		return value, nil
	}

	// Try Redis
	value, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal value
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	// Store in local cache
	s.local.Store(key, result)

	return result, nil
}

func (s *cacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Marshal value
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Store in Redis
	if err := s.redis.Set(ctx, key, bytes, expiration).Err(); err != nil {
		return err
	}

	// Store in local cache
	s.local.Store(key, value)

	return nil
}

func (s *cacheService) Delete(ctx context.Context, key string) error {
	// Remove from Redis
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	// Remove from local cache
	s.local.Delete(key)

	return nil
}

func (s *cacheService) Clear(ctx context.Context) error {
	// Clear Redis
	if err := s.redis.FlushAll(ctx).Err(); err != nil {
		return err
	}

	// Clear local cache
	s.local = sync.Map{}

	return nil
}

func (s *cacheService) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error), expiration time.Duration) (interface{}, error) {
	// Try getting from cache first
	if value, err := s.Get(ctx, key); err == nil && value != nil {
		return value, nil
	}

	// Generate value
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Cache value
	if err := s.Set(ctx, key, value, expiration); err != nil {
		return nil, err
	}

	return value, nil
}

// Helper functions for generating cache keys
func generateUserCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s", userID.String())
}

func generateOrgCacheKey(orgID uuid.UUID) string {
	return fmt.Sprintf("org:%s", orgID.String())
}

func generateVaultItemCacheKey(itemID uuid.UUID) string {
	return fmt.Sprintf("vault_item:%s", itemID.String())
}

func generateCollectionCacheKey(collectionID uuid.UUID) string {
	return fmt.Sprintf("collection:%s", collectionID.String())
}
