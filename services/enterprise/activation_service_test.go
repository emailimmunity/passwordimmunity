package enterprise

import (
	"context"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) ProcessPayment(ctx context.Context, req payment.PaymentRequest) (*payment.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if resp := args.Get(0); resp != nil {
		return resp.(*payment.PaymentResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentService) VerifyPayment(ctx context.Context, paymentID string) (*payment.PaymentResponse, error) {
	args := m.Called(ctx, paymentID)
	if resp := args.Get(0); resp != nil {
		return resp.(*payment.PaymentResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, activation *FeatureActivation) error {
	args := m.Called(ctx, activation)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, featureID, organizationID string) error {
	args := m.Called(ctx, featureID, organizationID)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error) {
	args := m.Called(ctx, featureID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FeatureActivation), args.Error(1)
}

func (m *MockRepository) GetAllActive(ctx context.Context, organizationID string) ([]FeatureActivation, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]FeatureActivation), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, activation *FeatureActivation) error {
	args := m.Called(ctx, activation)
	return args.Error(0)
}

func (m *MockRepository) GetHistory(ctx context.Context, featureID, orgID string) ([]*FeatureActivation, error) {
	args := m.Called(ctx, featureID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*FeatureActivation), args.Error(1)
}

type MockLicenseService struct {
	mock.Mock
}

func (m *MockLicenseService) GetActiveLicense(ctx context.Context, orgID string) (*licensing.License, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*licensing.License), args.Error(1)
}

func TestActivationService_ActivateFeature(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name      string
		orgID     string
		featureID string
		license   *licensing.License
		wantErr   error
		setup     func()
	}{
		{
			name:      "valid enterprise feature activation",
			orgID:     "org1",
			featureID: "advanced_sso",
			license: &licensing.License{
				Tier: "enterprise",
			},
			wantErr: nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(activation *FeatureActivation) bool {
					return activation.FeatureID == "advanced_sso" && activation.OrganizationID == "org1"
				})).Return(nil)
			},
		},
		{
			name:      "invalid feature",
			orgID:     "org1",
			featureID: "nonexistent",
			wantErr:   ErrInvalidFeature,
			setup:     func() {},
		},
		{
			name:      "no license",
			orgID:     "org2",
			featureID: "advanced_sso",
			wantErr:   ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
		{
			name:      "feature not in tier",
			orgID:     "org3",
			featureID: "advanced_sso",
			license: &licensing.License{
				Tier: "free",
			},
			wantErr: ErrFeatureNotAllowed,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org3").
					Return(&licensing.License{Tier: "free"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.ActivateFeature(context.Background(), tt.orgID, tt.featureID)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestActivationService_ActivateBundle(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name     string
		orgID    string
		bundleID string
		wantErr  bool
		setup    func()
	}{
		{
			name:     "valid bundle activation",
			orgID:    "org1",
			bundleID: "security",
			wantErr:  false,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil).Times(3)
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(activation *FeatureActivation) bool {
					return activation.OrganizationID == "org1" && activation.Active == true
				})).Return(nil).Times(3)
			},
		},
		{
			name:     "invalid bundle",
			orgID:    "org1",
			bundleID: "nonexistent",
			wantErr:  true,
			setup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.ActivateBundle(context.Background(), tt.orgID, tt.bundleID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestActivationService_HasBundleAccess(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name      string
		orgID     string
		bundleID  string
		wantHas   bool
		wantErr   error
		setup     func()
	}{
		{
			name:     "has complete bundle access",
			orgID:    "org1",
			bundleID: "security",
			wantHas:  true,
			wantErr:  nil,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.Anything, "org1").
					Return(&FeatureActivation{Active: true}, nil).Times(3)
			},
		},
		{
			name:     "missing some features",
			orgID:    "org2",
			bundleID: "security",
			wantHas:  false,
			wantErr:  nil,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.Anything, "org2").
					Return(&FeatureActivation{Active: false}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			has, err := service.HasBundleAccess(context.Background(), tt.orgID, tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantHas, has)
		})
	}
}

