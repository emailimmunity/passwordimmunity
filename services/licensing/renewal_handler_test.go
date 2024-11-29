package licensing

import (
	"context"
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestRenewLicense(t *testing.T) {
	svc := GetService()
	ctx := context.Background()
	now := time.Now()

	// Create initial test license
	existingLicense := &License{
		ID:             "lic_test",
		OrganizationID: "org_test",
		Features:       []string{"advanced_sso", "custom_roles"},
		Bundles:        []string{"business_bundle"},
		IssuedAt:       now.Add(-330 * 24 * time.Hour),
		ExpiresAt:      now.Add(30 * 24 * time.Hour),
		Status:         "active",
		PaymentID:      "pay_old",
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(249.97),
	}
	svc.licenses[existingLicense.OrganizationID] = existingLicense

	tests := []struct {
		name    string
		req     *RenewalRequest
		wantErr error
	}{
		{
			name: "Valid renewal",
			req: &RenewalRequest{
				OrganizationID: "org_test",
				PaymentID:      "pay_new",
				Amount:         decimal.NewFromFloat(249.97),
				Currency:       "USD",
				Duration:      365 * 24 * time.Hour,
			},
			wantErr: nil,
		},
		{
			name: "Insufficient payment",
			req: &RenewalRequest{
				OrganizationID: "org_test",
				PaymentID:      "pay_insufficient",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				Duration:      365 * 24 * time.Hour,
			},
			wantErr: ErrInsufficientPayment,
		},
		{
			name: "Invalid currency",
			req: &RenewalRequest{
				OrganizationID: "org_test",
				PaymentID:      "pay_invalid",
				Amount:         decimal.NewFromFloat(249.97),
				Currency:       "INVALID",
				Duration:      365 * 24 * time.Hour,
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "Non-existent license",
			req: &RenewalRequest{
				OrganizationID: "nonexistent",
				PaymentID:      "pay_nonexistent",
				Amount:         decimal.NewFromFloat(249.97),
				Currency:       "USD",
				Duration:      365 * 24 * time.Hour,
			},
			wantErr: ErrLicenseNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.RenewLicense(ctx, tt.req)

			if err != tt.wantErr {
				t.Errorf("RenewLicense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				if got == nil {
					t.Error("RenewLicense() returned nil license for successful renewal")
					return
				}

				if got.PaymentID != tt.req.PaymentID {
					t.Errorf("PaymentID = %v, want %v", got.PaymentID, tt.req.PaymentID)
				}

				if !got.Amount.Equal(tt.req.Amount) {
					t.Errorf("Amount = %v, want %v", got.Amount, tt.req.Amount)
				}

				if got.Status != "active" {
					t.Errorf("Status = %v, want active", got.Status)
				}

				expectedExpiry := time.Now().Add(tt.req.Duration)
				if got.ExpiresAt.Sub(expectedExpiry) > time.Minute {
					t.Errorf("ExpiresAt = %v, want close to %v", got.ExpiresAt, expectedExpiry)
				}
			}
		})
	}
}
