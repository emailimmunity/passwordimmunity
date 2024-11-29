package licensing

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestGetRenewalStatus(t *testing.T) {
	svc := GetService()
	now := time.Now()

	tests := []struct {
		name    string
		license *License
		want    *RenewalStatus
	}{
		{
			name: "License expiring soon",
			license: &License{
				OrganizationID: "org_expiring",
				ExpiresAt:      now.Add(20 * 24 * time.Hour),
				Status:         "active",
				Amount:         decimal.NewFromFloat(100),
			},
			want: &RenewalStatus{
				NeedsRenewal:     true,
				DaysUntilExpiry:  20,
				InGracePeriod:    false,
				PaymentRequired:  false,
				RenewalAvailable: true,
			},
		},
		{
			name: "Expired license",
			license: &License{
				OrganizationID: "org_expired",
				ExpiresAt:      now.Add(-24 * time.Hour),
				Status:         "expired",
				Amount:         decimal.NewFromFloat(100),
			},
			want: &RenewalStatus{
				NeedsRenewal:     true,
				DaysUntilExpiry:  0,
				InGracePeriod:    true,
				PaymentRequired:  true,
				RenewalAvailable: true,
			},
		},
		{
			name: "Valid license not needing renewal",
			license: &License{
				OrganizationID: "org_valid",
				ExpiresAt:      now.Add(60 * 24 * time.Hour),
				Status:         "active",
				Amount:         decimal.NewFromFloat(100),
			},
			want: &RenewalStatus{
				NeedsRenewal:     false,
				DaysUntilExpiry:  60,
				InGracePeriod:    false,
				PaymentRequired:  false,
				RenewalAvailable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.license != nil {
				svc.licenses[tt.license.OrganizationID] = tt.license
			}

			got := svc.GetRenewalStatus(tt.license.OrganizationID)

			if got.NeedsRenewal != tt.want.NeedsRenewal {
				t.Errorf("NeedsRenewal = %v, want %v", got.NeedsRenewal, tt.want.NeedsRenewal)
			}

			if got.DaysUntilExpiry != tt.want.DaysUntilExpiry {
				t.Errorf("DaysUntilExpiry = %v, want %v", got.DaysUntilExpiry, tt.want.DaysUntilExpiry)
			}

			if got.InGracePeriod != tt.want.InGracePeriod {
				t.Errorf("InGracePeriod = %v, want %v", got.InGracePeriod, tt.want.InGracePeriod)
			}

			if got.PaymentRequired != tt.want.PaymentRequired {
				t.Errorf("PaymentRequired = %v, want %v", got.PaymentRequired, tt.want.PaymentRequired)
			}

			if got.RenewalAvailable != tt.want.RenewalAvailable {
				t.Errorf("RenewalAvailable = %v, want %v", got.RenewalAvailable, tt.want.RenewalAvailable)
			}
		})
	}
}
