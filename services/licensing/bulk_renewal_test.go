package licensing

import (
	"context"
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestRenewLicensesBulk(t *testing.T) {
	svc := GetService()
	ctx := context.Background()
	now := time.Now()

	// Setup test licenses
	testLicenses := map[string]*License{
		"org1": {
			ID:             "lic_1",
			OrganizationID: "org1",
			Features:       []string{"advanced_sso"},
			Bundles:        []string{"business_bundle"},
			ExpiresAt:      now.Add(15 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(149.98),
		},
		"org2": {
			ID:             "lic_2",
			OrganizationID: "org2",
			Features:       []string{"custom_roles"},
			Bundles:        []string{"business_bundle"},
			ExpiresAt:      now.Add(10 * 24 * time.Hour),
			Status:         "active",
			Amount:         decimal.NewFromFloat(149.98),
		},
	}

	for _, license := range testLicenses {
		svc.licenses[license.OrganizationID] = license
	}

	tests := []struct {
		name    string
		req     *BulkRenewalRequest
		wantErr error
	}{
		{
			name: "Successful bulk renewal",
			req: &BulkRenewalRequest{
				OrganizationIDs: []string{"org1", "org2"},
				PaymentID:       "pay_bulk_1",
				Amount:         decimal.NewFromFloat(299.96),
				Currency:       "USD",
				Duration:       365 * 24 * time.Hour,
			},
			wantErr: nil,
		},
		{
			name: "Insufficient payment",
			req: &BulkRenewalRequest{
				OrganizationIDs: []string{"org1", "org2"},
				PaymentID:       "pay_bulk_2",
				Amount:         decimal.NewFromFloat(100.00),
				Currency:       "USD",
				Duration:       365 * 24 * time.Hour,
			},
			wantErr: ErrInsufficientPayment,
		},
		{
			name: "Mixed valid and invalid organizations",
			req: &BulkRenewalRequest{
				OrganizationIDs: []string{"org1", "nonexistent"},
				PaymentID:       "pay_bulk_3",
				Amount:         decimal.NewFromFloat(299.96),
				Currency:       "USD",
				Duration:       365 * 24 * time.Hour,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.RenewLicensesBulk(ctx, tt.req)

			if err != tt.wantErr {
				t.Errorf("RenewLicensesBulk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				// Check successful renewals
				for _, orgID := range tt.req.OrganizationIDs {
					if _, exists := testLicenses[orgID]; exists {
						if _, ok := result.SuccessfulRenewals[orgID]; !ok {
							t.Errorf("Expected successful renewal for %s", orgID)
						}
					} else {
						if _, ok := result.FailedRenewals[orgID]; !ok {
							t.Errorf("Expected failed renewal for %s", orgID)
						}
					}
				}

				// Verify renewed licenses
				for orgID, license := range result.SuccessfulRenewals {
					if license.PaymentID != tt.req.PaymentID {
						t.Errorf("PaymentID = %v, want %v", license.PaymentID, tt.req.PaymentID)
					}

					expectedExpiry := time.Now().Add(tt.req.Duration)
					if license.ExpiresAt.Sub(expectedExpiry) > time.Minute {
						t.Errorf("ExpiresAt = %v, want close to %v", license.ExpiresAt, expectedExpiry)
					}
				}
			}
		})
	}
}