func TestActivationService_GetActiveFeatures(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name         string
		orgID        string
		wantFeatures []string
		wantErr      error
		setup        func()
	}{
		{
			name:         "has active features",
			orgID:        "org1",
			wantFeatures: []string{"advanced_sso", "audit_logs"},
			wantErr:      nil,
			setup: func() {
				mockRepo.On("GetAllActive", mock.Anything, "org1").
					Return([]FeatureActivation{
						{FeatureID: "advanced_sso", Active: true},
						{FeatureID: "audit_logs", Active: true},
					}, nil)
			},
		},
		{
			name:         "no active features",
			orgID:        "org2",
			wantFeatures: []string{},
			wantErr:      nil,
			setup: func() {
				mockRepo.On("GetAllActive", mock.Anything, "org2").
					Return([]FeatureActivation{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			features, err := service.GetActiveFeatures(context.Background(), tt.orgID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantFeatures, features)
		})
	}
}

func TestActivationService_GetActiveBundles(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name         string
		orgID        string
		wantBundles  []string
		wantErr      error
		setup        func()
	}{
		{
			name:         "has active bundles",
			orgID:        "org1",
			wantBundles:  []string{"security", "compliance"},
			wantErr:      nil,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.MatchedBy(func(featureID string) bool {
					return featureID == "feature1" || featureID == "feature2"
				}), "org1").Return(&FeatureActivation{Active: true}, nil).Times(4)
			},
		},
		{
			name:         "no active bundles",
			orgID:        "org2",
			wantBundles:  []string{},
			wantErr:      nil,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.Anything, "org2").
					Return(&FeatureActivation{Active: false}, nil)
			},
		},
		{
			name:         "repository error",
			orgID:        "org3",
			wantBundles:  nil,
			wantErr:      ErrRepositoryFailure,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.Anything, "org3").
					Return(nil, ErrRepositoryFailure)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			bundles, err := service.GetActiveBundles(context.Background(), tt.orgID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantBundles, bundles)
			}
		})
	}
}

func TestActivationService_IsBundleAvailable(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name        string
		orgID       string
		bundleID    string
		wantAvail   bool
		wantErr     error
		setup       func()
	}{
		{
			name:      "bundle available in tier",
			orgID:     "org1",
			bundleID:  "security",
			wantAvail: true,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
			},
		},
		{
			name:      "bundle not in tier",
			orgID:     "org2",
			bundleID:  "security",
			wantAvail: false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(&licensing.License{Tier: "free"}, nil)
			},
		},
		{
			name:      "invalid bundle",
			orgID:     "org1",
			bundleID:  "nonexistent",
			wantAvail: false,
			wantErr:   config.ErrInvalidBundle,
			setup:     func() {},
		},
		{
			name:      "no license",
			orgID:     "org3",
			bundleID:  "security",
			wantAvail: false,
			wantErr:   ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org3").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			available, err := service.IsBundleAvailable(context.Background(), tt.orgID, tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantAvail, available)
		})
	}
}

func TestActivationService_GetAvailableBundles(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name         string
		orgID        string
		wantBundles  []string
		wantErr      error
		setup        func()
	}{
		{
			name:         "enterprise tier bundles",
			orgID:        "org1",
			wantBundles:  []string{"security", "compliance", "advanced_auth"},
			wantErr:      nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
			},
		},
		{
			name:         "no license",
			orgID:        "org2",
			wantBundles:  nil,
			wantErr:      ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			bundles, err := service.GetAvailableBundles(context.Background(), tt.orgID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantBundles, bundles)
			}
		})
	}
}

func TestActivationService_GetAvailableFeatures(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name          string
		orgID         string
		wantFeatures  []string
		wantErr       error
		setup         func()
	}{
		{
			name:          "enterprise tier features",
			orgID:         "org1",
			wantFeatures:  []string{"advanced_sso", "audit_logs", "custom_roles"},
			wantErr:       nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
			},
		},
		{
			name:          "no license",
			orgID:         "org2",
			wantFeatures:  nil,
			wantErr:       ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			features, err := service.GetAvailableFeatures(context.Background(), tt.orgID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantFeatures, features)
			}
		})
	}
}

