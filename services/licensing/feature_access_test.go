package licensing

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
)

func TestGetFeatureAccessStatus(t *testing.T) {
	svc := GetService()
	now := time.Now()
	futureTime := now.Add(30 * 24 * time.Hour)

	// Create test license
	license := &License{
		ID:             "lic_test",
		OrganizationID: "org_test",
		Features:       []string{"feature1", "feature2"},
		Bundles:        []string{"bundle1"},
		IssuedAt:       now,
		ExpiresAt:      futureTime,
		Status:         "active",
		PaymentID:      "pay_test",
		Currency:       "USD",
		Amount:         decimal.NewFromFloat(100.00),
	}
	svc.licenses[license.OrganizationID] = license

	tests := []struct {
		name      string
		orgID     string
		featureID string
		want      *FeatureAccessStatus
	}{
		{
			name:      "Feature exists and license valid",
			orgID:     "org_test",
			featureID: "feature1",
			want: &FeatureAccessStatus{
				HasAccess:     true,
				IsActive:      true,
				ExpiresAt:     futureTime,
				InGracePeriod: false,
				PaymentValid:  true,
			},
		},
		{
			name:      "Feature does not exist",
			orgID:     "org_test",
			featureID: "nonexistent",
			want: &FeatureAccessStatus{
				HasAccess:     false,
				IsActive:      true,
				ExpiresAt:     futureTime,
				InGracePeriod: false,
				PaymentValid:  true,
			},
		},
		{
			name:      "Organization does not exist",
			orgID:     "nonexistent",
			featureID: "feature1",
			want:      &FeatureAccessStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GetFeatureAccessStatus(tt.orgID, tt.featureID)

			if got.HasAccess != tt.want.HasAccess {
				t.Errorf("HasAccess = %v, want %v", got.HasAccess, tt.want.HasAccess)
			}
			if got.IsActive != tt.want.IsActive {
				t.Errorf("IsActive = %v, want %v", got.IsActive, tt.want.IsActive)
			}
			if !got.ExpiresAt.Equal(tt.want.ExpiresAt) {
				t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, tt.want.ExpiresAt)
			}
			if got.InGracePeriod != tt.want.InGracePeriod {
				t.Errorf("InGracePeriod = %v, want %v", got.InGracePeriod, tt.want.InGracePeriod)
			}
			if got.PaymentValid != tt.want.PaymentValid {
				t.Errorf("PaymentValid = %v, want %v", got.PaymentValid, tt.want.PaymentValid)
			}
		})
	}
}
