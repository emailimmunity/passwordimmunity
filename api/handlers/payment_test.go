package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPaymentService struct {
	mock.Mock
}

func (m *mockPaymentService) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.Payment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.Payment), args.Error(1)
}

func (m *mockPaymentService) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.Payment), args.Error(1)
}

func (m *mockPaymentService) CancelPayment(ctx context.Context, paymentID string) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

func (m *mockPaymentService) HandleWebhook(ctx context.Context, req *payment.WebhookRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestCreatePaymentHandler(t *testing.T) {
	mockService := new(mockPaymentService)
	handler := NewPaymentHandler(mockService)

	req := &payment.CreatePaymentRequest{
		Amount:      "10.00",
		Currency:    "EUR",
		Description: "Test payment",
	}

	expectedPayment := &payment.Payment{
		ID:          "payment_123",
		Status:      payment.PaymentStatusPending,
		Amount:      "10.00",
		Currency:    "EUR",
		Description: "Test payment",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/payments", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("CreatePayment", httpReq.Context(), req).Return(expectedPayment, nil)

	handler.CreatePayment(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response payment.Payment
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedPayment.ID, response.ID)
	mockService.AssertExpectations(t)
}

func TestGetPaymentHandler(t *testing.T) {
	mockService := new(mockPaymentService)
	handler := NewPaymentHandler(mockService)

	expectedPayment := &payment.Payment{
		ID:          "payment_123",
		Status:      payment.PaymentStatusPaid,
		Amount:      "10.00",
		Currency:    "EUR",
		Description: "Test payment",
	}

	httpReq := httptest.NewRequest("GET", "/api/v1/payments?id=payment_123", nil)
	w := httptest.NewRecorder()

	mockService.On("GetPayment", httpReq.Context(), "payment_123").Return(expectedPayment, nil)

	handler.GetPayment(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response payment.Payment
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedPayment.ID, response.ID)
	mockService.AssertExpectations(t)
}

func TestHandleWebhookHandler(t *testing.T) {
	mockService := new(mockPaymentService)
	handler := NewPaymentHandler(mockService)

	webhookReq := &payment.WebhookRequest{
		PaymentID: "payment_123",
		Status:    payment.PaymentStatusPaid,
	}

	body, _ := json.Marshal(webhookReq)
	httpReq := httptest.NewRequest("POST", "/api/v1/payments/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockService.On("HandleWebhook", httpReq.Context(), webhookReq).Return(nil)

	handler.HandleWebhook(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCancelPaymentHandler(t *testing.T) {
	mockService := new(mockPaymentService)
	handler := NewPaymentHandler(mockService)

	httpReq := httptest.NewRequest("POST", "/api/v1/payments/cancel?id=payment_123", nil)
	w := httptest.NewRecorder()

	mockService.On("CancelPayment", httpReq.Context(), "payment_123").Return(nil)

	handler.CancelPayment(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