func TestActivationService_DeactivateBundle(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name     string
		orgID    string
		bundleID string
		wantErr  error
		setup    func()
	}{
		{
			name:     "successful bundle deactivation",
			orgID:    "org1",
			bundleID: "security",
			wantErr:  nil,
			setup: func() {
				mockRepo.On("Get", mock.Anything, mock.Anything, "org1").
					Return(&FeatureActivation{Active: true}, nil).Times(3)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(activation *FeatureActivation) bool {
					return !activation.Active
				})).Return(nil).Times(3)
			},
		},
		{
			name:     "invalid bundle",
			orgID:    "org1",
			bundleID: "nonexistent",
			wantErr:  config.ErrInvalidBundle,
			setup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.DeactivateBundle(context.Background(), tt.orgID, tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestActivationService_IsFeatureAvailable(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name        string
		orgID       string
		featureID   string
		wantAvail   bool
		wantErr     error
		setup       func()
	}{
		{
			name:      "feature available in tier",
			orgID:     "org1",
			featureID: "advanced_sso",
			wantAvail: true,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
			},
		},
		{
			name:      "feature not in tier",
			orgID:     "org2",
			featureID: "advanced_sso",
			wantAvail: false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(&licensing.License{Tier: "free"}, nil)
			},
		},
		{
			name:      "invalid feature",
			orgID:     "org1",
			featureID: "nonexistent",
			wantAvail: false,
			wantErr:   ErrInvalidFeature,
			setup:     func() {},
		},
		{
			name:      "no license",
			orgID:     "org3",
			featureID: "advanced_sso",
			wantAvail: false,
			wantErr:   ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org3").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			available, err := service.IsFeatureAvailable(context.Background(), tt.orgID, tt.featureID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantAvail, available)
		})
	}
}

func TestActivationService_GetFeatureActivationHistory(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	timeNow := time.Now()
	timePast := timeNow.Add(-24 * time.Hour)

	tests := []struct {
		name        string
		orgID       string
		featureID   string
		wantHistory []*FeatureActivation
		wantErr     error
		setup       func()
	}{
		{
			name:      "valid feature with history",
			orgID:     "org1",
			featureID: "advanced_sso",
			wantHistory: []*FeatureActivation{
				{
					FeatureID:   "advanced_sso",
					OrganizationID: "org1",
					ActivatedAt: timeNow,
					Active:      true,
				},
				{
					FeatureID:   "advanced_sso",
					OrganizationID: "org1",
					ActivatedAt: timePast,
					Active:      false,
				},
			},
			wantErr: nil,
			setup: func() {
				mockRepo.On("GetHistory", mock.Anything, "advanced_sso", "org1").
					Return([]*FeatureActivation{
						{
							FeatureID:   "advanced_sso",
							OrganizationID: "org1",
							ActivatedAt: timeNow,
							Active:      true,
						},
						{
							FeatureID:   "advanced_sso",
							OrganizationID: "org1",
							ActivatedAt: timePast,
							Active:      false,
						},
					}, nil)
			},
		},
		{
			name:        "invalid feature",
			orgID:       "org1",
			featureID:   "invalid_feature",
			wantHistory: nil,
			wantErr:     ErrInvalidFeature,
			setup:       func() {},
		},
		{
			name:        "repository error",
			orgID:       "org1",
			featureID:   "advanced_sso",
			wantHistory: nil,
			wantErr:     ErrRepositoryFailure,
			setup: func() {
				mockRepo.On("GetHistory", mock.Anything, "advanced_sso", "org1").
					Return(nil, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			history, err := service.GetFeatureActivationHistory(context.Background(), tt.orgID, tt.featureID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantHistory, history)
			}
		})
	}
}

func TestActivationService_CalculateFeaturePrice(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name          string
		orgID         string
		itemID        string
		isBundle      bool
		billingPeriod string
		currency      string
		wantPrice     *pricing.Price
		wantErr       error
		setup         func()
	}{
		{
			name:          "valid feature monthly price",
			orgID:         "org1",
			itemID:        "advanced_sso",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "USD",
			wantPrice:     &pricing.Price{Amount: "9.99", Currency: "USD"},
			wantErr:       nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
			},
		},
		{
			name:          "valid bundle yearly price",
			orgID:         "org2",
			itemID:        "security",
			isBundle:      true,
			billingPeriod: "yearly",
			currency:      "EUR",
			wantPrice:     &pricing.Price{Amount: "99.99", Currency: "EUR"},
			wantErr:       nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(&licensing.License{Tier: "business"}, nil)
			},
		},
		{
			name:          "invalid billing period",
			orgID:         "org1",
			itemID:        "advanced_sso",
			isBundle:      false,
			billingPeriod: "invalid",
			currency:      "USD",
			wantPrice:     nil,
			wantErr:       pricing.ErrInvalidBillingPeriod,
			setup:         func() {},
		},
		{
			name:          "invalid currency",
			orgID:         "org1",
			itemID:        "advanced_sso",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "INVALID",
			wantPrice:     nil,
			wantErr:       pricing.ErrInvalidCurrency,
			setup:         func() {},
		},
		{
			name:          "invalid feature",
			orgID:         "org1",
			itemID:        "invalid_feature",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "USD",
			wantPrice:     nil,
			wantErr:       ErrInvalidFeature,
			setup:         func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			price, err := service.CalculateFeaturePrice(context.Background(), tt.orgID, tt.itemID, tt.isBundle, tt.billingPeriod, tt.currency)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantPrice, price)
			}
		})
	}
}

