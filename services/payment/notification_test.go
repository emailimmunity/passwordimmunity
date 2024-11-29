package payment

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type MockEmailService struct {
	sentEmails []EmailMessage
	shouldFail bool
}

func (m *MockEmailService) SendEmail(ctx context.Context, msg EmailMessage) error {
	if m.shouldFail {
		return fmt.Errorf("mock email error")
	}
	m.sentEmails = append(m.sentEmails, msg)
	return nil
}

func TestNotificationService(t *testing.T) {
	tests := []struct {
		name         string
		mockEmail    *MockEmailService
		ctx          context.Context
		orgEmail     string
		metadata     interface{}
		purchaseType string
		wantErr      bool
		checkEmail   func(t *testing.T, email EmailMessage)
	}{
		{
			name:      "successful bundle notification",
			mockEmail: &MockEmailService{},
			ctx: context.WithValue(context.Background(), "payment_metadata", struct {
				Features []string
				Bundles  []string
				Duration string
			}{
				Features: []string{"2fa", "sso"},
				Bundles:  []string{"security"},
				Duration: "30 days",
			}),
			orgEmail:     "org@example.com",
			purchaseType: "bundle_purchase",
			wantErr:      false,
			checkEmail: func(t *testing.T, email EmailMessage) {
				if !strings.Contains(email.Body, "2fa") {
					t.Error("Email missing feature information")
				}
				if !strings.Contains(email.Body, "security") {
					t.Error("Email missing bundle information")
				}
				if !strings.Contains(email.Body, "30 days") {
					t.Error("Email missing duration information")
				}
			},
		},
		{
			name:      "email service failure",
			mockEmail: &MockEmailService{shouldFail: true},
			ctx: context.WithValue(context.Background(), "payment_metadata", struct {
				Features []string
				Bundles  []string
				Duration string
			}{
				Features: []string{"2fa"},
				Bundles:  []string{"basic"},
				Duration: "30 days",
			}),
			orgEmail:     "org@example.com",
			purchaseType: "bundle_purchase",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewNotificationService(tt.mockEmail)
			err := service.NotifyPaymentSuccess(tt.ctx, tt.orgEmail, tt.purchaseType)

			if (err != nil) != tt.wantErr {
				t.Errorf("NotifyPaymentSuccess() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(tt.mockEmail.sentEmails) != 1 {
				t.Errorf("Expected 1 email to be sent, got %d", len(tt.mockEmail.sentEmails))
			}

			if !tt.wantErr && tt.checkEmail != nil && len(tt.mockEmail.sentEmails) > 0 {
				tt.checkEmail(t, tt.mockEmail.sentEmails[0])
			}
		})
	}
}
