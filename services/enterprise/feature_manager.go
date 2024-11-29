package enterprise

import (
	"context"
	"fmt"
	"time"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/shopspring/decimal"
)

type FeatureManager struct {
	logger     Logger
	repository RepositoryInterface
	pricing    *PricingManager
}

type FeatureActivation struct {
	OrganizationID string
	FeatureID      string
	BundleID       string
	TierID         string
	ExpiresAt      time.Time
	PaymentID      string
	Currency       string
	Amount         decimal.Decimal
	IsYearly       bool
}

func NewFeatureManager(logger Logger, repository RepositoryInterface, pricing *PricingManager) *FeatureManager {
	return &FeatureManager{
		logger:     logger,
		repository: repository,
		pricing:    pricing,
	}
}

func (m *FeatureManager) ActivateFeature(ctx context.Context, activation FeatureActivation) error {
	if err := m.validateActivation(&activation); err != nil {
		return fmt.Errorf("invalid activation: %w", err)
	}

	// Handle different activation types
	switch {
	case activation.TierID != "":
		return m.activateTier(ctx, &activation)
	case activation.BundleID != "":
		return m.activateBundle(ctx, &activation)
	default:
		return m.activateSingleFeature(ctx, &activation)
	}
}

func (m *FeatureManager) activateTier(ctx context.Context, activation *FeatureActivation) error {
	price, err := config.GetTierPrice(activation.TierID, activation.Currency, activation.IsYearly)
	if err != nil {
		return fmt.Errorf("invalid tier: %w", err)
	}

	if !price.Equal(activation.Amount) {
		return fmt.Errorf("price mismatch for tier %s", activation.TierID)
	}

	features, err := config.GetTierFeatures(activation.TierID)
	if err != nil {
		return fmt.Errorf("failed to get tier features: %w", err)
	}

	// Create individual activations for each feature in the tier
	for _, featureID := range features {
		featureActivation := *activation
		featureActivation.FeatureID = featureID

		if err := m.repository.StoreActivation(ctx, featureActivation); err != nil {
			return fmt.Errorf("failed to store feature activation: %w", err)
		}
	}

	m.logger.Info("Tier activation successful",
		"tier_id", activation.TierID,
		"organization_id", activation.OrganizationID,
		"expires_at", activation.ExpiresAt,
	)

	return nil
}

func (m *FeatureManager) activateBundle(ctx context.Context, activation *FeatureActivation) error {
	price, features, err := config.GetBundlePrice(activation.BundleID, activation.Currency)
	if err != nil {
		return fmt.Errorf("invalid bundle: %w", err)
	}

	if !price.Equal(activation.Amount) {
		return fmt.Errorf("price mismatch for bundle %s", activation.BundleID)
	}

	// Create individual activations for each feature in the bundle
	for _, featureID := range features {
		featureActivation := *activation
		featureActivation.FeatureID = featureID
		featureActivation.BundleID = activation.BundleID

		if err := m.repository.StoreActivation(ctx, featureActivation); err != nil {
			return fmt.Errorf("failed to store feature activation: %w", err)
		}
	}

	return nil
}

func (m *FeatureManager) activateSingleFeature(ctx context.Context, activation *FeatureActivation) error {
	if err := m.checkFeatureDependencies(ctx, activation.OrganizationID, activation.FeatureID); err != nil {
		return err
	}

	return m.repository.StoreActivation(ctx, *activation)
}

func (m *FeatureManager) validateActivation(activation *FeatureActivation) error {
	if activation.OrganizationID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if activation.FeatureID == "" && activation.BundleID == "" && activation.TierID == "" {
		return fmt.Errorf("either feature ID, bundle ID, or tier ID is required")
	}
	if activation.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration time must be in the future")
	}
	return nil
}

func (m *FeatureManager) checkFeatureDependencies(ctx context.Context, orgID, featureID string) error {
	deps, err := config.GetFeatureDependencies(featureID)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if !m.IsFeatureEnabled(ctx, orgID, dep) {
			return fmt.Errorf("required feature %s is not enabled", dep)
		}
	}
	return nil
}

func (m *FeatureManager) IsFeatureEnabled(ctx context.Context, organizationID, featureID string) bool {
	activation, err := m.repository.GetActivation(ctx, organizationID, featureID)
	if err != nil {
		m.logger.Error("Failed to get activation",
			"error", err,
			"feature_id", featureID,
			"organization_id", organizationID,
		)
		return false
	}

	if activation == nil {
		return false
	}

	if time.Now().After(activation.ExpiresAt) {
		m.logger.Info("Feature activation expired",
			"feature_id", featureID,
			"organization_id", organizationID,
			"expired_at", activation.ExpiresAt,
		)
		return false
	}

	return true
}

func (m *FeatureManager) GetActiveFeatures(ctx context.Context, organizationID string) ([]string, error) {
	return m.repository.GetActiveFeatures(ctx, organizationID)
}
