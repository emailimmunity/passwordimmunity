package enterprise

import (
	"context"
	"errors"
	"time"
)

// FeatureActivation represents the activation status of an enterprise feature
type FeatureActivation struct {
	FeatureID      string    `json:"feature_id"`
	OrganizationID string    `json:"organization_id"`
	ActivatedAt    time.Time `json:"activated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Active         bool      `json:"active"`
}

// FeatureActivationService defines the interface for managing enterprise feature activation
type FeatureActivationService interface {
	// ActivateFeature activates a feature for an organization
	ActivateFeature(ctx context.Context, featureID, organizationID string, duration time.Duration) (*FeatureActivation, error)

	// DeactivateFeature deactivates a feature for an organization
	DeactivateFeature(ctx context.Context, featureID, organizationID string) error

	// IsFeatureActive checks if a feature is active for an organization
	IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error)

	// GetActiveFeatures returns all active features for an organization
	GetActiveFeatures(ctx context.Context, organizationID string) ([]FeatureActivation, error)

	// ExtendFeatureActivation extends the activation period for a feature
	ExtendFeatureActivation(ctx context.Context, featureID, organizationID string, duration time.Duration) (*FeatureActivation, error)

	// GetFeatureActivation gets the activation status for a specific feature
	GetFeatureActivation(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error)
}

// ErrFeatureNotFound indicates that the requested feature was not found
var ErrFeatureNotFound = errors.New("feature not found")

// ErrFeatureAlreadyActive indicates that the feature is already active
var ErrFeatureAlreadyActive = errors.New("feature already active")

// ErrFeatureNotActive indicates that the feature is not active
var ErrFeatureNotActive = errors.New("feature not active")

// ErrInvalidDuration indicates that the provided duration is invalid
var ErrInvalidDuration = errors.New("invalid duration")

// ErrFeatureExpired indicates that the feature activation has expired
var ErrFeatureExpired = errors.New("feature activation expired")
