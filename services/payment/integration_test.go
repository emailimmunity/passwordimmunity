package payment

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

func (m *mockRepository) GetPaymentByProviderID(ctx context.Context, providerID string) (*models.Payment, error) {
    args := m.Called(ctx, providerID)
    return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *mockRepository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
    args := m.Called(ctx, payment)
    return args.Error(0)
}

func (m *mockRepository) GetLicenseByPaymentID(ctx context.Context, paymentID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, paymentID)
    return args.Get(0).(*models.License), args.Error(1)
}

func (m *mockRepository) CreateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func (m *mockRepository) UpdateLicense(ctx context.Context, license *models.License) error {
    args := m.Called(ctx, license)
    return args.Error(0)
}

func TestPaymentWorkflow(t *testing.T) {
    // Set up test environment
    os.Setenv("MOLLIE_API_KEY", "test_key")
    os.Setenv("MOLLIE_TEST_MODE", "true")
    defer os.Clearenv()

    mockRepo := new(mockRepository)
    mockEmail := new(mockEmailService)

    // Override the NewNotificationService creation in NewService
    originalNewNotificationService := NewNotificationService
    defer func() { NewNotificationService = originalNewNotificationService }()

    NewNotificationService = func(emailService EmailService) NotificationService {
        return NewNotificationServiceWithEmail(mockEmail)
    }

    service, err := NewService(mockRepo)
    require.NoError(t, err)

    ctx := context.Background()
    orgID := uuid.New()

    // Test full payment workflow
    t.Run("successful payment workflow", func(t *testing.T) {
        // Step 1: Create payment
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: orgID,
            ProviderID:     "tr_test123",
            Amount:         "999.00",
            Currency:       "EUR",
            Status:         "pending",
            LicenseType:    "enterprise",
            Period:         "yearly",
        }

        mockRepo.On("CreatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)
        mockRepo.On("GetPaymentByProviderID", ctx, payment.ProviderID).Return(payment, nil)
        mockRepo.On("UpdatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)
        mockRepo.On("CreateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

        // Create initial payment
        createdPayment, err := service.CreatePayment(ctx, orgID, "enterprise", "yearly")
        require.NoError(t, err)
        assert.Equal(t, "pending", createdPayment.Status)

        // Step 2: Handle successful payment webhook
        err = service.HandlePaymentWebhook(ctx, payment.ProviderID, "paid")
        require.NoError(t, err)

        // Verify license creation
        mockRepo.AssertCalled(t, "CreateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
            return license.Type == "enterprise" &&
                   license.Status == "active" &&
                   license.OrganizationID == orgID
        }))
    })

    t.Run("failed payment workflow", func(t *testing.T) {
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: orgID,
            ProviderID:     "tr_test456",
            Amount:         "999.00",
            Currency:       "EUR",
            Status:         "pending",
            LicenseType:    "enterprise",
            Period:         "yearly",
        }

        license := &models.License{
            ID:             uuid.New(),
            OrganizationID: orgID,
            PaymentID:      payment.ID,
            Status:         "active",
        }

        mockRepo.On("GetPaymentByProviderID", ctx, payment.ProviderID).Return(payment, nil)
        mockRepo.On("UpdatePayment", ctx, mock.AnythingOfType("*models.Payment")).Return(nil)
        mockRepo.On("GetLicenseByPaymentID", ctx, payment.ID).Return(license, nil)
        mockRepo.On("UpdateLicense", ctx, mock.AnythingOfType("*models.License")).Return(nil)

        // Set up notification expectations
        mockEmail.On("SendEmail",
            ctx,
            "admin@passwordimmunity.com",
            mock.MatchedBy(func(s string) bool { return true }),
            mock.MatchedBy(func(s string) bool { return true }),
        ).Return(nil)

        // Handle failed payment webhook
        err = service.HandlePaymentWebhook(ctx, payment.ProviderID, "failed")
        require.NoError(t, err)

        // Verify license cancellation and notification
        mockRepo.AssertCalled(t, "UpdateLicense", ctx, mock.MatchedBy(func(license *models.License) bool {
            return license.Status == "canceled"
        }))
        mockEmail.AssertExpectations(t)
    })
}
