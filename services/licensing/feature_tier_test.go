package licensing

import (
	"testing"
	"github.com/shopspring/decimal"
)

func TestGetFeatureTier(t *testing.T) {
	svc := GetService()

	tests := []struct {
		name      string
		featureID string
		want      *FeatureTier
	}{
		{
			name:      "Enterprise feature",
			featureID: "advanced_sso",
			want: &FeatureTier{
				RequiresPayment: true,
				MinimumAmount:   decimal.NewFromFloat(99.99),
				IsEnterprise:    true,
			},
		},
		{
			name:      "Free feature",
			featureID: "basic_auth",
			want: &FeatureTier{
				RequiresPayment: false,
				MinimumAmount:   decimal.Zero,
				IsEnterprise:    false,
			},
		},
		{
			name:      "Bundle feature",
			featureID: "role_management",
			want: &FeatureTier{
				RequiresPayment: true,
				MinimumAmount:   decimal.NewFromFloat(49.99),
				BundleID:        "business_bundle",
				IsEnterprise:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GetFeatureTier(tt.featureID)

			if got.RequiresPayment != tt.want.RequiresPayment {
				t.Errorf("RequiresPayment = %v, want %v", got.RequiresPayment, tt.want.RequiresPayment)
			}
			if !got.MinimumAmount.Equal(tt.want.MinimumAmount) {
				t.Errorf("MinimumAmount = %v, want %v", got.MinimumAmount, tt.want.MinimumAmount)
			}
			if got.BundleID != tt.want.BundleID {
				t.Errorf("BundleID = %v, want %v", got.BundleID, tt.want.BundleID)
			}
			if got.IsEnterprise != tt.want.IsEnterprise {
				t.Errorf("IsEnterprise = %v, want %v", got.IsEnterprise, tt.want.IsEnterprise)
			}
		})
	}
}
