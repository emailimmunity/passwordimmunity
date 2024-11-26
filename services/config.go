package services

import (
	"context"
	"time"
	"sync"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
)

type ConfigService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	GetAll(ctx context.Context) (map[string]interface{}, error)
	Watch(ctx context.Context, key string) (<-chan ConfigUpdate, error)
}

type ConfigUpdate struct {
	Key      string
	Value    interface{}
	IsDelete bool
}

type configService struct {
	repo      repository.Repository
	cache     CacheService
	watchers  map[string][]chan ConfigUpdate
	mu        sync.RWMutex
}

func NewConfigService(
	repo repository.Repository,
	cache CacheService,
) ConfigService {
	return &configService{
		repo:     repo,
		cache:    cache,
		watchers: make(map[string][]chan ConfigUpdate),
	}
}

func (s *configService) Get(ctx context.Context, key string) (interface{}, error) {
	// Try cache first
	value, err := s.cache.Get(ctx, getConfigCacheKey(key))
	if err == nil && value != nil {
		return value, nil
	}

	// Get from database
	config, err := s.repo.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}

	// Cache the value
	if err := s.cache.Set(ctx, getConfigCacheKey(key), config.Value, time.Hour); err != nil {
		return nil, err
	}

	return config.Value, nil
}

func (s *configService) Set(ctx context.Context, key string, value interface{}) error {
	config := &models.Config{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := s.repo.SetConfig(ctx, config); err != nil {
		return err
	}

	// Update cache
	if err := s.cache.Set(ctx, getConfigCacheKey(key), value, time.Hour); err != nil {
		return err
	}

	// Notify watchers
	s.notifyWatchers(key, ConfigUpdate{
		Key:   key,
		Value: value,
	})

	return nil
}

func (s *configService) Delete(ctx context.Context, key string) error {
	// Delete from database
	if err := s.repo.DeleteConfig(ctx, key); err != nil {
		return err
	}

	// Delete from cache
	if err := s.cache.Delete(ctx, getConfigCacheKey(key)); err != nil {
		return err
	}

	// Notify watchers
	s.notifyWatchers(key, ConfigUpdate{
		Key:      key,
		IsDelete: true,
	})

	return nil
}

func (s *configService) GetAll(ctx context.Context) (map[string]interface{}, error) {
	configs, err := s.repo.GetAllConfigs(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, config := range configs {
		result[config.Key] = config.Value
	}

	return result, nil
}

func (s *configService) Watch(ctx context.Context, key string) (<-chan ConfigUpdate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan ConfigUpdate, 1)
	s.watchers[key] = append(s.watchers[key], ch)

	// Clean up when context is done
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()

		watchers := s.watchers[key]
		for i, watcher := range watchers {
			if watcher == ch {
				s.watchers[key] = append(watchers[:i], watchers[i+1:]...)
				close(ch)
				break
			}
		}
	}()

	return ch, nil
}

func (s *configService) notifyWatchers(key string, update ConfigUpdate) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.watchers[key] {
		select {
		case ch <- update:
		default:
			// Skip if channel is blocked
		}
	}
}

func getConfigCacheKey(key string) string {
	return fmt.Sprintf("config:%s", key)
}
