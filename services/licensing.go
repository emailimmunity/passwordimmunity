package services

import (
	"context"
	"time"
	"encoding/json"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type LicenseType string

const (
	LicenseFree       LicenseType = "free"
	LicensePremium    LicenseType = "premium"
	LicenseEnterprise LicenseType = "enterprise"
)

type LicensingService interface {
	ValidateLicense(ctx context.Context, orgID uuid.UUID) error
	ActivateLicense(ctx context.Context, orgID uuid.UUID, key string) error
	DeactivateLicense(ctx context.Context, orgID uuid.UUID) error
	GetLicenseInfo(ctx context.Context, orgID uuid.UUID) (*models.License, error)
	CheckFeatureAccess(ctx context.Context, orgID uuid.UUID, feature string) (bool, error)
	ListFeatures(ctx context.Context, licenseType LicenseType) ([]string, error)
}

type licensingService struct {
	repo        repository.Repository
	audit       AuditService
}

func NewLicensingService(repo repository.Repository, audit AuditService) LicensingService {
	return &licensingService{
		repo:    repo,
		audit:   audit,
	}
}

func (s *licensingService) ValidateLicense(ctx context.Context, orgID uuid.UUID) error {
	license, err := s.GetLicenseInfo(ctx, orgID)
	if err != nil {
		return err
	}

	if license.ExpiresAt.Before(time.Now()) {
		return errors.New("license expired")
	}

	return nil
}

func (s *licensingService) ActivateLicense(ctx context.Context, orgID uuid.UUID, key string) error {
	// Validate license key format and signature
	licenseInfo, err := s.validateLicenseKey(key)
	if err != nil {
		return err
	}

	license := &models.License{
		ID:            uuid.New(),
		OrganizationID: orgID,
		Type:          string(licenseInfo.Type),
		Key:           key,
		Features:      licenseInfo.Features,
		IssuedAt:      time.Now(),
		ExpiresAt:     licenseInfo.ExpiresAt,
	}

	if err := s.repo.CreateLicense(ctx, license); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("license_activated", "License activated")
	metadata["license_type"] = license.Type
	if err := s.createAuditLog(ctx, "license.activated", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *licensingService) DeactivateLicense(ctx context.Context, orgID uuid.UUID) error {
	// Create audit log
	metadata := createBasicMetadata("license_deactivated", "License deactivated")
	if err := s.createAuditLog(ctx, "license.deactivated", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.DeleteLicense(ctx, orgID)
}

func (s *licensingService) GetLicenseInfo(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
	return s.repo.GetLicense(ctx, orgID)
}

func (s *licensingService) CheckFeatureAccess(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
	license, err := s.GetLicenseInfo(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Check if feature is included in license
	for _, f := range license.Features {
		if f == feature {
			return true, nil
		}
	}

	return false, nil
}

func (s *licensingService) ListFeatures(ctx context.Context, licenseType LicenseType) ([]string, error) {
	features := map[LicenseType][]string{
		LicenseFree: {
			"basic_vault",
			"basic_sharing",
			"basic_2fa",
		},
		LicensePremium: {
			"advanced_vault",
			"advanced_sharing",
			"advanced_2fa",
			"emergency_access",
			"priority_support",
		},
		LicenseEnterprise: {
			"sso_integration",
			"advanced_reporting",
			"custom_roles",
			"directory_sync",
			"api_access",
			"audit_logs",
			"enterprise_policies",
		},
	}

	return features[licenseType], nil
}

func (s *licensingService) validateLicenseKey(key string) (*models.LicenseInfo, error) {
	// Implement license key validation and parsing
	return nil, nil
}
