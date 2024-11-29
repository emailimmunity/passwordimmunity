package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestEnvironment(t *testing.T) (*PaymentService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/payments":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"id": "tr_test",
				"status": "open",
				"amount": {"currency": "EUR", "value": "10.00"},
				"metadata": {"featureId": "test_feature", "userId": "test_user"}
			}`))
		case "/v2/payments/tr_test":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "tr_test",
				"status": "paid",
				"amount": {"currency": "EUR", "value": "10.00"},
				"metadata": {"featureId": "test_feature", "userId": "test_user"}
			}`))
		}
	}))

	mollieService := &MollieService{
		client:     &http.Client{},
		apiKey:     "test_key",
		webhookURL: "https://test.com/webhook",
		baseURL:    server.URL + "/v2",
	}

	return NewPaymentService(mollieService), server
}

func TestPaymentIntegration(t *testing.T) {
	service, server := setupTestEnvironment(t)
	defer server.Close()

	// Test complete payment flow
	req := PaymentRequest{
		Amount: Amount{
			Currency: "EUR",
			Value:    "10.00",
		},
		Description: "Test payment",
		RedirectURL: "https://test.com/return",
		WebhookURL:  "https://test.com/webhook",
	}
	req.Metadata.FeatureID = "test_feature"
	req.Metadata.UserID = "test_user"

	// Create payment
	payment, err := service.ProcessPayment(context.Background(), req)
	if err != nil {
		t.Fatalf("ProcessPayment failed: %v", err)
	}

	if payment.ID != "tr_test" {
		t.Errorf("Expected payment ID tr_test, got %s", payment.ID)
	}

	// Verify and activate
	err = service.VerifyAndActivate(context.Background(), payment.ID)
	if err != nil {
		t.Fatalf("VerifyAndActivate failed: %v", err)
	}
}

func TestPaymentIntegrationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": 400, "message": "Invalid request"}`))
	}))
	defer server.Close()

	mollieService := &MollieService{
		client:     &http.Client{},
		apiKey:     "test_key",
		webhookURL: "https://test.com/webhook",
		baseURL:    server.URL + "/v2",
	}

	service := NewPaymentService(mollieService)

	req := PaymentRequest{
		Amount: Amount{
			Currency: "EUR",
			Value:    "10.00",
		},
	}

	_, err := service.ProcessPayment(context.Background(), req)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
