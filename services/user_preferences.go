package services

import (
	"context"
	"time"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type UserPreferencesService interface {
	GetPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserPreferences) error
	GetPreference(ctx context.Context, userID uuid.UUID, key string) (interface{}, error)
	SetPreference(ctx context.Context, userID uuid.UUID, key string, value interface{}) error
	ResetPreferences(ctx context.Context, userID uuid.UUID) error
}

type userPreferencesService struct {
	repo  repository.Repository
	cache CacheService
}

func NewUserPreferencesService(
	repo repository.Repository,
	cache CacheService,
) UserPreferencesService {
	return &userPreferencesService{
		repo:  repo,
		cache: cache,
	}
}

func (s *userPreferencesService) GetPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreferences, error) {
	cacheKey := fmt.Sprintf("user_prefs:%s", userID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(*models.UserPreferences), nil
	}

	prefs, err := s.repo.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if prefs != nil {
		s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)
	}

	return prefs, nil
}

func (s *userPreferencesService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserPreferences) error {
	prefs.UpdatedAt = time.Now()

	if err := s.repo.UpdateUserPreferences(ctx, userID, prefs); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("user_prefs:%s", userID)
	s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)

	return nil
}

func (s *userPreferencesService) GetPreference(ctx context.Context, userID uuid.UUID, key string) (interface{}, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if prefs == nil || prefs.Settings == nil {
		return nil, nil
	}

	return prefs.Settings[key], nil
}

func (s *userPreferencesService) SetPreference(ctx context.Context, userID uuid.UUID, key string, value interface{}) error {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return err
	}

	if prefs == nil {
		prefs = &models.UserPreferences{
			UserID:    userID,
			Settings:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}
	}

	if prefs.Settings == nil {
		prefs.Settings = make(map[string]interface{})
	}

	prefs.Settings[key] = value
	prefs.UpdatedAt = time.Now()

	return s.UpdatePreferences(ctx, userID, prefs)
}

func (s *userPreferencesService) ResetPreferences(ctx context.Context, userID uuid.UUID) error {
	defaultPrefs := &models.UserPreferences{
		UserID: userID,
		Settings: map[string]interface{}{
			"theme":              "light",
			"language":           "en",
			"timezone":           "UTC",
			"itemsPerPage":       25,
			"defaultVaultView":   "list",
			"autoLogoutMinutes": 30,
			"showFavorites":      true,
			"showCategories":     true,
			"sortBy":             "name",
			"sortDirection":      "asc",
		},
		UpdatedAt: time.Now(),
	}

	return s.UpdatePreferences(ctx, userID, defaultPrefs)
}
