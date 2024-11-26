package payment

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/mollie/mollie-api-go/v2/mollie"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

func (m *mockRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *mockRepository) GetPaymentByProviderID(ctx context.Context, providerID string) (*models.Payment, error) {
    args := m.Called(ctx, providerID)
    return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *mockRepository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

func (m *mockRepository) CreateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func (m *mockRepository) GetLicenseByPaymentID(ctx context.Context, paymentID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, paymentID)
    return args.Get(0).(*models.License), args.Error(1)
}

func (m *mockRepository) UpdateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func (m *mockRepository) GetLicenseByPaymentID(ctx context.Context, paymentID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, paymentID)
    return args.Get(0).(*models.License), args.Error(1)
}

type mockMollieClient struct {
    mock.Mock
}

func (m *mockMollieClient) CreatePayment(amount, currency, description string) (*mollie.Payment, error) {
    args := m.Called(amount, currency, description)
    return args.Get(0).(*mollie.Payment), args.Error(1)
}

func TestCreatePayment(t *testing.T) {
    // Set up test environment variables
    os.Setenv("MOLLIE_API_KEY", "test_key")
    os.Setenv("MOLLIE_TEST_MODE", "true")
    defer os.Clearenv()

    mockRepo := new(mockRepository)
    service, err := NewService(mockRepo)
    assert.NoError(t, err)

    ctx := context.Background()
    orgID := uuid.New()

    molliePayment := &mollie.Payment{
        ID: "tr_test123",
    }

    service.mollieClient.(*mockMollieClient).On("CreatePayment", "999.00", "EUR", "enterprise").Return(molliePayment, nil)
    mockRepo.On("CreatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)

    payment, err := service.CreatePayment(ctx, orgID, "enterprise", "yearly")
    assert.NoError(t, err)
    assert.Equal(t, "999.00", payment.Amount)
    assert.Equal(t, "tr_test123", payment.ProviderID)

    // Clean up test environment
    mockRepo.AssertExpectations(t)
}

func TestHandlePaymentWebhook(t *testing.T) {
    // Set up test environment variables
    os.Setenv("MOLLIE_API_KEY", "test_key")
    os.Setenv("MOLLIE_TEST_MODE", "true")

    mockRepo := new(mockRepository)
    service, err := NewService(mockRepo)
    assert.NoError(t, err)

    ctx := context.Background()

    t.Run("successful payment", func(t *testing.T) {
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: uuid.New(),
            ProviderID:     "tr_test123",
            Amount:         "999.00",
            LicenseType:    "enterprise",
            Period:         "yearly",
        }

        mockRepo.On("GetPaymentByProviderID", ctx, "tr_test123").Return(payment, nil)
        mockRepo.On("UpdatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)
        mockRepo.On("CreateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

        err = service.HandlePaymentWebhook(ctx, "tr_test123", "paid")
        assert.NoError(t, err)

        mockRepo.AssertCalled(t, "CreateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
            return license.Type == "enterprise" && license.Status == "active"
        }))
    })

    t.Run("failed payment", func(t *testing.T) {
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: uuid.New(),
            ProviderID:     "tr_test456",
            Amount:         "999.00",
            LicenseType:    "enterprise",
            Period:         "yearly",
        }

        license := &models.License{
            ID:             uuid.New(),
            OrganizationID: payment.OrganizationID,
            PaymentID:      payment.ID,
            Status:         "active",
        }

        mockRepo.On("GetPaymentByProviderID", ctx, "tr_test456").Return(payment, nil)
        mockRepo.On("UpdatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)
        mockRepo.On("GetLicenseByPaymentID", ctx, payment.ID).Return(license, nil)
        mockRepo.On("UpdateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

        err = service.HandlePaymentWebhook(ctx, "tr_test456", "failed")
        assert.NoError(t, err)

        mockRepo.AssertCalled(t, "UpdateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
            return license.Status == "canceled"
        }))
    })
}

func TestCalculateAmount(t *testing.T) {
    tests := []struct {
        name        string
        licenseType string
        period      string
        expected    string
    }{
        {"Enterprise Yearly", "enterprise", "yearly", "999.00"},
        {"Enterprise Monthly", "enterprise", "monthly", "99.00"},
        {"Premium Yearly", "premium", "yearly", "499.00"},
        {"Premium Monthly", "premium", "monthly", "49.00"},
        {"Free", "free", "monthly", "0.00"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            amount := calculateAmount(tt.licenseType, tt.period)
            assert.Equal(t, tt.expected, amount)
        })
    }
}

