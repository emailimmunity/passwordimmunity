package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookHandler_HandleWebhook(t *testing.T) {
	mockService := NewPaymentService(&MockPaymentProvider{
		verifyPaymentFunc: func(ctx context.Context, paymentID string) (*PaymentResponse, error) {
			return &PaymentResponse{
				ID:     paymentID,
				Status: "paid",
				Metadata: PaymentMetadata{
					Features: []string{"2fa", "sso"},
					Bundles:  []string{"enterprise"},
					Duration: "30 days",
				},
			}, nil
		},
		handleWebhookFunc: func(payment *PaymentResponse) error {
			return nil
		},
	})

	handler := NewWebhookHandler(mockService)

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
	}{
		{
			name:   "valid webhook",
			method: http.MethodPost,
			body: WebhookRequest{
				ID: "test_payment",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid body",
			method:     http.MethodPost,
			body:       "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				if err := json.NewEncoder(&body).Encode(tt.body); err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(tt.method, "/webhook", &body)
			rec := httptest.NewRecorder()

			handler.HandleWebhook(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleWebhook() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				// Wait for async processing
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
}

func TestWebhookHandler_DuplicateProcessing(t *testing.T) {
	processCount := 0
	mockService := NewPaymentService(&MockPaymentProvider{
		verifyPaymentFunc: func(ctx context.Context, paymentID string) (*PaymentResponse, error) {
			processCount++
			return &PaymentResponse{
				ID:     paymentID,
				Status: "paid",
			}, nil
		},
		handleWebhookFunc: func(payment *PaymentResponse) error {
			return nil
		},
	})

	handler := NewWebhookHandler(mockService)

	body := WebhookRequest{
		ID: "test_payment",
	}
	bodyBytes, _ := json.Marshal(body)

	// Send duplicate requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()
		handler.HandleWebhook(rec, req)
	}

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	if processCount > 1 {
		t.Errorf("Expected single processing, got %d", processCount)
	}
}

func TestWebhookHandler_FeatureActivation(t *testing.T) {
	activatedFeatures := make(map[string]bool)
	activatedBundles := make(map[string]bool)

	mockService := NewPaymentService(&MockPaymentProvider{
		verifyPaymentFunc: func(ctx context.Context, paymentID string) (*PaymentResponse, error) {
			return &PaymentResponse{
				ID:     paymentID,
				Status: "paid",
				Metadata: PaymentMetadata{
					Features: []string{"2fa", "sso"},
					Bundles:  []string{"enterprise"},
					Duration: "30 days",
				},
			}, nil
		},
		handleWebhookFunc: func(payment *PaymentResponse) error {
			for _, feature := range payment.Metadata.Features {
				activatedFeatures[feature] = true
			}
			for _, bundle := range payment.Metadata.Bundles {
				activatedBundles[bundle] = true
			}
			return nil
		},
	})

	handler := NewWebhookHandler(mockService)

	body := WebhookRequest{
		ID: "test_payment_features",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.HandleWebhook(rec, req)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	expectedFeatures := []string{"2fa", "sso"}
	expectedBundles := []string{"enterprise"}

	for _, feature := range expectedFeatures {
		if !activatedFeatures[feature] {
			t.Errorf("Expected feature %s to be activated", feature)
		}
	}

	for _, bundle := range expectedBundles {
		if !activatedBundles[bundle] {
			t.Errorf("Expected bundle %s to be activated", bundle)
		}
	}
}