func TestActivationService_InitiateFeaturePayment(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name          string
		orgID         string
		itemID        string
		isBundle      bool
		billingPeriod string
		currency      string
		wantDetails   *payment.PaymentResponse
		wantErr       error
		setup         func()
	}{
		{
			name:          "successful feature payment",
			orgID:         "org1",
			itemID:        "advanced_sso",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "USD",
			wantDetails: &payment.PaymentResponse{
				ID:     "pay_123",
				Status: "pending",
				Amount: payment.Amount{
					Value:    "9.99",
					Currency: "USD",
				},
				RedirectURL: "https://passwordimmunity.com/payment/success",
				WebhookURL:  "https://api.passwordimmunity.com/webhooks/mollie",
				Metadata: payment.PaymentMetadata{
					OrganizationID: "org1",
					Features:       []string{"advanced_sso"},
					Duration:      "monthly",
				},
			},
			wantErr: nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockPaymentService.On("ProcessPayment", mock.Anything, mock.MatchedBy(func(req payment.PaymentRequest) bool {
					return req.Amount.Value == "9.99" && req.Amount.Currency == "USD" && req.Metadata.OrganizationID == "org1"
				})).Return(&payment.PaymentResponse{
					ID:     "pay_123",
					Status: "pending",
					Amount: payment.Amount{
						Value:    "9.99",
						Currency: "USD",
					},
					RedirectURL: "https://passwordimmunity.com/payment/success",
					WebhookURL:  "https://api.passwordimmunity.com/webhooks/mollie",
					Metadata: payment.PaymentMetadata{
						OrganizationID: "org1",
						Features:       []string{"advanced_sso"},
						Duration:      "monthly",
					},
				}, nil)
			},
		},
		{
			name:          "invalid feature",
			orgID:         "org1",
			itemID:        "invalid_feature",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "USD",
			wantDetails:   nil,
			wantErr:       ErrInvalidFeature,
			setup:         func() {},
		},
		{
			name:          "payment creation failed",
			orgID:         "org1",
			itemID:        "advanced_sso",
			isBundle:      false,
			billingPeriod: "monthly",
			currency:      "USD",
			wantDetails:   nil,
			wantErr:       fmt.Errorf("failed to create payment: payment service error"),
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockPaymentService.On("CreatePayment", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("payment service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			details, err := service.InitiateFeaturePayment(context.Background(), tt.orgID, tt.itemID, tt.isBundle, tt.billingPeriod, tt.currency)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDetails, details)
			}
		})
	}
}

func TestActivationService_IsFeatureAvailableForPurchase(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	service := NewActivationService(mockLicenseService, mockRepo)

	tests := []struct {
		name      string
		orgID     string
		featureID string
		want      bool
		wantErr   error
		setup     func()
	}{
		{
			name:      "available for purchase",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      true,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockRepo.On("Get", mock.Anything, "basic_sso", "org1").
					Return(&FeatureActivation{Active: true}, nil)
			},
		},
		{
			name:      "insufficient tier",
			orgID:     "org2",
			featureID: "advanced_sso",
			want:      false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org2").
					Return(&licensing.License{Tier: "free"}, nil)
			},
		},
		{
			name:      "missing dependency",
			orgID:     "org3",
			featureID: "advanced_sso",
			want:      false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org3").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockRepo.On("Get", mock.Anything, "basic_sso", "org3").
					Return(&FeatureActivation{Active: false}, nil)
			},
		},
		{
			name:      "invalid feature",
			orgID:     "org1",
			featureID: "invalid_feature",
			want:      false,
			wantErr:   ErrInvalidFeature,
			setup:     func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			available, err := service.IsFeatureAvailableForPurchase(context.Background(), tt.orgID, tt.featureID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, available)
		})
	}
}

