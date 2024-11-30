package featureflag

import (
    "context"
    "time"

    "github.com/emailimmunity/passwordimmunity/config"
    "github.com/emailimmunity/passwordimmunity/services/licensing"
    "github.com/google/uuid"
)

type Service interface {
    IsFeatureEnabled(ctx context.Context, orgID uuid.UUID, feature string) (bool, error)
    GetAvailableFeatures(ctx context.Context, orgID uuid.UUID) ([]string, error)
    IsFeatureEnabledForTime(ctx context.Context, orgID uuid.UUID, feature string, at time.Time) (bool, error)
    GetFeatureMetadata(ctx context.Context, orgID uuid.UUID, feature string) (map[string]interface{}, error)
}

type service struct {
    licenseService licensing.Service
}

func NewService(licenseService licensing.Service) Service {
    return &service{
        licenseService: licenseService,
    }
}

func (s *service) IsFeatureEnabled(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
    return s.IsFeatureEnabledForTime(ctx, orgID, feature, time.Now())
}

func (s *service) IsFeatureEnabledForTime(ctx context.Context, orgID uuid.UUID, feature string, at time.Time) (bool, error) {
    featureConfig, exists := config.Features[feature]
    if !exists {
        return false, nil
    }

    license, err := s.licenseService.GetActiveLicense(ctx, orgID)
    if err != nil {
        return false, err
    }

    if license == nil || !license.IsValidAt(at) {
        return config.IsFeatureAvailableForTier(featureConfig.MinTier, "free"), nil
    }

    return config.IsFeatureAvailableForTier(featureConfig.MinTier, license.Type), nil
}

func (s *service) GetAvailableFeatures(ctx context.Context, orgID uuid.UUID) ([]string, error) {
    licenseType, err := s.licenseService.GetLicenseType(ctx, orgID)
    if err != nil {
        return nil, err
    }

    features := config.GetFeaturesByTier(licenseType)
    result := make([]string, len(features))
    for i, feature := range features {
        result[i] = feature.Name
    }

    return result, nil
}

func (s *service) GetFeatureMetadata(ctx context.Context, orgID uuid.UUID, feature string) (map[string]interface{}, error) {
    featureConfig, exists := config.Features[feature]
    if !exists {
        return nil, nil
    }

    enabled, err := s.IsFeatureEnabled(ctx, orgID, feature)
    if err != nil {
        return nil, err
    }

    license, err := s.licenseService.GetActiveLicense(ctx, orgID)
    if err != nil {
        return nil, err
    }

    metadata := map[string]interface{}{
        "enabled":          enabled,
        "name":            featureConfig.Name,
        "description":     featureConfig.Description,
        "min_tier":        featureConfig.MinTier,
        "dependencies":    featureConfig.Dependencies,
        "requires_license": featureConfig.MinTier != "free",
    }

    if license != nil {
        metadata["license_type"] = license.Type
        metadata["valid_until"] = license.ValidUntil
    }

    return metadata, nil
}
