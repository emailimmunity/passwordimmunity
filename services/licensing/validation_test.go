package licensing

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestValidateLicensePayment(t *testing.T) {
	tests := []struct {
		name    string
		license *License
		wantErr error
	}{
		{
			name: "Valid license payment",
			license: &License{
				ID:             "lic_123",
				OrganizationID: "org_123",
				Features:       []string{"feature1"},
				Bundles:        []string{"bundle1"},
				IssuedAt:       time.Now(),
				ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
				Status:         "active",
				PaymentID:      "pay_123",
				Currency:       "USD",
				Amount:         decimal.NewFromFloat(100.00),
			},
			wantErr: nil,
		},
		{
			name:    "Nil license",
			license: nil,
			wantErr: ErrLicenseNotFound,
		},
		{
			name: "Invalid amount",
			license: &License{
				ID:             "lic_124",
				OrganizationID: "org_124",
				Features:       []string{"feature1"},
				Bundles:        []string{"bundle1"},
				IssuedAt:       time.Now(),
				ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
				Status:         "active",
				PaymentID:      "pay_124",
				Currency:       "USD",
				Amount:         decimal.Zero,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "Invalid currency",
			license: &License{
				ID:             "lic_125",
				OrganizationID: "org_125",
				Features:       []string{"feature1"},
				Bundles:        []string{"bundle1"},
				IssuedAt:       time.Now(),
				ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
				Status:         "active",
				PaymentID:      "pay_125",
				Currency:       "XXX",
				Amount:         decimal.NewFromFloat(100.00),
			},
			wantErr: ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLicensePayment(tt.license)
			if err != tt.wantErr {
				t.Errorf("ValidateLicensePayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
