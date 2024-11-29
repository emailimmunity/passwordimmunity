package enterprise

import (
	"context"
	"errors"
	"fmt"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/emailimmunity/passwordimmunity/services/pricing"
)

var (
	ErrInvalidFeature     = errors.New("invalid feature")
	ErrFeatureNotAllowed  = errors.New("feature not allowed in current tier")
	ErrActivationFailed   = errors.New("feature activation failed")
	ErrLicenseRequired    = errors.New("valid license required for enterprise features")
	ErrFeatureNotFound    = errors.New("feature activation not found")
	ErrRepositoryFailure  = errors.New("repository operation failed")
)

type Repository interface {
	Create(ctx context.Context, activation *FeatureActivation) error
	Delete(ctx context.Context, featureID, orgID string) error
	Get(ctx context.Context, featureID, orgID string) (*FeatureActivation, error)
	GetAllActive(ctx context.Context, orgID string) ([]FeatureActivation, error)
	Update(ctx context.Context, activation *FeatureActivation) error
	GetHistory(ctx context.Context, featureID, orgID string) ([]*FeatureActivation, error)
}

type ActivationService struct {
	licenseService *licensing.Service
	repository    Repository
	paymentService payment.Service
}

func NewActivationService(licenseService *licensing.Service, repository Repository, paymentService payment.Service) *ActivationService {
	return &ActivationService{
		licenseService: licenseService,
		repository:    repository,
		paymentService: paymentService,
	}
}

// ActivateFeature attempts to activate an enterprise feature for an organization
func (s *ActivationService) ActivateFeature(ctx context.Context, orgID string, featureID string) error {
	// Check if feature exists and is an enterprise feature
	if !config.IsEnterpriseFeature(featureID) {
		return ErrInvalidFeature
	}

	// Verify organization has valid license
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return ErrLicenseRequired
	}

	// Check if feature is allowed in organization's tier
	if !config.IsFeatureInTier(license.Tier, featureID) {
		return ErrFeatureNotAllowed
	}

	// Perform actual feature activation
	err = s.activateFeatureForOrg(ctx, orgID, featureID)
	if err != nil {
		return ErrActivationFailed
	}

	return nil
}

