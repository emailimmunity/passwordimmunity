package payment

import (
	"fmt"
	"github.com/emailimmunity/passwordimmunity/config"
)

func validateEnterpriseFeatures(features []string) error {
	for _, feature := range features {
		if !config.IsValidEnterpriseFeature(feature) {
			return fmt.Errorf("invalid enterprise feature: %s", feature)
		}
	}
	return nil
}

func validateFeatureBundles(bundles []string) error {
	for _, bundle := range bundles {
		if !config.IsValidFeatureBundle(bundle) {
			return fmt.Errorf("invalid feature bundle: %s", bundle)
		}
	}
	return nil
}

func validateFeatureDependencies(features []string) error {
	for _, feature := range features {
		dependencies, err := config.GetFeatureDependencies(feature)
		if err != nil {
			return fmt.Errorf("failed to get dependencies for feature %s: %w", feature, err)
		}

		for _, dep := range dependencies {
			if !config.IsFeatureEnabled(dep) {
				return fmt.Errorf("missing required dependency %s for feature %s", dep, feature)
			}
		}
	}
	return nil
}
