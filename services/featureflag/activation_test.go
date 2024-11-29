package featureflag

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/emailimmunity/passwordimmunity/config"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) ActivateFeature(ctx context.Context, featureID string) error {
	args := m.Called(ctx, featureID)
	return args.Error(0)
}

func TestFeatureManager_ActivateFeature(t *testing.T) {
	// Setup test config
	config.FeatureTiers = map[string]config.Tier{
		"enterprise": {
			Features: []string{"advanced_sso", "multi_tenant"},
		},
	}

	config.FeatureBundles = map[string]config.Bundle{
		"security": {
			Features: []string{"advanced_audit", "custom_policies"},
		},
	}

	config.Features = map[string]config.Feature{
		"advanced_sso":     {},
		"multi_tenant":     {},
		"advanced_audit":   {},
		"custom_policies": {},
	}

	tests := []struct {
		name      string
		activation FeatureActivation
		setupMock func(*MockRepository)
		wantErr   bool
	}{
		{
			name: "successful tier activation",
			activation: FeatureActivation{
				TierID: "enterprise",
			},
			setupMock: func(m *MockRepository) {
				m.On("ActivateFeature", mock.Anything, "advanced_sso").Return(nil)
				m.On("ActivateFeature", mock.Anything, "multi_tenant").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful bundle activation",
			activation: FeatureActivation{
				BundleID: "security",
			},
			setupMock: func(m *MockRepository) {
				m.On("ActivateFeature", mock.Anything, "advanced_audit").Return(nil)
				m.On("ActivateFeature", mock.Anything, "custom_policies").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful single feature activation",
			activation: FeatureActivation{
				FeatureID: "advanced_sso",
			},
			setupMock: func(m *MockRepository) {
				m.On("ActivateFeature", mock.Anything, "advanced_sso").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid tier ID",
			activation: FeatureActivation{
				TierID: "invalid_tier",
			},
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
		},
		{
			name: "invalid bundle ID",
			activation: FeatureActivation{
				BundleID: "invalid_bundle",
			},
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
		},
		{
			name: "invalid feature ID",
			activation: FeatureActivation{
				FeatureID: "invalid_feature",
			},
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			fm := &FeatureManager{
				repository: mockRepo,
			}

			err := fm.ActivateFeature(context.Background(), tt.activation)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
