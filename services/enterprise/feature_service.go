package enterprise

import (
    "context"
    "fmt"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/emailimmunity/passwordimmunity/db/repository"
)

type FeatureService interface {
    ActivateFeature(ctx context.Context, orgID, featureID, paymentID string, duration time.Duration) error
    ActivateBundle(ctx context.Context, orgID, bundleID, paymentID string, duration time.Duration) error
    IsFeatureActive(ctx context.Context, orgID, featureID string) (bool, error)
    IsBundleActive(ctx context.Context, orgID, bundleID string) (bool, error)
    GetActiveFeatures(ctx context.Context, orgID string) ([]string, error)
    DeactivateFeature(ctx context.Context, orgID, featureID string) error
    DeactivateBundle(ctx context.Context, orgID, bundleID string) error
}

type featureService struct {
    repo repository.FeatureActivationRepository
}

func NewFeatureService(repo repository.FeatureActivationRepository) FeatureService {
    return &featureService{repo: repo}
}

func (s *featureService) ActivateFeature(ctx context.Context, orgID, featureID, paymentID string, duration time.Duration) error {
    if err := validateFeature(featureID); err != nil {
        return fmt.Errorf("invalid feature: %w", err)
    }

    activation := &models.FeatureActivation{
        OrganizationID: orgID,
        FeatureID:      &featureID,
        Status:         "active",
        ExpiresAt:      time.Now().Add(duration),
        PaymentID:      paymentID,
        Currency:       "EUR", // Default currency, should be configurable
        Amount:         getFeaturePrice(featureID),
    }

    return s.repo.Create(ctx, activation)
}

func (s *featureService) ActivateBundle(ctx context.Context, orgID, bundleID, paymentID string, duration time.Duration) error {
    if err := validateBundle(bundleID); err != nil {
        return fmt.Errorf("invalid bundle: %w", err)
    }

    activation := &models.FeatureActivation{
        OrganizationID: orgID,
        BundleID:      &bundleID,
        Status:        "active",
        ExpiresAt:     time.Now().Add(duration),
        PaymentID:     paymentID,
        Currency:      "EUR", // Default currency, should be configurable
        Amount:        getBundlePrice(bundleID),
    }

    return s.repo.Create(ctx, activation)
}

func (s *featureService) IsFeatureActive(ctx context.Context, orgID, featureID string) (bool, error) {
    activation, err := s.repo.GetByFeature(ctx, orgID, featureID)
    if err != nil {
        return false, err
    }
    return activation != nil && activation.IsActive(), nil
}

func (s *featureService) IsBundleActive(ctx context.Context, orgID, bundleID string) (bool, error) {
    activation, err := s.repo.GetByBundle(ctx, orgID, bundleID)
    if err != nil {
        return false, err
    }
    return activation != nil && activation.IsActive(), nil
}

func (s *featureService) GetActiveFeatures(ctx context.Context, orgID string) ([]string, error) {
    activations, err := s.repo.GetActiveByOrganization(ctx, orgID)
    if err != nil {
        return nil, err
    }

    var features []string
    for _, activation := range activations {
        if activation.FeatureID != nil {
            features = append(features, *activation.FeatureID)
        }
        if activation.BundleID != nil {
            bundleFeatures := getBundleFeatures(*activation.BundleID)
            features = append(features, bundleFeatures...)
        }
    }
    return features, nil
}

func (s *featureService) DeactivateFeature(ctx context.Context, orgID, featureID string) error {
    activation, err := s.repo.GetByFeature(ctx, orgID, featureID)
    if err != nil {
        return err
    }
    if activation == nil {
        return fmt.Errorf("feature not found")
    }
    return s.repo.UpdateStatus(ctx, activation.ID, "cancelled")
}

func (s *featureService) DeactivateBundle(ctx context.Context, orgID, bundleID string) error {
    activation, err := s.repo.GetByBundle(ctx, orgID, bundleID)
    if err != nil {
        return err
    }
    if activation == nil {
        return fmt.Errorf("bundle not found")
    }
    return s.repo.UpdateStatus(ctx, activation.ID, "cancelled")
}

func validateFeature(featureID string) error {
    validFeatures := map[string]bool{
        "advanced_sso":       true,
        "directory_sync":     true,
        "advanced_reporting": true,
        "custom_roles":       true,
        "advanced_policies":  true,
        "priority_support":   true,
        "advanced_audit":     true,
        "emergency_access":   true,
        "custom_groups":      true,
        "advanced_api":       true,
    }
    if !validFeatures[featureID] {
        return fmt.Errorf("invalid feature ID: %s", featureID)
    }
    return nil
}

func validateBundle(bundleID string) error {
    validBundles := map[string]bool{
        "security":    true,
        "enterprise": true,
        "compliance": true,
        "advanced":   true,
    }
    if !validBundles[bundleID] {
        return fmt.Errorf("invalid bundle ID: %s", bundleID)
    }
    return nil
}

func getFeaturePrice(featureID string) float64 {
    prices := map[string]float64{
        "advanced_sso":       49.99,
        "directory_sync":     39.99,
        "advanced_reporting": 29.99,
        "custom_roles":       19.99,
        "advanced_policies":  29.99,
        "priority_support":   99.99,
        "advanced_audit":     39.99,
        "emergency_access":   49.99,
        "custom_groups":      19.99,
        "advanced_api":       59.99,
    }
    return prices[featureID]
}

func getBundlePrice(bundleID string) float64 {
    prices := map[string]float64{
        "security":    199.99,
        "enterprise": 299.99,
        "compliance": 249.99,
        "advanced":   399.99,
    }
    return prices[bundleID]
}

func getBundleFeatures(bundleID string) []string {
    bundles := map[string][]string{
        "security": {
            "advanced_sso",
            "emergency_access",
            "advanced_policies",
        },
        "enterprise": {
            "advanced_sso",
            "directory_sync",
            "custom_roles",
            "custom_groups",
            "advanced_api",
        },
        "compliance": {
            "advanced_audit",
            "advanced_reporting",
            "advanced_policies",
        },
        "advanced": {
            "advanced_sso",
            "directory_sync",
            "advanced_reporting",
            "custom_roles",
            "advanced_policies",
            "priority_support",
            "advanced_audit",
            "emergency_access",
            "custom_groups",
            "advanced_api",
        },
    }
    return bundles[bundleID]
}
