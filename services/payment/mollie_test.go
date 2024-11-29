package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

func TestMollieService_CreatePayment(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test_key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id": "tr_test",
			"status": "open",
			"createdAt": "2023-01-01T12:00:00Z",
			"amount": {
				"currency": "EUR",
				"value": "100.00"
			},
			"description": "Enterprise Bundle",
			"redirectUrl": "https://example.com/redirect",
			"webhookUrl": "https://example.com/webhook",
			"metadata": {
				"organizationId": "org_test",
				"bundles": ["security"],
				"features": ["2fa", "sso"],
				"duration": "720h"
			}
		}`))
	}))
	defer server.Close()

	tests := []struct {
		name    string
		req     PaymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid amount - negative",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "-100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "amount value cannot be negative",
		},
		{
			name: "invalid currency",
			req: PaymentRequest{
				Amount: Amount{Currency: "XXX", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "invalid currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MollieService{
				client:     &http.Client{Timeout: 10 * time.Second},
				apiKey:     "test_key",
				webhookURL: "https://example.com/webhook",
				baseURL:    server.URL,
			}

			_, err := service.CreatePayment(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("CreatePayment() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestMollieService_VerifyPayment(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		wantErr    bool
		errContains string
	}{
		{
			name: "valid payment",
			response: `{
				"id": "tr_test",
				"status": "paid",
				"createdAt": "2023-01-01T12:00:00Z",
				"amount": {
					"currency": "EUR",
					"value": "100.00"
				},
				"description": "Enterprise Bundle",
				"redirectUrl": "https://example.com/redirect",
				"webhookUrl": "https://example.com/webhook",
				"metadata": {
					"organizationId": "org_test",
					"bundles": ["security"],
					"features": ["2fa", "sso"],
					"duration": "720h"
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid status",
			response: `{
				"id": "tr_test",
				"status": "invalid",
				"createdAt": "2023-01-01T12:00:00Z",
				"amount": {
					"currency": "EUR",
					"value": "100.00"
				},
				"description": "Enterprise Bundle",
				"redirectUrl": "https://example.com/redirect",
				"webhookUrl": "https://example.com/webhook",
				"metadata": {
					"organizationId": "org_test",
					"bundles": ["security"],
					"features": ["2fa", "sso"],
					"duration": "720h"
				}
			}`,
			wantErr: true,
			errContains: "oneof",
		},
		{
			name: "missing required fields",
			response: `{
				"id": "tr_test",
				"status": "paid"
			}`,
			wantErr: true,
			errContains: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test_key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			service := &MollieService{
				client:  &http.Client{Timeout: 10 * time.Second},
				apiKey:  "test_key",
				baseURL: server.URL,
			}

			resp, err := service.VerifyPayment(context.Background(), "tr_test")
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("VerifyPayment() error = %v, want error containing %v", err, tt.errContains)
			}
			if err == nil && resp.Status != "paid" {
				t.Errorf("Expected status paid, got %s", resp.Status)
			}
		})
	}
}

func TestPaymentResponse_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payment PaymentResponse
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid payment",
			payment: PaymentResponse{
				ID:          "tr_test",
				Status:      "paid",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			payment: PaymentResponse{
				ID:          "tr_test",
				Status:      "invalid",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "oneof",
		},
		{
			name: "invalid amount",
			payment: PaymentResponse{
				ID:          "tr_test",
				Status:      "paid",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "-100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "amount value cannot be negative",
		},
		{
			name: "missing required fields",
			payment: PaymentResponse{
				ID:     "tr_test",
				Status: "paid",
			},
			wantErr: true,
			errMsg:  "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PaymentResponse.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("PaymentResponse.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestPaymentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PaymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr: false,
		},
		{
			name: "missing features and bundles",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "at least one feature or bundle must be specified",
		},
		{
			name: "invalid duration",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid duration",
		},
		{
			name: "negative duration",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_1",
					Features:       []string{"2fa"},
					Duration:       "-24h",
				},
			},
			wantErr: true,
			errMsg:  "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PaymentRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("PaymentRequest.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestAmount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		amount  Amount
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid EUR amount",
			amount:  Amount{Currency: "EUR", Value: "100.00"},
			wantErr: false,
		},
		{
			name:    "valid USD amount",
			amount:  Amount{Currency: "USD", Value: "50.00"},
			wantErr: false,
		},
		{
			name:    "invalid currency",
			amount:  Amount{Currency: "XXX", Value: "100.00"},
			wantErr: true,
			errMsg:  "oneof",
		},
		{
			name:    "negative amount",
			amount:  Amount{Currency: "EUR", Value: "-10.00"},
			wantErr: true,
			errMsg:  "amount value cannot be negative",
		},
		{
			name:    "below minimum amount",
			amount:  Amount{Currency: "EUR", Value: "0.50"},
			wantErr: true,
			errMsg:  "amount must be at least",
		},
		{
			name:    "invalid amount format",
			amount:  Amount{Currency: "EUR", Value: "invalid"},
			wantErr: true,
			errMsg:  "invalid amount value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.amount.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Amount.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Amount.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestMollieService_HandleWebhook(t *testing.T) {
	tests := []struct {
		name    string
		payment *PaymentResponse
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid payment",
			payment: &PaymentResponse{
				ID:          "tr_test",
				Status:      "paid",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Features:       []string{"2fa", "sso"},
					Duration:       "720h",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil payment",
			payment: nil,
			wantErr: true,
			errMsg:  "payment response is nil",
		},
		{
			name: "invalid payment status",
			payment: &PaymentResponse{
				ID:          "tr_test",
				Status:      "pending",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Features:       []string{"2fa", "sso"},
					Duration:       "720h",
				},
			},
			wantErr: true,
			errMsg:  "payment not completed",
		},
		{
			name: "invalid duration",
			payment: &PaymentResponse{
				ID:          "tr_test",
				Status:      "paid",
				CreatedAt:   time.Now(),
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Test Payment",
				RedirectURL: "https://example.com/redirect",
				WebhookURL:  "https://example.com/webhook",
				Metadata: struct {
					OrganizationID string   `json:"organizationId"`
					Features       []string `json:"features"`
					Bundles        []string `json:"bundles"`
					Duration       string   `json:"duration"`
				}{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewMollieService("test_key", "https://example.com/webhook")
			err := service.HandleWebhook(tt.payment)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("HandleWebhook() error = %v, want error containing %v", err, tt.errMsg)
			}
			if err == nil {
				licensingSvc := licensing.GetService()
				if !licensingSvc.HasBundleAccess(tt.payment.Metadata.OrganizationID, tt.payment.Metadata.Bundles[0]) {
					t.Error("Bundle not activated after payment")
				}
			}
		})
	}
}
