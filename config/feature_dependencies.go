package config

import (
	"fmt"
)

// FeatureDependencyMap defines dependencies between features
var FeatureDependencyMap = map[string][]string{
	"advanced_reporting": {"basic_reporting"},
	"custom_roles":      {"basic_roles"},
	"advanced_sso":      {"basic_auth"},
	"multi_tenant":      {"custom_roles", "advanced_sso"},
}

// GetFeatureDependencies returns the list of required features for a given feature
func GetFeatureDependencies(featureID string) ([]string, error) {
	deps, exists := FeatureDependencyMap[featureID]
	if !exists {
		return nil, fmt.Errorf("unknown feature: %s", featureID)
	}
	return deps, nil
}

// ValidateFeatureDependencies checks if all dependencies for a feature are satisfied
func ValidateFeatureDependencies(featureID string, enabledFeatures []string) error {
	deps, err := GetFeatureDependencies(featureID)
	if err != nil {
		return err
	}

	enabledMap := make(map[string]bool)
	for _, f := range enabledFeatures {
		enabledMap[f] = true
	}

	for _, dep := range deps {
		if !enabledMap[dep] {
			return fmt.Errorf("missing required feature: %s", dep)
		}
	}

	return nil
}
