package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/emailimmunity/passwordimmunity/config"
)

// Mock payment service
type mockPaymentService struct {
	sessions map[string]*payment.Session
}

func (m *mockPaymentService) CreateSession(ctx context.Context, req payment.SessionRequest) (*payment.Session, error) {
	session := &payment.Session{
		ID:         "test_session",
		PaymentURL: "https://test.mollie.com/pay/test_session",
		Status:     "pending",
	}
	m.sessions[session.ID] = session
	return session, nil
}

func TestHandleFeatureActivation(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		request        ActivationRequest
		wantStatus     int
		wantSuccess    bool
		wantPaymentURL bool
		wantPrice      float64
	}{
		{
			name:  "new feature activation with payment",
			orgID: "org1",
			request: ActivationRequest{
				FeatureID: "advanced_sso",
			},
			wantStatus:     http.StatusOK,
			wantSuccess:    false,
			wantPaymentURL: true,
			wantPrice:      99.99,
		},
		{
			name:  "bundle activation with payment",
			orgID: "org1",
			request: ActivationRequest{
				FeatureID: "advanced_sso",
				BundleID:  "security",
			},
			wantStatus:     http.StatusOK,
			wantSuccess:    false,
			wantPaymentURL: true,
			wantPrice:      299.99,
		},
		{
			name:  "already active feature",
			orgID: "org_with_feature",
			request: ActivationRequest{
				FeatureID: "advanced_sso",
			},
			wantStatus:     http.StatusOK,
			wantSuccess:    true,
			wantPaymentURL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services
			mockPayment := &mockPaymentService{
				sessions: make(map[string]*payment.Session),
			}
			payment.SetService(mockPayment)

			// Setup test license for org_with_feature
			if tt.orgID == "org_with_feature" {
				svc := licensing.GetService()
				_, _ = svc.ActivateLicense(context.Background(), tt.orgID,
					[]string{"advanced_sso"},
					[]string{},
					30*24*60*60*1000000000)
			}

			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/features/activate", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			HandleFeatureActivation(rr, req)

			// Check status code
			if rr.Code != tt.wantStatus {
				t.Errorf("HandleFeatureActivation() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			// Parse response
			var response ActivationResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check response fields
			if response.Success != tt.wantSuccess {
				t.Errorf("HandleFeatureActivation() success = %v, want %v", response.Success, tt.wantSuccess)
			}

			if tt.wantPaymentURL && response.PaymentURL == "" {
				t.Error("Expected payment URL in response")
			}

			if !tt.wantPaymentURL && response.PaymentURL != "" {
				t.Error("Unexpected payment URL in response")
			}
		})
	}
}
