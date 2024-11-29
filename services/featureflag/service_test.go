package featureflag

import (
	"context"
	"testing"
	"time"

	"github.com/emailimmunity/passwordimmunity/config"
)

func TestFeatureManager(t *testing.T) {
	fm := NewFeatureManager()
	ctx := context.Background()
	orgID := "test_org"

	t.Run("ActivateFeature", func(t *testing.T) {
		feature := "advanced_sso"
		err := fm.ActivateFeature(orgID, feature)
		if err != nil {
			t.Fatalf("Failed to activate feature: %v", err)
		}

		active := fm.IsFeatureActive(ctx, orgID, feature)
		if !active {
			t.Error("Expected feature to be active")
		}
	})

	t.Run("DeactivateFeature", func(t *testing.T) {
		feature := "advanced_sso"
		err := fm.DeactivateFeature(orgID, feature)
		if err != nil {
			t.Fatalf("Failed to deactivate feature: %v", err)
		}

		active := fm.IsFeatureActive(ctx, orgID, feature)
		if active {
			t.Error("Expected feature to be inactive")
		}
	})

	t.Run("GracePeriod", func(t *testing.T) {
		feature := "multi_tenant"
		featureConfig := config.FeatureConfig{
			GracePeriod: 7,
		}

		err := fm.ActivateFeatureWithGracePeriod(orgID, feature, featureConfig)
		if err != nil {
			t.Fatalf("Failed to activate feature with grace period: %v", err)
		}

		// Check feature is active during grace period
		active := fm.IsFeatureActive(ctx, orgID, feature)
		if !active {
			t.Error("Expected feature to be active during grace period")
		}

		// Simulate grace period expiration
		fm.features[orgID][feature] = time.Now().Add(-8 * 24 * time.Hour)

		// Check feature is inactive after grace period
		active = fm.IsFeatureActive(ctx, orgID, feature)
		if active {
			t.Error("Expected feature to be inactive after grace period")
		}
	})

	t.Run("ListActiveFeatures", func(t *testing.T) {
		// Activate multiple features
		features := []string{"advanced_sso", "multi_tenant"}
		for _, feature := range features {
			err := fm.ActivateFeature(orgID, feature)
			if err != nil {
				t.Fatalf("Failed to activate feature %s: %v", feature, err)
			}
		}

		activeFeatures := fm.ListActiveFeatures(ctx, orgID)
		if len(activeFeatures) != len(features) {
			t.Errorf("Expected %d active features, got %d", len(features), len(activeFeatures))
		}

		// Verify all activated features are in the list
		for _, feature := range features {
			found := false
			for _, active := range activeFeatures {
				if active == feature {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected feature %s to be in active features list", feature)
			}
		}
	})
}
