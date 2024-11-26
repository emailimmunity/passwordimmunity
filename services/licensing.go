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
    // Decode and verify license key
    parts := strings.Split(key, ".")
    if len(parts) != 3 {
        return nil, errors.New("invalid license key format")
    }

    // Verify signature
    signature := parts[2]
    payload := parts[0] + "." + parts[1]
    if err := s.verifySignature(payload, signature); err != nil {
        return nil, fmt.Errorf("invalid license signature: %w", err)
    }

    // Decode license info
    var licenseInfo models.LicenseInfo
    decodedPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return nil, fmt.Errorf("failed to decode license payload: %w", err)
    }

    if err := json.Unmarshal(decodedPayload, &licenseInfo); err != nil {
        return nil, fmt.Errorf("failed to parse license info: %w", err)
    }

    // Validate license info
    if licenseInfo.ExpiresAt.Before(time.Now()) {
        return nil, errors.New("license has expired")
    }

    if !s.isValidLicenseType(licenseInfo.Type) {
        return nil, errors.New("invalid license type")
    }

    return &licenseInfo, nil
}

func (s *licensingService) verifySignature(payload, signature string) error {
    // Decode signature
    sig, err := base64.RawURLEncoding.DecodeString(signature)
    if err != nil {
        return fmt.Errorf("failed to decode signature: %w", err)
    }

    // Get public key for verification
    publicKey, err := s.repo.GetLicensePublicKey(context.Background())
    if err != nil {
        return fmt.Errorf("failed to get license public key: %w", err)
    }

    // Create hash of payload
    h := sha256.New()
    h.Write([]byte(payload))
    hash := h.Sum(nil)

    // Verify signature
    err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash, sig)
    if err != nil {
        return fmt.Errorf("signature verification failed: %w", err)
    }

    return nil
}

func (s *licensingService) isValidLicenseType(licenseType LicenseType) bool {
    validTypes := map[LicenseType]bool{
        LicenseFree:       true,
        LicensePremium:    true,
        LicenseEnterprise: true,
    }
    return validTypes[licenseType]
}
