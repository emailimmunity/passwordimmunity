package enterprise

import (
	"context"
	"sync"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/logger"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

// featureActivationService implements FeatureActivationService interface
type featureActivationService struct {
	repo    FeatureActivationRepository
	payment payment.Service
	logger  logger.Logger
	mu      sync.RWMutex
}

// NewFeatureActivationService creates a new instance of FeatureActivationService
func NewFeatureActivationService(repo FeatureActivationRepository, payment payment.Service, logger logger.Logger) FeatureActivationService {
	return &featureActivationService{
		repo:    repo,
		payment: payment,
		logger:  logger,
	}
}

func (s *featureActivationService) ActivateFeature(ctx context.Context, featureID, organizationID string, duration time.Duration) (*FeatureActivation, error) {
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	// Validate payment for feature activation
	if err := s.payment.ValidateFeaturePayment(ctx, organizationID, featureID); err != nil {
		s.logger.Error("Payment validation failed for feature activation",
			"error", err,
			"organization_id", organizationID,
			"feature_id", featureID,
		)
		return nil, err
	}

	// Check if feature is already active
	existing, err := s.repo.Get(ctx, featureID, organizationID)
	if err != nil && err != ErrFeatureNotFound {
		return nil, err
	}
	if existing != nil && existing.Active && time.Now().Before(existing.ExpiresAt) {
		return nil, ErrFeatureAlreadyActive
	}

	activation := &FeatureActivation{
		FeatureID:      featureID,
		OrganizationID: organizationID,
		ActivatedAt:    time.Now(),
		ExpiresAt:      time.Now().Add(duration),
		Active:         true,
	}

	if err := s.repo.Create(ctx, activation); err != nil {
		s.logger.Error("Failed to create feature activation",
			"error", err,
			"organization_id", organizationID,
			"feature_id", featureID,
		)
		return nil, err
	}

	s.logger.Info("Feature activated successfully",
		"organization_id", organizationID,
		"feature_id", featureID,
		"expires_at", activation.ExpiresAt,
	)

	return activation, nil
}

func (s *featureActivationService) DeactivateFeature(ctx context.Context, featureID, organizationID string) error {
	return s.repo.Delete(ctx, featureID, organizationID)
}

func (s *featureActivationService) IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error) {
	activation, err := s.repo.Get(ctx, featureID, organizationID)
	if err == ErrFeatureNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return activation.Active && time.Now().Before(activation.ExpiresAt), nil
}

func (s *featureActivationService) GetActiveFeatures(ctx context.Context, organizationID string) ([]FeatureActivation, error) {
	return s.repo.GetAllActive(ctx, organizationID)
}

func (s *featureActivationService) ExtendFeatureActivation(ctx context.Context, featureID, organizationID string, duration time.Duration) (*FeatureActivation, error) {
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	// Validate payment for extension
	if err := s.payment.ValidateFeatureExtension(ctx, organizationID, featureID, duration); err != nil {
		s.logger.Error("Payment validation failed for feature extension",
			"error", err,
			"organization_id", organizationID,
			"feature_id", featureID,
		)
		return nil, err
	}

	activation, err := s.repo.Get(ctx, featureID, organizationID)
	if err != nil {
		s.logger.Error("Failed to get feature activation for extension",
			"error", err,
			"organization_id", organizationID,
			"feature_id", featureID,
		)
		return nil, err
	}

	activation.ExpiresAt = activation.ExpiresAt.Add(duration)
	if err := s.repo.Update(ctx, activation); err != nil {
		s.logger.Error("Failed to update feature activation",
			"error", err,
			"organization_id", organizationID,
			"feature_id", featureID,
		)
		return nil, err
	}

	s.logger.Info("Feature activation extended successfully",
		"organization_id", organizationID,
		"feature_id", featureID,
		"new_expiry", activation.ExpiresAt,
	)

	return activation, nil
}

func (s *featureActivationService) GetFeatureActivation(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error) {
	return s.repo.Get(ctx, featureID, organizationID)
}
