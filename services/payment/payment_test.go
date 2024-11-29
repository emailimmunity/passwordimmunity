package payment

import (
	"context"
	"testing"
)

// MockPaymentProvider implements PaymentProvider for testing
type MockPaymentProvider struct {
	createPaymentFunc  func(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	verifyPaymentFunc  func(ctx context.Context, paymentID string) (*PaymentResponse, error)
	handleWebhookFunc  func(payment *PaymentResponse) error
}

func (m *MockPaymentProvider) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	return m.createPaymentFunc(ctx, req)
}

func (m *MockPaymentProvider) VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return m.verifyPaymentFunc(ctx, paymentID)
}

func (m *MockPaymentProvider) HandleWebhook(payment *PaymentResponse) error {
	return m.handleWebhookFunc(payment)
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	mockProvider := &MockPaymentProvider{
		createPaymentFunc: func(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
			return &PaymentResponse{
				ID:     "test_payment",
				Status: "pending",
			}, nil
		},
	}

	service := NewPaymentService(mockProvider)

	tests := []struct {
		name    string
		req     PaymentRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: PaymentRequest{
				Amount: Amount{
					Currency: "EUR",
					Value:    "10.00",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid amount",
			req: PaymentRequest{
				Amount: Amount{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ProcessPayment(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentService_VerifyAndActivate(t *testing.T) {
	mockProvider := &MockPaymentProvider{
		verifyPaymentFunc: func(ctx context.Context, paymentID string) (*PaymentResponse, error) {
			return &PaymentResponse{
				ID:     paymentID,
				Status: "paid",
			}, nil
		},
		handleWebhookFunc: func(payment *PaymentResponse) error {
			return nil
		},
	}

	service := NewPaymentService(mockProvider)

	err := service.VerifyAndActivate(context.Background(), "test_payment")
	if err != nil {
		t.Errorf("VerifyAndActivate() error = %v", err)
	}
}
