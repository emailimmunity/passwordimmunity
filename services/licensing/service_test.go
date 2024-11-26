package licensing

import (
    "context"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) GetActiveLicenseByOrganization(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.License), args.Error(1)
}

func (m *mockRepository) GetExpiredLicenses(ctx context.Context) ([]models.License, error) {
    args := m.Called(ctx)
    return args.Get(0).([]models.License), args.Error(1)
}

func (m *mockRepository) UpdateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func (m *mockRepository) CreateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func (m *mockRepository) DeactivateOrganizationLicenses(ctx context.Context, orgID uuid.UUID) error {
    args := m.Called(ctx, orgID)
    return args.Error(0)
}

func TestGetActiveLicense(t *testing.T) {
    mockRepo := new(mockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()
    orgID := uuid.New()

    validLicense := &models.License{
        OrganizationID: orgID,
        Type:           "enterprise",
        Status:         "active",
        ExpiresAt:      time.Now().Add(24 * time.Hour),
        Features:       []string{"sso", "api_access"},
    }

    mockRepo.On("GetActiveLicenseByOrganization", ctx, orgID).Return(validLicense, nil)

    license, err := service.GetActiveLicense(ctx, orgID)
    assert.NoError(t, err)
    assert.Equal(t, validLicense, license)
}

func TestHasFeature(t *testing.T) {
    mockRepo := new(mockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()
    orgID := uuid.New()

    tests := []struct {
        name     string
        license  *models.License
        feature  string
        expected bool
    }{
        {
            name: "Has feature",
            license: &models.License{
                Features: []string{"sso", "api_access"},
            },
            feature:  "sso",
            expected: true,
        },
        {
            name: "Does not have feature",
            license: &models.License{
                Features: []string{"api_access"},
            },
            feature:  "sso",
            expected: false,
        },
        {
            name:     "No license",
            license:  nil,
            feature:  "sso",
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo.On("GetActiveLicenseByOrganization", ctx, orgID).Return(tt.license, nil).Once()

            hasFeature, err := service.HasFeature(ctx, orgID, tt.feature)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, hasFeature)
        })
    }
}

func TestCheckLicenseExpiry(t *testing.T) {
    mockRepo := new(mockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()

    expiredLicenses := []models.License{
        {
            ID:        uuid.New(),
            Status:    "active",
            ExpiresAt: time.Now().Add(-24 * time.Hour),
        },
    }

    mockRepo.On("GetExpiredLicenses", ctx).Return(expiredLicenses, nil)
    mockRepo.On("UpdateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

    err := service.CheckLicenseExpiry(ctx)
    assert.NoError(t, err)

    mockRepo.AssertCalled(t, "UpdateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
        return license.Status == "expired"
    }))
}

func TestActivateLicense(t *testing.T) {
    mockRepo := new(mockRepository)
    service := NewService(mockRepo)
    ctx := context.Background()
    orgID := uuid.New()
    validUntil := time.Now().Add(24 * time.Hour)

    tests := []struct {
        name        string
        licenseType string
        setupMock   func()
        wantErr     bool
    }{
        {
            name:        "successful activation",
            licenseType: "enterprise",
            setupMock: func() {
                mockRepo.On("DeactivateOrganizationLicenses", ctx, orgID).Return(nil)
                mockRepo.On("CreateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
                    return license.OrganizationID == orgID &&
                           license.Type == "enterprise" &&
                           len(license.Features) > 0
                })).Return(nil)
            },
            wantErr: false,
        },
        {
            name:        "deactivation fails",
            licenseType: "premium",
            setupMock: func() {
                mockRepo.On("DeactivateOrganizationLicenses", ctx, orgID).Return(assert.AnError)
            },
            wantErr: true,
        },
        {
            name:        "creation fails",
            licenseType: "premium",
            setupMock: func() {
                mockRepo.On("DeactivateOrganizationLicenses", ctx, orgID).Return(nil)
                mockRepo.On("CreateLicense", ctx, mock.Anything).Return(assert.AnError)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo = new(mockRepository)
            service = NewService(mockRepo)
            tt.setupMock()

            err := service.ActivateLicense(ctx, orgID, tt.licenseType, validUntil)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            mockRepo.AssertExpectations(t)
        })
    }
}
