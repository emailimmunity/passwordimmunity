package payment

import (
    "context"
    "testing"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockEmailService struct {
    mock.Mock
}

func (m *mockEmailService) SendEmail(ctx context.Context, to string, subject string, body string) error {
    args := m.Called(ctx, to, subject, body)
    return args.Error(0)
}

func TestNotificationService(t *testing.T) {
    mockEmail := new(mockEmailService)
    service := NewNotificationService(mockEmail)
    ctx := context.Background()

    t.Run("notify payment failed", func(t *testing.T) {
        payment := &models.Payment{
            ID:             uuid.New(),
            OrganizationID: uuid.New(),
            ProviderID:     "tr_test123",
            Amount:         "999.00",
            Currency:       "EUR",
            LicenseType:    "enterprise",
            Period:         "yearly",
        }

        mockEmail.On("SendEmail",
            ctx,
            "admin@passwordimmunity.com",
            mock.MatchedBy(func(s string) bool { return true }), // Subject contains payment ID
            mock.MatchedBy(func(s string) bool { return true }), // Body contains payment details
        ).Return(nil)

        err := service.NotifyPaymentFailed(ctx, payment, "Payment declined")
        assert.NoError(t, err)
        mockEmail.AssertExpectations(t)
    })

    t.Run("notify license canceled", func(t *testing.T) {
        license := &models.License{
            ID:             uuid.New(),
            OrganizationID: uuid.New(),
            Type:           "enterprise",
            Status:         "canceled",
        }

        mockEmail.On("SendEmail",
            ctx,
            "admin@passwordimmunity.com",
            mock.MatchedBy(func(s string) bool { return true }), // Subject contains license ID
            mock.MatchedBy(func(s string) bool { return true }), // Body contains license details
        ).Return(nil)

        err := service.NotifyLicenseCanceled(ctx, license, "Payment failed")
        assert.NoError(t, err)
        mockEmail.AssertExpectations(t)
    })
}
