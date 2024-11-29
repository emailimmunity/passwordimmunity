package enterprise

import (
	"context"
	"testing"
	"time"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
	RepositoryInterface
}

func (m *MockRepository) StoreActivation(ctx context.Context, activation FeatureActivation) error {
	args := m.Called(ctx, activation)
	return args.Error(0)
}

func (m *MockRepository) GetActivation(ctx context.Context, orgID, featureID string) (*FeatureActivation, error) {
	args := m.Called(ctx, orgID, featureID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FeatureActivation), args.Error(1)
}

type MockLogger struct {
	mock.Mock
	Logger
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestFeatureManager_ActivateFeature_Tier(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{}
	logger := &MockLogger{}
	pricing := NewPricingManager("USD")
	manager := NewFeatureManager(logger, repo, pricing)

	activation := FeatureActivation{
		OrganizationID: "org1",
		TierID:         "enterprise",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(199.99),
		IsYearly:       true,
	}

	// Expect storing activation for each feature in the enterprise tier
	repo.On("StoreActivation", ctx, mock.AnythingOfType("FeatureActivation")).Return(nil).Times(7)
	logger.On("Info", mock.Anything, mock.Anything).Return(nil)

	err := manager.ActivateFeature(ctx, activation)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestFeatureManager_ActivateFeature_InvalidTier(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{}
	logger := &MockLogger{}
	pricing := NewPricingManager("USD")
	manager := NewFeatureManager(logger, repo, pricing)

	activation := FeatureActivation{
		OrganizationID: "org1",
		TierID:         "nonexistent",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(199.99),
		IsYearly:       true,
	}

	logger.On("Error", mock.Anything, mock.Anything).Return(nil)

	err := manager.ActivateFeature(ctx, activation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tier")

	repo.AssertNotCalled(t, "StoreActivation")
}

func TestFeatureManager_ActivateFeature_InvalidPrice(t *testing.T) {
	ctx := context.Background()
	repo := &MockRepository{}
	logger := &MockLogger{}
	pricing := NewPricingManager("USD")
	manager := NewFeatureManager(logger, repo, pricing)

	activation := FeatureActivation{
		OrganizationID: "org1",
		TierID:         "enterprise",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(150.00),
		IsYearly:       true,
	}

	logger.On("Error", mock.Anything, mock.Anything).Return(nil)

	err := manager.ActivateFeature(ctx, activation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price mismatch")

	repo.AssertNotCalled(t, "StoreActivation")
}
