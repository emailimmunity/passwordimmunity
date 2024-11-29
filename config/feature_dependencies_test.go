package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetFeatureDependencies(t *testing.T) {
	tests := []struct {
		name          string
		featureID     string
		expectedDeps  []string
		expectError   bool
		errorContains string
	}{
		{
			name:         "multi_tenant dependencies",
			featureID:    "multi_tenant",
			expectedDeps: []string{"custom_roles", "advanced_sso"},
			expectError:  false,
		},
		{
			name:          "unknown feature",
			featureID:     "nonexistent_feature",
			expectedDeps:  nil,
			expectError:   true,
			errorContains: "unknown feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, err := GetFeatureDependencies(tt.featureID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDeps, deps)
			}
		})
	}
}

func TestValidateFeatureDependencies(t *testing.T) {
	tests := []struct {
		name            string
		featureID       string
		enabledFeatures []string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "all dependencies satisfied",
			featureID:       "multi_tenant",
			enabledFeatures: []string{"custom_roles", "advanced_sso", "basic_auth"},
			expectError:     false,
		},
		{
			name:            "missing dependency",
			featureID:       "multi_tenant",
			enabledFeatures: []string{"custom_roles"},
			expectError:     true,
			errorContains:   "missing required feature",
		},
		{
			name:            "unknown feature",
			featureID:       "nonexistent_feature",
			enabledFeatures: []string{},
			expectError:     true,
			errorContains:   "unknown feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFeatureDependencies(tt.featureID, tt.enabledFeatures)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
