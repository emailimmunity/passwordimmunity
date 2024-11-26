package testutil

import (
    "context"
    "time"

    "github.com/emailimmunity/passwordimmunity/config"
    "github.com/emailimmunity/passwordimmunity/services/licensing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
)

// MockLicenseService is a mock implementation of the licensing.Service interface
type MockLicenseService struct {
    mock.Mock
}

func (m *MockLicenseService) GetActiveLicense(ctx context.Context, orgID uuid.UUID) (*licensing.License, error) {
    args := m.Called(ctx, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*licensing.License), args.Error(1)
}

func (m *MockLicenseService) GetLicenseType(ctx context.Context, orgID uuid.UUID) (string, error) {
    args := m.Called(ctx, orgID)
    return args.String(0), args.Error(1)
}

// CreateTestLicense creates a test license with the specified type
func CreateTestLicense(licenseType string, validDays int) *licensing.License {
    return &licensing.License{
        ID:         uuid.New(),
        Type:       licenseType,
        Status:     "active",
        ValidUntil: time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
        Features:   GetFeaturesForTier(licenseType),
    }
}

// GetFeaturesForTier returns a list of feature names available for the specified tier
func GetFeaturesForTier(tier string) []string {
    features := config.GetFeaturesByTier(tier)
    result := make([]string, len(features))
    for i, f := range features {
        result[i] = f.Name
    }
    return result
}

// SetupMockLicenseService configures a mock license service with common expectations
func SetupMockLicenseService(orgID uuid.UUID, licenseType string) *MockLicenseService {
    mockLS := new(MockLicenseService)
    license := CreateTestLicense(licenseType, 30)

    mockLS.On("GetActiveLicense", mock.Anything, orgID).Return(license, nil)
    mockLS.On("GetLicenseType", mock.Anything, orgID).Return(licenseType, nil)

    return mockLS
}