// ActivateBundle activates all features in a bundle for an organization
func (s *ActivationService) ActivateBundle(ctx context.Context, orgID string, bundleID string) error {
	features, err := config.GetBundleFeatures(bundleID)
	if err != nil {
		return err
	}

	for _, feature := range features {
		err := s.ActivateFeature(ctx, orgID, feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeactivateBundle deactivates all features in a bundle for an organization
func (s *ActivationService) DeactivateBundle(ctx context.Context, orgID string, bundleID string) error {
	features, err := config.GetBundleFeatures(bundleID)
	if err != nil {
		return err
	}

	for _, feature := range features {
		err := s.DeactivateFeature(ctx, orgID, feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsFeatureAvailable checks if a feature is available in the organization's current tier
func (s *ActivationService) IsFeatureAvailable(ctx context.Context, orgID string, featureID string) (bool, error) {
	if !config.IsEnterpriseFeature(featureID) {
		return false, ErrInvalidFeature
	}

	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return false, ErrLicenseRequired
	}

	return config.IsFeatureInTier(license.Tier, featureID), nil
}

// GetAvailableFeatures returns all features available in the organization's current tier
func (s *ActivationService) GetAvailableFeatures(ctx context.Context, orgID string) ([]string, error) {
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return nil, ErrLicenseRequired
	}

	return config.GetFeaturesForTier(license.Tier), nil
}

// GetAvailableBundles returns all feature bundles available in the organization's current tier
func (s *ActivationService) GetAvailableBundles(ctx context.Context, orgID string) ([]string, error) {
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return nil, ErrLicenseRequired
	}

	return config.GetBundlesForTier(license.Tier), nil
}

// IsBundleAvailable checks if a bundle is available in the organization's current tier
func (s *ActivationService) IsBundleAvailable(ctx context.Context, orgID string, bundleID string) (bool, error) {
	if !config.IsValidBundle(bundleID) {
		return false, config.ErrInvalidBundle
	}

	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return false, ErrLicenseRequired
	}

	return config.IsBundleInTier(license.Tier, bundleID), nil
}

// HasBundleAccess checks if an organization has access to all features in a bundle
func (s *ActivationService) HasBundleAccess(ctx context.Context, orgID string, bundleID string) (bool, error) {
    features, err := config.GetBundleFeatures(bundleID)
    if err != nil {
        return false, err
    }

    for _, feature := range features {
        active, err := s.IsFeatureActive(ctx, orgID, feature)
        if err != nil {
            return false, err
        }
        if !active {
            return false, nil
        }
    }

    return true, nil
}

// GetActiveFeatures returns all active features for an organization
func (s *ActivationService) GetActiveFeatures(ctx context.Context, orgID string) ([]string, error) {
	activations, err := s.repository.GetAllActive(ctx, orgID)
	if err != nil {
		return nil, ErrRepositoryFailure
	}

	features := make([]string, 0, len(activations))
	for _, activation := range activations {
		features = append(features, activation.FeatureID)
	}

	return features, nil
}

// GetBundleFeatures returns all features that are part of a specific bundle
func (s *ActivationService) GetBundleFeatures(ctx context.Context, bundleID string) ([]string, error) {
	if !config.IsValidBundle(bundleID) {
		return nil, config.ErrInvalidBundle
	}

	features, err := config.GetBundleFeatures(bundleID)
	if err != nil {
		return nil, err
	}

	return features, nil
}

// IsFeatureInBundle checks if a feature is part of a specific bundle
func (s *ActivationService) IsFeatureInBundle(ctx context.Context, featureID string, bundleID string) (bool, error) {
	if !config.IsEnterpriseFeature(featureID) {
		return false, ErrInvalidFeature
	}

	if !config.IsValidBundle(bundleID) {
		return false, config.ErrInvalidBundle
	}

	return config.IsFeatureInBundle(featureID, bundleID), nil
}

// GetActiveBundles returns all active bundles for an organization
func (s *ActivationService) GetActiveBundles(ctx context.Context, orgID string) ([]string, error) {
	allBundles, err := config.GetAllBundles()
	if err != nil {
		return nil, err
	}

	activeBundles := make([]string, 0)
	for _, bundle := range allBundles {
		hasAccess, err := s.HasBundleAccess(ctx, orgID, bundle)
		if err != nil {
			return nil, err
		}
		if hasAccess {
			activeBundles = append(activeBundles, bundle)
		}
	}

	return activeBundles, nil
}

// GetFeatureActivationHistory returns the activation history for a specific feature
func (s *ActivationService) GetFeatureActivationHistory(ctx context.Context, orgID string, featureID string) ([]*FeatureActivation, error) {
	if !config.IsEnterpriseFeature(featureID) {
		return nil, ErrInvalidFeature
	}

	history, err := s.repository.GetHistory(ctx, featureID, orgID)
	if err != nil {
		return nil, ErrRepositoryFailure
	}

	return history, nil
}

// DeactivateFeature deactivates an enterprise feature for an organization
func (s *ActivationService) DeactivateFeature(ctx context.Context, orgID string, featureID string) error {
	activation, err := s.repository.Get(ctx, featureID, orgID)
	if err != nil {
		if errors.Is(err, ErrFeatureNotFound) {
			return nil // Already deactivated
		}
		return ErrRepositoryFailure
	}

	activation.Active = false
	activation.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, activation); err != nil {
		return ErrRepositoryFailure
	}

	return nil
}

// IsFeatureActive checks if a feature is currently active for an organization
// and validates the license and feature availability
func (s *ActivationService) IsFeatureActive(ctx context.Context, orgID string, featureID string) (bool, error) {
	// Check if feature is valid and available in org's tier
	available, err := s.IsFeatureAvailable(ctx, orgID, featureID)
	if err != nil {
		return false, err
	}
	if !available {
		return false, nil
	}

	// Check if feature is activated
	activation, err := s.repository.Get(ctx, featureID, orgID)
	if err != nil {
		if errors.Is(err, ErrFeatureNotFound) {
			return false, nil
		}
		return false, ErrRepositoryFailure
	}

	// Verify license is still valid
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil {
		return false, ErrLicenseRequired
	}
	if license == nil {
		return false, nil
	}

	return activation.Active, nil
}

// InitiateFeaturePayment creates a payment for a feature or bundle using Mollie
func (s *ActivationService) InitiateFeaturePayment(ctx context.Context, orgID string, itemID string, isBundle bool, billingPeriod string, currency string) (*payment.PaymentResponse, error) {
	// Verify the feature/bundle exists and is available for purchase
	if !s.IsFeatureAvailableForPurchase(ctx, orgID, itemID, isBundle) {
		return nil, ErrInvalidFeature
	}

	// Calculate price
	price, err := s.CalculateFeaturePrice(ctx, orgID, itemID, isBundle, billingPeriod, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	// Create payment request
	paymentReq := payment.PaymentRequest{
		Amount: payment.Amount{
			Currency: currency,
			Value:    price.Amount,
		},
		Description: fmt.Sprintf("PasswordImmunity %s: %s (%s)",
			map[bool]string{true: "Bundle", false: "Feature"}[isBundle],
			itemID,
			billingPeriod,
		),
		RedirectURL: "https://passwordimmunity.com/payment/success",
		WebhookURL:  "https://api.passwordimmunity.com/webhooks/mollie",
		Metadata: payment.PaymentMetadata{
			OrganizationID: orgID,
			Features:       []string{itemID},
			Duration:      billingPeriod,
		},
	}

	// If it's a bundle, move the ID to bundles instead of features
	if isBundle {
		paymentReq.Metadata.Features = nil
		paymentReq.Metadata.Bundles = []string{itemID}
	}

	// Create payment via payment service
	return s.paymentService.ProcessPayment(ctx, paymentReq)
}

// CalculateFeaturePrice calculates the price for a feature or bundle based on billing period and currency
func (s *ActivationService) CalculateFeaturePrice(ctx context.Context, orgID string, itemID string, isBundle bool, billingPeriod string, currency string) (*pricing.Price, error) {
	// Validate inputs
	if !pricing.IsValidBillingPeriod(billingPeriod) {
		return nil, pricing.ErrInvalidBillingPeriod
	}
	if !pricing.IsValidCurrency(currency) {
		return nil, pricing.ErrInvalidCurrency
	}

	// Validate feature/bundle
	if isBundle {
		if !config.IsValidBundle(itemID) {
			return nil, config.ErrInvalidBundle
		}
	} else {
		if !config.IsValidFeature(itemID) {
			return nil, ErrInvalidFeature
		}
	}

	// Get organization's current tier
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil && !errors.Is(err, licensing.ErrNoActiveLicense) {
		return nil, err
	}

	currentTier := "free"
	if license != nil {
		currentTier = license.Tier
	}

	// Calculate price based on tier and billing period
	price, err := pricing.CalculatePrice(itemID, isBundle, currentTier, billingPeriod, currency)
	if err != nil {
		return nil, err
	}

	return price, nil
}

// IsFeatureAvailableForPurchase checks if a feature can be purchased by an organization
// considering their current tier and any feature dependencies
func (s *ActivationService) IsFeatureAvailableForPurchase(ctx context.Context, orgID string, featureID string) (bool, error) {
	if !config.IsValidFeature(featureID) {
		return false, ErrInvalidFeature
	}

	// Check current license
	license, err := s.licenseService.GetActiveLicense(ctx, orgID)
	if err != nil && !errors.Is(err, licensing.ErrNoActiveLicense) {
		return false, err
	}

	// Get required tier for feature
	requiredTier := config.GetFeatureRequiredTier(featureID)

	// Check if current tier allows feature
	currentTier := "free"
	if license != nil {
		currentTier = license.Tier
	}

	if !config.IsTierSufficientForFeature(currentTier, requiredTier) {
		return false, nil
	}

	// Check feature dependencies
	dependencies := config.GetFeatureDependencies(featureID)
	for _, dep := range dependencies {
		active, err := s.IsFeatureActive(ctx, orgID, dep)
		if err != nil || !active {
			return false, nil
		}
	}

	return true, nil
}

// activateFeatureForOrg handles the actual feature activation in the system
func (s *ActivationService) activateFeatureForOrg(ctx context.Context, orgID string, featureID string) error {
	// Check if feature is already activated
	existing, err := s.repository.Get(ctx, featureID, orgID)
	if err != nil && !errors.Is(err, ErrFeatureNotFound) {
		return ErrRepositoryFailure
	}
	if existing != nil && existing.Active {
		return nil // Feature already activated
	}

	activation := &FeatureActivation{
		FeatureID:      featureID,
		OrganizationID: orgID,
		ActivatedAt:    time.Now(),
		Active:         true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repository.Create(ctx, activation); err != nil {
		return ErrRepositoryFailure
	}

	return nil
}
