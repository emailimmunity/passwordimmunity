package payment

import (
    "context"
    "testing"

    "github.com/mollie/mollie-api-go/v2/mollie"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockMollieClient struct {
    mock.Mock
}

func (m *mockMollieClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*mollie.Payment, error) {
    args := m.Called(ctx, req)
    if p, ok := args.Get(0).(*mollie.Payment); ok {
        return p, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *mockMollieClient) GetPayment(ctx context.Context, paymentID string) (*mollie.Payment, error) {
    args := m.Called(ctx, paymentID)
    if p, ok := args.Get(0).(*mollie.Payment); ok {
        return p, args.Error(1)
    }
    return nil, args.Error(1)
}

func TestMollieProvider_CreatePayment(t *testing.T) {
    mockClient := &mockMollieClient{}
    provider := NewMollieProvider(mockClient)

    t.Run("successful payment creation", func(t *testing.T) {
        req := PaymentRequest{
            Amount: Amount{
                Currency: "EUR",
                Value:    "10.00",
            },
            Description: "Test payment",
            RedirectURL: "https://example.com/return",
            WebhookURL:  "https://example.com/webhook",
            Metadata: PaymentMetadata{
                OrganizationID: "test-org",
                Features:       []string{"advanced_sso"},
                Duration:      "monthly",
            },
        }

        expectedMolliePayment := &mollie.Payment{
            ID: "tr_test123",
            Amount: &mollie.Amount{
                Currency: "EUR",
                Value:    "10.00",
            },
            Status:      "open",
            RedirectURL: "https://example.com/return",
            WebhookURL:  "https://example.com/webhook",
        }

        mockClient.On("CreatePayment", mock.Anything, mock.AnythingOfType("*payment.CreatePaymentRequest")).
            Return(expectedMolliePayment, nil)

        response, err := provider.CreatePayment(context.Background(), req)

        assert.NoError(t, err)
        assert.Equal(t, expectedMolliePayment.ID, response.ID)
        assert.Equal(t, expectedMolliePayment.Status, response.Status)
        assert.Equal(t, req.Amount, response.Amount)
        mockClient.AssertExpectations(t)
    })
}

func TestMollieProvider_VerifyPayment(t *testing.T) {
    mockClient := &mockMollieClient{}
    provider := NewMollieProvider(mockClient)

    t.Run("successful payment verification", func(t *testing.T) {
        paymentID := "tr_test123"
        molliePayment := &mollie.Payment{
            ID: paymentID,
            Amount: &mollie.Amount{
                Currency: "EUR",
                Value:    "10.00",
            },
            Status: "paid",
            Metadata: map[string]interface{}{
                "org_id":     "test-org",
                "feature_id": "advanced_sso",
                "duration":   "monthly",
            },
        }

        mockClient.On("GetPayment", mock.Anything, paymentID).Return(molliePayment, nil)

        response, err := provider.VerifyPayment(context.Background(), paymentID)

        assert.NoError(t, err)
        assert.Equal(t, molliePayment.ID, response.ID)
        assert.Equal(t, molliePayment.Status, response.Status)
        assert.Equal(t, molliePayment.Amount.Currency, response.Amount.Currency)
        assert.Equal(t, molliePayment.Amount.Value, response.Amount.Value)
        assert.Equal(t, "test-org", response.Metadata.OrganizationID)
        assert.Equal(t, []string{"advanced_sso"}, response.Metadata.Features)
        mockClient.AssertExpectations(t)
    })

    t.Run("bundle payment verification", func(t *testing.T) {
        paymentID := "tr_test456"
        molliePayment := &mollie.Payment{
            ID: paymentID,
            Amount: &mollie.Amount{
                Currency: "EUR",
                Value:    "50.00",
            },
            Status: "paid",
            Metadata: map[string]interface{}{
                "org_id":    "test-org",
                "bundle_id": "enterprise",
                "duration":  "yearly",
            },
        }

        mockClient.On("GetPayment", mock.Anything, paymentID).Return(molliePayment, nil)

        response, err := provider.VerifyPayment(context.Background(), paymentID)

        assert.NoError(t, err)
        assert.Equal(t, molliePayment.ID, response.ID)
        assert.Equal(t, molliePayment.Status, response.Status)
        assert.Equal(t, []string{"enterprise"}, response.Metadata.Bundles)
        mockClient.AssertExpectations(t)
    })
}