func TestNewServiceWithConfig(t *testing.T) {
    tests := []struct {
        name          string
        envVars       map[string]string
        expectedError string
    }{
        {
            name: "successful initialization in live mode",
            envVars: map[string]string{
                "MOLLIE_API_KEY":     "live_test_key",
                "MOLLIE_TEST_API_KEY": "test_key",
                "MOLLIE_TEST_MODE":    "false",
            },
        },
        {
            name: "missing required API key",
            envVars: map[string]string{
                "MOLLIE_TEST_MODE": "false",
            },
            expectedError: "failed to load Mollie configuration: live mode enabled but MOLLIE_API_KEY not set",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Clear environment
            os.Clearenv()

            // Set test environment variables
            for k, v := range tt.envVars {
                os.Setenv(k, v)
            }

            mockRepo := new(mockRepository)
            service, err := NewService(mockRepo)

            if tt.expectedError != "" {
                assert.EqualError(t, err, tt.expectedError)
                assert.Nil(t, service)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, service)
                assert.NotNil(t, service.config)
                assert.NotNil(t, service.client)
                assert.Equal(t, mockRepo, service.repository)
            }
        })
    }
}

func TestCancelLicense(t *testing.T) {
    // Set up test environment
    os.Setenv("MOLLIE_API_KEY", "test_key")
    os.Setenv("MOLLIE_TEST_MODE", "true")
    defer os.Clearenv()

    mockRepo := new(mockRepository)
    service, err := NewService(mockRepo)
    assert.NoError(t, err)

    t.Run("successful cancellation", func(t *testing.T) {
        ctx := context.Background()
        paymentID := "tr_test789"
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: uuid.New(),
            ProviderID:     paymentID,
            Status:         "failed",
        }
        license := &models.License{
            ID:             uuid.New(),
            OrganizationID: payment.OrganizationID,
            PaymentID:      payment.ID,
            Status:         "active",
        }

        mockRepo.On("GetPaymentByProviderID", ctx, paymentID).Return(payment, nil)
        mockRepo.On("GetLicenseByPaymentID", ctx, payment.ID).Return(license, nil)
        mockRepo.On("UpdateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

        err = service.CancelLicense(ctx, paymentID)
        assert.NoError(t, err)

        mockRepo.AssertCalled(t, "UpdateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
            return license.Status == "canceled"
        }))
    })

    t.Run("payment not found", func(t *testing.T) {
        ctx := context.Background()
        mockRepo.On("GetPaymentByProviderID", ctx, "nonexistent").Return((*models.Payment)(nil), assert.AnError)

        err = service.CancelLicense(ctx, "nonexistent")
        assert.Error(t, err)
    })

    t.Run("license not found", func(t *testing.T) {
        ctx := context.Background()
        payment := &models.Payment{
            ID:         uuid.New(),
            ProviderID: "tr_test999",
        }

        mockRepo.On("GetPaymentByProviderID", ctx, payment.ProviderID).Return(payment, nil)
        mockRepo.On("GetLicenseByPaymentID", ctx, payment.ID).Return((*models.License)(nil), assert.AnError)

        err = service.CancelLicense(ctx, payment.ProviderID)
        assert.Error(t, err)
    })
}

func TestCancelLicense(t *testing.T) {
    // Set up test environment
    os.Setenv("MOLLIE_API_KEY", "test_key")
    os.Setenv("MOLLIE_TEST_MODE", "true")
    defer os.Clearenv()

    mockRepo := new(mockRepository)
    service, err := NewService(mockRepo)
    assert.NoError(t, err)

    ctx := context.Background()
    paymentID := "tr_test789"
    payment := &models.Payment{
        ID:             uuid.New(),
        OrganizationID: uuid.New(),
        ProviderID:     paymentID,
        Status:         "failed",
    }
    license := &models.License{
        ID:             uuid.New(),
        OrganizationID: payment.OrganizationID,
        PaymentID:      payment.ID,
        Status:         "active",
    }

    mockRepo.On("GetPaymentByProviderID", ctx, paymentID).Return(payment, nil)
    mockRepo.On("GetLicenseByPaymentID", ctx, payment.ID).Return(license, nil)
    mockRepo.On("UpdateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

    err = service.CancelLicense(ctx, paymentID)
    assert.NoError(t, err)

    mockRepo.AssertCalled(t, "UpdateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
        return license.Status == "canceled"
    }))
}
