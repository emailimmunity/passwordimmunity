package payment

import (
    "context"
    "net/http"
    "net/url"
    "testing"
    "strings"

    "github.com/mollie/mollie-api-go/v2/mollie"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockMollieAPI struct {
    mock.Mock
}

func (m *mockMollieAPI) Create(payment *mollie.Payment) (*mollie.Payment, error) {
    args := m.Called(payment)
    if p, ok := args.Get(0).(*mollie.Payment); ok {
        return p, args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *mockMollieAPI) Get(id string) (*mollie.Payment, error) {
    args := m.Called(id)
    if p, ok := args.Get(0).(*mollie.Payment); ok {
        return p, args.Error(1)
    }
    return nil, args.Error(1)
}

func TestMollieClient_CreatePayment(t *testing.T) {
    mockAPI := &mockMollieAPI{}
    client := &MollieClient{
        client: &mollie.Client{},
        config: &Config{
            WebhookURL:  "https://example.com/webhook",
            RedirectURL: "https://example.com/return",
        },
    }

    t.Run("successful payment creation", func(t *testing.T) {
        req := &CreatePaymentRequest{
            Amount:      10.00,
            Currency:    "EUR",
            Description: "Test payment",
            FeatureID:   "advanced_sso",
            OrgID:       "test-org",
        }

        expectedPayment := &mollie.Payment{
            ID: "tr_test123",
            Amount: &mollie.Amount{
                Currency: "EUR",
                Value:    "10.00",
            },
            Status: "open",
        }

        mockAPI.On("Create", mock.AnythingOfType("*mollie.Payment")).Return(expectedPayment, nil)

        payment, err := client.CreatePayment(context.Background(), req)

        assert.NoError(t, err)
        assert.Equal(t, expectedPayment.ID, payment.ID)
        mockAPI.AssertExpectations(t)
    })
}

func TestMollieClient_HandleWebhook(t *testing.T) {
    mockAPI := &mockMollieAPI{}
    client := &MollieClient{
        client: &mollie.Client{},
        config: &Config{},
    }

    t.Run("successful webhook handling", func(t *testing.T) {
        paymentID := "tr_test123"
        payment := &mollie.Payment{
            ID: paymentID,
            Status: "paid",
            Metadata: map[string]interface{}{
                "feature_id": "advanced_sso",
                "org_id":    "test-org",
            },
        }

        mockAPI.On("Get", paymentID).Return(payment, nil)

        err := client.HandleWebhook(context.Background(), paymentID)
        assert.NoError(t, err)
        mockAPI.AssertExpectations(t)
    })

    t.Run("unpaid payment status", func(t *testing.T) {
        paymentID := "tr_test123"
        payment := &mollie.Payment{
            ID: paymentID,
            Status: "pending",
        }

        mockAPI.On("Get", paymentID).Return(payment, nil)

        err := client.HandleWebhook(context.Background(), paymentID)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "payment status is pending")
        mockAPI.AssertExpectations(t)
    })
}

func TestMollieClient_ValidateWebhook(t *testing.T) {
    client := &MollieClient{}

    t.Run("valid webhook request", func(t *testing.T) {
        form := url.Values{}
        form.Add("id", "tr_test123")

        req, _ := http.NewRequest("POST", "/webhook", strings.NewReader(form.Encode()))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        paymentID, err := client.ValidateWebhook(req)

        assert.NoError(t, err)
        assert.Equal(t, "tr_test123", paymentID)
    })

    t.Run("missing payment ID", func(t *testing.T) {
        req, _ := http.NewRequest("POST", "/webhook", strings.NewReader(""))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        paymentID, err := client.ValidateWebhook(req)

        assert.Error(t, err)
        assert.Equal(t, "", paymentID)
        assert.Contains(t, err.Error(), "missing payment ID")
    })
}
