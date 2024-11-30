package featureflag

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockLicenseService struct {
    mock.Mock
}

func (m *mockLicenseService) GetActiveLicense(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, orgID)
    return args.Get(0).(*models.License), args.Error(1)
}

func (m *mockLicenseService) GetLicenseType(ctx context.Context, orgID uuid.UUID) (string, error) {
    args := m.Called(ctx, orgID)
    return args.String(0), args.Error(1)
}

func TestIsFeatureEnabled(t *testing.T) {
    tests := []struct {
        name        string
        licenseType string
        feature     string
        expected    bool
    }{
        {"Free feature on free tier", "free", "basic_vault", true},
        {"Premium feature on free tier", "free", "advanced_2fa", false},
        {"Enterprise feature on free tier", "free", "sso", false},
        {"Premium feature on premium tier", "premium", "advanced_2fa", true},
        {"Enterprise feature on premium tier", "premium", "sso", false},
        {"Enterprise feature on enterprise tier", "enterprise", "sso", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockLS := new(mockLicenseService)
            service := NewService(mockLS)
            ctx := context.Background()
            orgID := uuid.New()

            mockLS.On("GetLicenseType", ctx, orgID).Return(tt.licenseType, nil)

            result, err := service.IsFeatureEnabled(ctx, orgID, tt.feature)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestGetAvailableFeatures(t *testing.T) {
    tests := []struct {
        name        string
        licenseType string
        minFeatures int
    }{
        {"Free tier features", "free", 3},
        {"Premium tier features", "premium", 8},
        {"Enterprise tier features", "enterprise", 18},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockLS := new(mockLicenseService)
            service := NewService(mockLS)
            ctx := context.Background()
            orgID := uuid.New()

            mockLS.On("GetLicenseType", ctx, orgID).Return(tt.licenseType, nil)

            features, err := service.GetAvailableFeatures(ctx, orgID)
            assert.NoError(t, err)
            assert.GreaterOrEqual(t, len(features), tt.minFeatures)
        })
    }
}
