package testutil

import (
    "context"
    "time"

    "github.com/emailimmunity/passwordimmunity/services/payment"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
)

// MockPaymentService is a mock implementation of the payment.Service interface
type MockPaymentService struct {
    mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, orgID uuid.UUID, amount float64, currency string) (*payment.Payment, error) {
    args := m.Called(ctx, orgID, amount, currency)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*payment.Payment), args.Error(1)
}

func (m *MockPaymentService) HandleWebhook(ctx context.Context, payload []byte) error {
    args := m.Called(ctx, payload)
    return args.Error(0)
}

// CreateTestPayment creates a test payment with the specified status
func CreateTestPayment(status string) *payment.Payment {
    return &payment.Payment{
        ID:            uuid.New(),
        OrganizationID: uuid.New(),
        Amount:        999.00,
        Currency:      "EUR",
        Status:        status,
        Provider:      "mollie",
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
}

// SetupMockPaymentService configures a mock payment service with common expectations
func SetupMockPaymentService() *MockPaymentService {
    mockPS := new(MockPaymentService)
    payment := CreateTestPayment("paid")

    mockPS.On("CreatePayment", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(payment, nil)
    mockPS.On("HandleWebhook", mock.Anything, mock.Anything).Return(nil)

    return mockPS
}

// GetTestPricing returns test pricing data for different license tiers
func GetTestPricing() map[string]map[string]float64 {
    return map[string]map[string]float64{
        "premium": {
            "monthly": 49.00,
            "yearly":  499.00,
        },
        "enterprise": {
            "monthly": 99.00,
            "yearly":  999.00,
        },
    }
}
