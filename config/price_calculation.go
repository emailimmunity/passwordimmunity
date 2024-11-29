package config

// CalculateTotalPrice calculates the total price for a combination of features and bundles
func CalculateTotalPrice(features []string, bundles []string, currency string) float64 {
	total := 0.0

	// Calculate bundle prices
	for _, bundleID := range bundles {
		total += GetBundlePrice(bundleID, currency)
	}

	// Calculate individual feature prices
	for _, featureID := range features {
		// Skip if feature is already included in any bundle
		if !isFeatureInBundles(featureID, bundles) {
			total += GetFeaturePrice(featureID, currency)
		}
	}

	return total
}

// isFeatureInBundles checks if a feature is included in any of the specified bundles
func isFeatureInBundles(featureID string, bundles []string) bool {
	for _, bundleID := range bundles {
		features := GetFeaturesInBundle(bundleID)
		for _, bundleFeature := range features {
			if bundleFeature == featureID {
				return true
			}
		}
	}
	return false
}

// ValidateFeatureAndBundleCombination checks if the feature and bundle combination is valid
func ValidateFeatureAndBundleCombination(features []string, bundles []string) bool {
	// Validate all features exist
	for _, featureID := range features {
		if !IsEnterpriseFeature(featureID) {
			return false
		}
	}

	// Validate all bundles exist
	for _, bundleID := range bundles {
		if _, exists := GetBundleByID(bundleID); !exists {
			return false
		}
	}

	// Validate feature dependencies
	for _, featureID := range features {
		feature, _ := GetFeatureByID(featureID)
		for _, dep := range feature.Dependencies {
			if !isFeatureIncluded(dep, features, bundles) {
				return false
			}
		}
	}

	return true
}

// isFeatureIncluded checks if a feature is included either directly or via bundles
func isFeatureIncluded(featureID string, features []string, bundles []string) bool {
	// Check direct inclusion
	for _, f := range features {
		if f == featureID {
			return true
		}
	}

	// Check bundle inclusion
	return isFeatureInBundles(featureID, bundles)
}
