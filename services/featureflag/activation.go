package featureflag

import (
	"context"
	"errors"
	"github.com/emailimmunity/passwordimmunity/config"
)

// FeatureActivation represents a request to activate features
type FeatureActivation struct {
	TierID    string
	FeatureID string
	BundleID  string
}

// ActivateFeature handles feature activation based on tier, bundle, or individual feature
func (m *FeatureManager) ActivateFeature(ctx context.Context, activation FeatureActivation) error {
	switch {
	case activation.TierID != "":
		return m.activateTierFeatures(ctx, activation.TierID)
	case activation.BundleID != "":
		return m.activateBundleFeatures(ctx, activation.BundleID)
	case activation.FeatureID != "":
		return m.activateSingleFeature(ctx, activation.FeatureID)
	default:
		return errors.New("invalid activation request: must specify tier, bundle, or feature")
	}
}

func (m *FeatureManager) activateTierFeatures(ctx context.Context, tierID string) error {
	features, ok := config.FeatureTiers[tierID]
	if !ok {
		return errors.New("invalid tier ID")
	}

	for _, featureID := range features.Features {
		if err := m.activateSingleFeature(ctx, featureID); err != nil {
			return err
		}
	}

	return nil
}

func (m *FeatureManager) activateBundleFeatures(ctx context.Context, bundleID string) error {
	bundle, ok := config.FeatureBundles[bundleID]
	if !ok {
		return errors.New("invalid bundle ID")
	}

	for _, featureID := range bundle.Features {
		if err := m.activateSingleFeature(ctx, featureID); err != nil {
			return err
		}
	}

	return nil
}

func (m *FeatureManager) activateSingleFeature(ctx context.Context, featureID string) error {
	if !m.isValidFeature(featureID) {
		return errors.New("invalid feature ID")
	}

	return m.repository.ActivateFeature(ctx, featureID)
}

func (m *FeatureManager) isValidFeature(featureID string) bool {
	_, ok := config.Features[featureID]
	return ok
}