func TestActivationService_IsFeatureActive(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name      string
		orgID     string
		featureID string
		want      bool
		wantErr   error
		setup     func()
	}{
		{
			name:      "feature active with valid license",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      true,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockRepo.On("Get", mock.Anything, "advanced_sso", "org1").
					Return(&FeatureActivation{Active: true}, nil)
			},
		},
		{
			name:      "feature not found",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "enterprise"}, nil)
				mockRepo.On("Get", mock.Anything, "advanced_sso", "org1").
					Return(nil, ErrFeatureNotFound)
			},
		},
		{
			name:      "invalid feature",
			orgID:     "org1",
			featureID: "invalid_feature",
			want:      false,
			wantErr:   ErrInvalidFeature,
			setup:     func() {},
		},
		{
			name:      "no license",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      false,
			wantErr:   ErrLicenseRequired,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(nil, licensing.ErrNoActiveLicense)
			},
		},
		{
			name:      "feature not in tier",
			orgID:     "org1",
			featureID: "advanced_sso",
			want:      false,
			wantErr:   nil,
			setup: func() {
				mockLicenseService.On("GetActiveLicense", mock.Anything, "org1").
					Return(&licensing.License{Tier: "free"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			active, err := service.IsFeatureActive(context.Background(), tt.orgID, tt.featureID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, active)
		})
	}
}

func TestActivationService_GetBundleFeatures(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	mockPaymentService := new(MockPaymentService)
	service := NewActivationService(mockLicenseService, mockRepo, mockPaymentService)

	tests := []struct {
		name         string
		bundleID     string
		wantFeatures []string
		wantErr      error
	}{
		{
			name:         "valid bundle",
			bundleID:     "security",
			wantFeatures: []string{"advanced_sso", "mfa", "audit_logs"},
			wantErr:      nil,
		},
		{
			name:         "invalid bundle",
			bundleID:     "nonexistent",
			wantFeatures: nil,
			wantErr:      config.ErrInvalidBundle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features, err := service.GetBundleFeatures(context.Background(), tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.wantFeatures, features)
			}
		})
	}
}

func TestActivationService_IsFeatureInBundle(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	service := NewActivationService(mockLicenseService, mockRepo)

	tests := []struct {
		name      string
		featureID string
		bundleID  string
		want      bool
		wantErr   error
	}{
		{
			name:      "feature in bundle",
			featureID: "advanced_sso",
			bundleID:  "security",
			want:      true,
			wantErr:   nil,
		},
		{
			name:      "feature not in bundle",
			featureID: "advanced_sso",
			bundleID:  "compliance",
			want:      false,
			wantErr:   nil,
		},
		{
			name:      "invalid feature",
			featureID: "invalid_feature",
			bundleID:  "security",
			want:      false,
			wantErr:   ErrInvalidFeature,
		},
		{
			name:      "invalid bundle",
			featureID: "advanced_sso",
			bundleID:  "invalid_bundle",
			want:      false,
			wantErr:   config.ErrInvalidBundle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBundle, err := service.IsFeatureInBundle(context.Background(), tt.featureID, tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.want, inBundle)
			}
		})
	}
}

func TestActivationService_IsFeatureInBundle(t *testing.T) {
	mockLicenseService := new(MockLicenseService)
	mockRepo := new(MockRepository)
	service := NewActivationService(mockLicenseService, mockRepo)

	tests := []struct {
		name      string
		featureID string
		bundleID  string
		want      bool
		wantErr   error
	}{
		{
			name:      "feature in bundle",
			featureID: "advanced_sso",
			bundleID:  "security",
			want:      true,
			wantErr:   nil,
		},
		{
			name:      "feature not in bundle",
			featureID: "advanced_sso",
			bundleID:  "compliance",
			want:      false,
			wantErr:   nil,
		},
		{
			name:      "invalid feature",
			featureID: "invalid_feature",
			bundleID:  "security",
			want:      false,
			wantErr:   ErrInvalidFeature,
		},
		{
			name:      "invalid bundle",
			featureID: "advanced_sso",
			bundleID:  "invalid_bundle",
			want:      false,
			wantErr:   config.ErrInvalidBundle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBundle, err := service.IsFeatureInBundle(context.Background(), tt.featureID, tt.bundleID)
			assert.Equal(t, tt.wantErr, err)
			if err == nil {
				assert.Equal(t, tt.want, inBundle)
			}
		})
	}
}
