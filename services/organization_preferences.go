package services

import (
	"context"
	"time"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type OrganizationPreferencesService interface {
	GetPreferences(ctx context.Context, orgID uuid.UUID) (*models.OrganizationPreferences, error)
	UpdatePreferences(ctx context.Context, orgID uuid.UUID, prefs *models.OrganizationPreferences) error
	GetPreference(ctx context.Context, orgID uuid.UUID, key string) (interface{}, error)
	SetPreference(ctx context.Context, orgID uuid.UUID, key string, value interface{}) error
	ResetPreferences(ctx context.Context, orgID uuid.UUID) error
	ApplyTemplate(ctx context.Context, orgID uuid.UUID, templateID string) error
}

type organizationPreferencesService struct {
	repo      repository.Repository
	cache     CacheService
	audit     AuditService
}

func NewOrganizationPreferencesService(
	repo repository.Repository,
	cache CacheService,
	audit AuditService,
) OrganizationPreferencesService {
	return &organizationPreferencesService{
		repo:  repo,
		cache: cache,
		audit: audit,
	}
}

func (s *organizationPreferencesService) GetPreferences(ctx context.Context, orgID uuid.UUID) (*models.OrganizationPreferences, error) {
	cacheKey := fmt.Sprintf("org_prefs:%s", orgID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(*models.OrganizationPreferences), nil
	}

	prefs, err := s.repo.GetOrganizationPreferences(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if prefs != nil {
		s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)
	}

	return prefs, nil
}

func (s *organizationPreferencesService) UpdatePreferences(ctx context.Context, orgID uuid.UUID, prefs *models.OrganizationPreferences) error {
	prefs.UpdatedAt = time.Now()

	if err := s.repo.UpdateOrganizationPreferences(ctx, orgID, prefs); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("org_prefs:%s", orgID)
	s.cache.Set(ctx, cacheKey, prefs, time.Hour*24)

	metadata := map[string]interface{}{
		"org_id": orgID,
		"changes": prefs.Settings,
	}
	if err := s.audit.Log(ctx, "organization.preferences.updated", metadata); err != nil {
		return err
	}

	return nil
}

func (s *organizationPreferencesService) GetPreference(ctx context.Context, orgID uuid.UUID, key string) (interface{}, error) {
	prefs, err := s.GetPreferences(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if prefs == nil || prefs.Settings == nil {
		return nil, nil
	}

	return prefs.Settings[key], nil
}

func (s *organizationPreferencesService) SetPreference(ctx context.Context, orgID uuid.UUID, key string, value interface{}) error {
	prefs, err := s.GetPreferences(ctx, orgID)
	if err != nil {
		return err
	}

	if prefs == nil {
		prefs = &models.OrganizationPreferences{
			OrganizationID: orgID,
			Settings:       make(map[string]interface{}),
			CreatedAt:      time.Now(),
		}
	}

	if prefs.Settings == nil {
		prefs.Settings = make(map[string]interface{})
	}

	prefs.Settings[key] = value
	prefs.UpdatedAt = time.Now()

	return s.UpdatePreferences(ctx, orgID, prefs)
}

func (s *organizationPreferencesService) ResetPreferences(ctx context.Context, orgID uuid.UUID) error {
	defaultPrefs := &models.OrganizationPreferences{
		OrganizationID: orgID,
		Settings: map[string]interface{}{
			"passwordPolicy": map[string]interface{}{
				"minLength":        12,
				"requireUppercase": true,
				"requireLowercase": true,
				"requireNumbers":   true,
				"requireSpecial":   true,
				"maxAge":          90,
			},
			"sessionPolicy": map[string]interface{}{
				"maxSessionDuration": 720,
				"requireMFA":         true,
				"mfaRememberDays":    30,
			},
			"vaultPolicy": map[string]interface{}{
				"allowPersonalVault":    true,
				"allowSharing":          true,
				"allowExport":           true,
				"requireCollections":    false,
				"autoRotatePasswords":   false,
				"passwordRotationDays":  180,
			},
		},
		UpdatedAt: time.Now(),
	}

	return s.UpdatePreferences(ctx, orgID, defaultPrefs)
}

func (s *organizationPreferencesService) ApplyTemplate(ctx context.Context, orgID uuid.UUID, templateID string) error {
	template, err := s.repo.GetPreferenceTemplate(ctx, templateID)
	if err != nil {
		return err
	}

	if template == nil {
		return fmt.Errorf("template not found: %s", templateID)
	}

	prefs := &models.OrganizationPreferences{
		OrganizationID: orgID,
		Settings:       template.Settings,
		UpdatedAt:      time.Now(),
	}

	if err := s.UpdatePreferences(ctx, orgID, prefs); err != nil {
		return err
	}

	metadata := map[string]interface{}{
		"org_id":      orgID,
		"template_id": templateID,
	}
	if err := s.audit.Log(ctx, "organization.preferences.template_applied", metadata); err != nil {
		return err
	}

	return nil
}
