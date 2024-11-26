package payment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMollieClient struct {
	mock.Mock
}

func (m *mockMollieClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Payment), args.Error(1)
}

func (m *mockMollieClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Payment), args.Error(1)
}

func (m *mockMollieClient) CancelPayment(ctx context.Context, paymentID string) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

func TestCreatePayment(t *testing.T) {
	mockClient := new(mockMollieClient)
	config := &Config{
		Currency:      "EUR",
		RetryAttempts: 1,
		RetryDelay:    time.Millisecond,
	}

	service := &service{
		client: mockClient,
		config: config,
	}

	ctx := context.Background()
	req := &CreatePaymentRequest{
		Amount:      "10.00",
		Description: "Test payment",
		OrderID:     "order_123",
		CustomerID:  "customer_123",
	}

	expectedPayment := &Payment{
		ID:          "payment_123",
		Status:      PaymentStatusPending,
		Amount:      "10.00",
		Currency:    "EUR",
		Description: "Test payment",
		OrderID:     "order_123",
		CustomerID:  "customer_123",
		CreatedAt:   time.Now(),
	}

	mockClient.On("CreatePayment", ctx, req).Return(expectedPayment, nil)

	payment, err := service.CreatePayment(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedPayment, payment)
	mockClient.AssertExpectations(t)
}

func TestGetPayment(t *testing.T) {
	mockClient := new(mockMollieClient)
	config := &Config{
		RetryAttempts: 1,
		RetryDelay:    time.Millisecond,
	}

	service := &service{
		client: mockClient,
		config: config,
	}

	ctx := context.Background()
	paymentID := "payment_123"

	expectedPayment := &Payment{
		ID:          paymentID,
		Status:      PaymentStatusPaid,
		Amount:      "10.00",
		Currency:    "EUR",
		Description: "Test payment",
		CreatedAt:   time.Now(),
	}

	mockClient.On("GetPayment", ctx, paymentID).Return(expectedPayment, nil)

	payment, err := service.GetPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPayment, payment)
	mockClient.AssertExpectations(t)
}

func TestHandleWebhook(t *testing.T) {
	mockClient := new(mockMollieClient)
	config := &Config{
		RetryAttempts: 1,
		RetryDelay:    time.Millisecond,
	}

	service := &service{
		client: mockClient,
		config: config,
	}

	ctx := context.Background()
	webhookReq := &WebhookRequest{
		PaymentID: "payment_123",
		Status:    PaymentStatusPaid,
	}

	payment := &Payment{
		ID:     webhookReq.PaymentID,
		Status: webhookReq.Status,
	}

	mockClient.On("GetPayment", ctx, webhookReq.PaymentID).Return(payment, nil)

	err := service.HandleWebhook(ctx, webhookReq)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
