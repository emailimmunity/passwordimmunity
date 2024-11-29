package licensing

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestGetLicenseStatus(t *testing.T) {
	svc := GetService()
	now := time.Now()
	futureTime := now.Add(30 * 24 * time.Hour)

	// Create test license
	license := &License{
		ID:             "lic_test",
		OrganizationID: "org_test",
		Features:       []string{"advanced_sso", "custom_roles"},
		Bundles:        []string{"business_bundle"},
		IssuedAt:       now,
		ExpiresAt:      futureTime,
		Status:         "active",
		PaymentID:      "pay_test",
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(249.97),
	}
	svc.licenses[license.OrganizationID] = license

	tests := []struct {
		name  string
		orgID string
		want  *LicenseStatus
	}{
		{
			name:  "Active license with valid payment",
			orgID: "org_test",
			want: &LicenseStatus{
				IsActive:       true,
				ExpiresAt:      futureTime,
				ActiveFeatures: []string{"advanced_sso", "custom_roles"},
				ActiveBundles:  []string{"business_bundle"},
				PaymentStatus: PaymentStatus{
					LastPaymentID: "pay_test",
					Amount:        decimal.NewFromFloat(249.97),
					Currency:      "USD",
					ValidUntil:    futureTime,
				},
			},
		},
		{
			name:  "Non-existent organization",
			orgID: "nonexistent",
			want: &LicenseStatus{
				IsActive:       false,
				ActiveFeatures: []string{},
				ActiveBundles:  []string{},
				FeatureAccess:  make(map[string]AccessDetails),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GetLicenseStatus(tt.orgID)

			if got.IsActive != tt.want.IsActive {
				t.Errorf("IsActive = %v, want %v", got.IsActive, tt.want.IsActive)
			}

			if tt.orgID == "org_test" {
				if !got.ExpiresAt.Equal(tt.want.ExpiresAt) {
					t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, tt.want.ExpiresAt)
				}

				if len(got.ActiveFeatures) != len(tt.want.ActiveFeatures) {
					t.Errorf("ActiveFeatures length = %v, want %v", len(got.ActiveFeatures), len(tt.want.ActiveFeatures))
				}

				if got.PaymentStatus.LastPaymentID != tt.want.PaymentStatus.LastPaymentID {
					t.Errorf("LastPaymentID = %v, want %v", got.PaymentStatus.LastPaymentID, tt.want.PaymentStatus.LastPaymentID)
				}

				if !got.PaymentStatus.Amount.Equal(tt.want.PaymentStatus.Amount) {
					t.Errorf("Amount = %v, want %v", got.PaymentStatus.Amount, tt.want.PaymentStatus.Amount)
				}
			}
		})
	}
}
