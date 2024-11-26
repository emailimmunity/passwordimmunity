package services

import (
	"context"
	"time"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type NotificationPreferencesService interface {
	GetPreferences(ctx context.Context, userID uuid.UUID) (*models.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *models.NotificationPreferences) error
	GetChannelPreference(ctx context.Context, userID uuid.UUID, channel string) (bool, error)
	GetEventPreference(ctx context.Context, userID uuid.UUID, eventType string) (bool, error)
	SetDefaultPreferences(ctx context.Context, userID uuid.UUID) error
}

type notificationPreferencesService struct {
	repo  repository.Repository
	cache CacheService
}

func NewNotificationPreferencesService(
	repo repository.Repository,
	cache CacheService,
) NotificationPreferencesService {
	return &notificationPreferencesService{
		repo:  repo,
		cache: cache,
	}
}

func (s *notificationPreferencesService) GetPreferences(ctx context.Context, userID uuid.UUID) (*models.NotificationPreferences, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("notification_prefs:%s", userID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(*models.NotificationPreferences), nil
	}

	// Get from database
	prefs, err := s.repo.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the preferences
	if prefs != nil {
		s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)
	}

	return prefs, nil
}

func (s *notificationPreferencesService) UpdatePreferences(ctx context.Context, userID uuid.UUID, prefs *models.NotificationPreferences) error {
	// Update in database
	if err := s.repo.UpdateNotificationPreferences(ctx, userID, prefs); err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("notification_prefs:%s", userID)
	s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)

	return nil
}

func (s *notificationPreferencesService) GetChannelPreference(ctx context.Context, userID uuid.UUID, channel string) (bool, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return false, err
	}

	if prefs == nil {
		return true, nil // Default to enabled if no preferences set
	}

	enabled, exists := prefs.Channels[channel]
	if !exists {
		return true, nil // Default to enabled if channel not specified
	}

	return enabled, nil
}

func (s *notificationPreferencesService) GetEventPreference(ctx context.Context, userID uuid.UUID, eventType string) (bool, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return false, err
	}

	if prefs == nil {
		return true, nil // Default to enabled if no preferences set
	}

	enabled, exists := prefs.Events[eventType]
	if !exists {
		return true, nil // Default to enabled if event not specified
	}

	return enabled, nil
}

func (s *notificationPreferencesService) SetDefaultPreferences(ctx context.Context, userID uuid.UUID) error {
	defaultPrefs := &models.NotificationPreferences{
		UserID: userID,
		Channels: map[string]bool{
			"email":     true,
			"push":      true,
			"in_app":    true,
			"sms":       false,
		},
		Events: map[string]bool{
			"security_alerts":    true,
			"login_attempts":     true,
			"password_changes":   true,
			"sharing_events":     true,
			"collection_updates": true,
			"policy_changes":     true,
			"system_updates":     false,
		},
		UpdatedAt: time.Now(),
	}

	return s.UpdatePreferences(ctx, userID, defaultPrefs)
}
