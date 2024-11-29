package payment

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type mockPaymentProvider struct {
	createPaymentFunc func(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	verifyPaymentFunc func(ctx context.Context, paymentID string) (*PaymentResponse, error)
}

func (m *mockPaymentProvider) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	return m.createPaymentFunc(ctx, req)
}

func (m *mockPaymentProvider) VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return m.verifyPaymentFunc(ctx, paymentID)
}

type mockNotificationService struct {
	successNotifications []string
	failureNotifications []string
}

func (m *mockNotificationService) NotifyPaymentSuccess(ctx context.Context, orgID, eventType string) {
	m.successNotifications = append(m.successNotifications, fmt.Sprintf("%s:%s", orgID, eventType))
}

func (m *mockNotificationService) NotifyPaymentFailure(ctx context.Context, orgID, eventType, reason string) {
	m.failureNotifications = append(m.failureNotifications, fmt.Sprintf("%s:%s:%s", orgID, eventType, reason))
}

func TestNewPaymentService(t *testing.T) {
	provider := &mockPaymentProvider{}
	notifications := &mockNotificationService{}

	service := NewPaymentService(provider)
	if service.provider == nil {
		t.Error("Expected provider to be set")
	}

	service.notifications = notifications
	if service.notifications == nil {
		t.Error("Expected notifications to be set")
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
			name:    "valid amount",
			amount:  Amount{Currency: "EUR", Value: "100.00"},
			wantErr: false,
		},
		{
			name:    "below minimum",
			amount:  Amount{Currency: "EUR", Value: "0.50"},
			wantErr: true,
			errMsg:  "must be at least",
		},
		{
			name:    "negative amount",
			amount:  Amount{Currency: "USD", Value: "-10.00"},
			wantErr: true,
			errMsg:  "cannot be negative",
		},
		{
			name:    "invalid currency",
			amount:  Amount{Currency: "XXX", Value: "100.00"},
			wantErr: true,
			errMsg:  "oneof",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.amount.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Amount.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Amount.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPaymentService_ProcessPayment(t *testing.T) {
	tests := []struct {
		name    string
		req     PaymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid monthly billing period",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Duration:      "monthly",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid billing period",
			req: PaymentRequest{
				Amount:      Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:      "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid billing period",
		},
		{
			name: "valid EUR amount",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Bundles:        []string{"security"},
					Features:       []string{"2fa", "sso"},
					Duration:       "monthly",
				},
			},
			wantErr: false,
		},
		{
			name: "valid USD amount",
			req: PaymentRequest{
				Amount: Amount{Currency: "USD", Value: "150.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "monthly",
				},
			},
			wantErr: false,
		},
		{
			name: "valid GBP amount",
			req: PaymentRequest{
				Amount: Amount{Currency: "GBP", Value: "75.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "yearly",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid amount format",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "abc"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "monthly",
				},
			},
			wantErr: true,
			errMsg:  "invalid amount value",
		},
		{
			name: "negative amount",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "-10.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "monthly",
				},
			},
			wantErr: true,
			errMsg:  "invalid amount value",
		},
		{
			name: "invalid currency",
			req: PaymentRequest{
				Amount: Amount{Currency: "XXX", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "monthly",
				},
			},
			wantErr: true,
			errMsg:  "invalid currency",
		},
		{
			name: "amount too low",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "0.50"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "monthly",
				},
			},
			wantErr: true,
			errMsg:  "must be at least",
		},
		{
			name: "missing organization ID",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					Features: []string{"2fa"},
					Duration: "monthly",
				},
			},
			wantErr: true,
			errMsg:  "organization_id",
		},
		{
			name: "no features or bundles",
			req: PaymentRequest{
				Amount: Amount{Currency: "EUR", Value: "100.00"},
				Description: "Enterprise Bundle",
				RedirectURL: "https://example.com/return",
				WebhookURL:  "https://example.com/webhook",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Duration:       "monthly",
				},
			},
			wantErr: true,
			errMsg:  "at least one feature or bundle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockPaymentProvider{
				createPaymentFunc: func(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
					return &PaymentResponse{
						ID:       "test_payment",
						Status:   "open",
						Metadata: req.Metadata,
					}, nil
				},
			}

			service := NewPaymentService(provider)
			resp, err := service.ProcessPayment(context.Background(), tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: ProcessPayment() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("%s: ProcessPayment() error = %v, want error containing %v", tt.name, err, tt.errMsg)
				}
				return
			}

			if !tt.wantErr && resp.Metadata.OrganizationID != tt.req.Metadata.OrganizationID {
				t.Errorf("%s: Expected org ID %s, got %s", tt.name, tt.req.Metadata.OrganizationID, resp.Metadata.OrganizationID)
			}
		})
	}
}

func TestPaymentService_VerifyAndActivate(t *testing.T) {
	tests := []struct {
		name             string
		paymentID        string
		mockResponse     *PaymentResponse
		mockError        error
		wantErr          bool
		errContains      string
		wantNotification bool
	}{
		{
			name:      "successful verification",
			paymentID: "test_payment",
			mockResponse: &PaymentResponse{
				ID:     "test_payment",
				Status: "paid",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa", "sso"},
					Bundles:        []string{"security"},
					Duration:       "720h",
				},
			},
			wantErr:          false,
			wantNotification: true,
		},
		{
			name:          "verification failure",
			paymentID:     "failed_payment",
			mockError:     fmt.Errorf("payment verification failed"),
			wantErr:       true,
			errContains:   "failed to verify payment",
			wantNotification: true,
		},
		{
			name:      "invalid payment status",
			paymentID: "pending_payment",
			mockResponse: &PaymentResponse{
				ID:     "pending_payment",
				Status: "pending",
				Metadata: PaymentMetadata{
					OrganizationID: "org_test",
					Features:       []string{"2fa"},
					Duration:       "720h",
				},
			},
			wantErr:          true,
			errContains:      "payment not completed",
			wantNotification: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockPaymentProvider{
				verifyPaymentFunc: func(ctx context.Context, paymentID string) (*PaymentResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			service := NewPaymentService(provider)
			err := service.VerifyAndActivate(context.Background(), tt.paymentID)

			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyAndActivate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
			}

			if !tt.wantErr {
				licensingSvc := licensing.GetService()
				for _, bundle := range tt.mockResponse.Metadata.Bundles {
					if !licensingSvc.HasBundleAccess(tt.mockResponse.Metadata.OrganizationID, bundle) {
						t.Errorf("Bundle %q not activated after successful payment", bundle)
					}
				}
			}
		})
	}
}

// contains checks if a string contains a substring, case-insensitive
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
