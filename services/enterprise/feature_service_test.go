package enterprise

import (
    "context"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockRepo struct {
    mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, activation *models.FeatureActivation) error {
    args := m.Called(ctx, activation)
    return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*models.FeatureActivation, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.FeatureActivation), args.Error(1)
}

func (m *mockRepo) GetActiveByOrganization(ctx context.Context, orgID string) ([]*models.FeatureActivation, error) {
    args := m.Called(ctx, orgID)
    return args.Get(0).([]*models.FeatureActivation), args.Error(1)
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
    args := m.Called(ctx, id, status)
    return args.Error(0)
}

func (m *mockRepo) GetByFeature(ctx context.Context, orgID string, featureID string) (*models.FeatureActivation, error) {
    args := m.Called(ctx, orgID, featureID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.FeatureActivation), args.Error(1)
}

func (m *mockRepo) GetByBundle(ctx context.Context, orgID string, bundleID string) (*models.FeatureActivation, error) {
    args := m.Called(ctx, orgID, bundleID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.FeatureActivation), args.Error(1)
}

func TestFeatureService(t *testing.T) {
    ctx := context.Background()
    repo := new(mockRepo)
    service := NewFeatureService(repo)

    t.Run("ActivateFeature", func(t *testing.T) {
        orgID := "test-org"
        featureID := "advanced_sso"
        paymentID := "test-payment"
        duration := 30 * 24 * time.Hour

        repo.On("Create", ctx, mock.MatchedBy(func(activation *models.FeatureActivation) bool {
            return activation.OrganizationID == orgID &&
                *activation.FeatureID == featureID &&
                activation.PaymentID == paymentID
        })).Return(nil)

        err := service.ActivateFeature(ctx, orgID, featureID, paymentID, duration)
        assert.NoError(t, err)
        repo.AssertExpectations(t)
    })

    t.Run("ActivateBundle", func(t *testing.T) {
        orgID := "test-org"
        bundleID := "enterprise"
        paymentID := "test-payment"
        duration := 30 * 24 * time.Hour

        repo.On("Create", ctx, mock.MatchedBy(func(activation *models.FeatureActivation) bool {
            return activation.OrganizationID == orgID &&
                *activation.BundleID == bundleID &&
                activation.PaymentID == paymentID
        })).Return(nil)

        err := service.ActivateBundle(ctx, orgID, bundleID, paymentID, duration)
        assert.NoError(t, err)
        repo.AssertExpectations(t)
    })

    t.Run("IsFeatureActive", func(t *testing.T) {
        orgID := "test-org"
        featureID := "advanced_sso"
        activation := &models.FeatureActivation{
            Status:    "active",
            ExpiresAt: time.Now().Add(24 * time.Hour),
        }

        repo.On("GetByFeature", ctx, orgID, featureID).Return(activation, nil)

        active, err := service.IsFeatureActive(ctx, orgID, featureID)
        assert.NoError(t, err)
        assert.True(t, active)
        repo.AssertExpectations(t)
    })

    t.Run("GetActiveFeatures", func(t *testing.T) {
        orgID := "test-org"
        featureID := "advanced_sso"
        bundleID := "enterprise"

        activations := []*models.FeatureActivation{
            {
                FeatureID: &featureID,
                Status:    "active",
                ExpiresAt: time.Now().Add(24 * time.Hour),
            },
            {
                BundleID:  &bundleID,
                Status:    "active",
                ExpiresAt: time.Now().Add(24 * time.Hour),
            },
        }

        repo.On("GetActiveByOrganization", ctx, orgID).Return(activations, nil)

        features, err := service.GetActiveFeatures(ctx, orgID)
        assert.NoError(t, err)
        assert.Contains(t, features, featureID)
        assert.True(t, len(features) > 1) // Should include bundle features
        repo.AssertExpectations(t)
    })

    t.Run("DeactivateFeature", func(t *testing.T) {
        orgID := "test-org"
        featureID := "advanced_sso"
        activation := &models.FeatureActivation{
            ID:        1,
            Status:    "active",
            ExpiresAt: time.Now().Add(24 * time.Hour),
        }

        repo.On("GetByFeature", ctx, orgID, featureID).Return(activation, nil)
        repo.On("UpdateStatus", ctx, activation.ID, "cancelled").Return(nil)

        err := service.DeactivateFeature(ctx, orgID, featureID)
        assert.NoError(t, err)
        repo.AssertExpectations(t)
    })
}
