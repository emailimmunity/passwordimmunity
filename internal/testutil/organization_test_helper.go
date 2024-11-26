package testutil

import (
    "context"
    "time"

    "github.com/emailimmunity/passwordimmunity/services/organization"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
)

type MockOrganizationService struct {
    mock.Mock
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*organization.Organization), args.Error(1)
}

func CreateTestOrganization(tier string) *organization.Organization {
    return &organization.Organization{
        ID:        uuid.New(),
        Name:      "Test Organization",
        Tier:      tier,
        Status:    "active",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func CreateTestUser(orgID uuid.UUID, role string) *organization.User {
    return &organization.User{
        ID:             uuid.New(),
        OrganizationID: orgID,
        Email:          "test@example.com",
        Role:           role,
        Status:         "active",
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
}

func SetupMockOrganizationService(tier string) *MockOrganizationService {
    mockOS := new(MockOrganizationService)
    org := CreateTestOrganization(tier)

    mockOS.On("GetOrganization", mock.Anything, mock.Anything).Return(org, nil)

    return mockOS
}

func GetTestRoles() []string {
    return []string{
        "owner",
        "admin",
        "manager",
        "user",
    }
}

func GetTestEnterpriseRoles() []string {
    return []string{
        "owner",
        "admin",
        "manager",
        "user",
        "auditor",
        "readonly",
        "custom",
    }
}
