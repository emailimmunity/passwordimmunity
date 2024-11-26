package licensing

import (
    "context"
    "fmt"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/emailimmunity/passwordimmunity/db/repository"
    "github.com/google/uuid"
)

type Service interface {
    GetActiveLicense(ctx context.Context, orgID uuid.UUID) (*models.License, error)
    HasFeature(ctx context.Context, orgID uuid.UUID, feature string) (bool, error)
    CheckLicenseExpiry(ctx context.Context) error
    GetLicenseType(ctx context.Context, orgID uuid.UUID) (string, error)
    IsLicenseValidAt(ctx context.Context, orgID uuid.UUID, at time.Time) (bool, error)
    AddFeature(ctx context.Context, orgID uuid.UUID, feature string) error
    RemoveFeature(ctx context.Context, orgID uuid.UUID, feature string) error
    ActivateLicense(ctx context.Context, orgID uuid.UUID, licenseType string, validUntil time.Time) error
}

type service struct {
    repo repository.Repository
}

func NewService(repo repository.Repository) Service {
    return &service{
        repo: repo,
    }
}

func (s *service) GetActiveLicense(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
    license, err := s.repo.GetActiveLicenseByOrganization(ctx, orgID)
    if err != nil {
        return nil, err
    }

    if license != nil && !license.IsValid() {
        license.Status = "expired"
        if err := s.repo.UpdateLicense(ctx, license); err != nil {
            return nil, err
        }
        return nil, nil
    }

    return license, nil
}

func (s *service) HasFeature(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
    license, err := s.GetActiveLicense(ctx, orgID)
    if err != nil {
        return false, err
    }

    if license == nil {
        return false, nil
    }

    for _, f := range license.Features {
        if f == feature {
            return true, nil
        }
    }

    return false, nil
}

func (s *service) CheckLicenseExpiry(ctx context.Context) error {
    expiredLicenses, err := s.repo.GetExpiredLicenses(ctx)
    if err != nil {
        return err
    }

    for _, license := range expiredLicenses {
        license.Status = "expired"
        if err := s.repo.UpdateLicense(ctx, &license); err != nil {
            return err
        }
    }

    return nil
}

func (s *service) GetLicenseType(ctx context.Context, orgID uuid.UUID) (string, error) {
    license, err := s.GetActiveLicense(ctx, orgID)
    if err != nil {
        return "", err
    }

    if license == nil {
        return "free", nil
    }

    return license.Type, nil
}

func (s *service) IsLicenseValidAt(ctx context.Context, orgID uuid.UUID, at time.Time) (bool, error) {
    license, err := s.repo.GetActiveLicenseByOrganization(ctx, orgID)
    if err != nil {
        return false, err
    }

    if license == nil {
        return false, nil
    }

    return license.Status == "active" && at.Before(license.ValidUntil), nil
}

func (s *service) AddFeature(ctx context.Context, orgID uuid.UUID, feature string) error {
    license, err := s.GetActiveLicense(ctx, orgID)
    if err != nil {
        return err
    }

    if license == nil {
        return fmt.Errorf("no active license found")
    }

    // Check if feature already exists
    for _, f := range license.Features {
        if f == feature {
            return nil
        }
    }

    license.Features = append(license.Features, feature)
    return s.repo.UpdateLicense(ctx, license)
}

func (s *service) RemoveFeature(ctx context.Context, orgID uuid.UUID, feature string) error {
    license, err := s.GetActiveLicense(ctx, orgID)
    if err != nil {
        return err
    }

    if license == nil {
        return fmt.Errorf("no active license found")
    }

    // Remove feature if it exists
    features := make([]string, 0, len(license.Features))
    for _, f := range license.Features {
        if f != feature {
            features = append(features, f)
        }
    }

    license.Features = features
    return s.repo.UpdateLicense(ctx, license)
}

func (s *service) ActivateLicense(ctx context.Context, orgID uuid.UUID, licenseType string, validUntil time.Time) error {
    // Create new license
    license := &models.License{
        ID:           uuid.New(),
        OrganizationID: orgID,
        Type:         licenseType,
        Status:       "active",
        ValidUntil:   validUntil,
        Features:     s.getFeaturesForLicenseType(licenseType),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    // Deactivate any existing licenses
    if err := s.repo.DeactivateOrganizationLicenses(ctx, orgID); err != nil {
        return fmt.Errorf("failed to deactivate existing licenses: %w", err)
    }

    // Create new license
    if err := s.repo.CreateLicense(ctx, license); err != nil {
        return fmt.Errorf("failed to create new license: %w", err)
    }

    return nil
}

func (s *service) getFeaturesForLicenseType(licenseType string) []string {
    switch licenseType {
    case "enterprise":
        return []string{
            "basic_vault", "basic_2fa", "advanced_2fa", "emergency_access",
            "priority_support", "basic_api_access", "basic_reporting",
            "sso", "directory_sync", "enterprise_policies", "advanced_reporting",
            "custom_roles", "advanced_groups", "multi_tenant", "advanced_vault",
            "cross_org_management", "api_access",
        }
    case "premium":
        return []string{
            "basic_vault", "basic_2fa", "advanced_2fa", "emergency_access",
            "priority_support", "basic_api_access", "basic_reporting",
        }
    default: // free
        return []string{
            "basic_vault", "basic_2fa",
        }
    }
}
