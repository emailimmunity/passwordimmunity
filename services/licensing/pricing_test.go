package licensing

import (
	"testing"
	"github.com/shopspring/decimal"
)

func TestCalculatePricing(t *testing.T) {
	svc := GetService()

	tests := []struct {
		name     string
		features []string
		bundles  []string
		currency string
		want     *PricingDetails
		wantErr  error
	}{
		{
			name:     "Valid features and bundles",
			features: []string{"advanced_sso", "custom_roles"},
			bundles:  []string{"business_bundle"},
			currency: "USD",
			want: &PricingDetails{
				TotalAmount: decimal.NewFromFloat(249.97), // 99.99 + 99.99 + 49.99
				FeaturePrices: map[string]decimal.Decimal{
					"advanced_sso":  decimal.NewFromFloat(99.99),
					"custom_roles":  decimal.NewFromFloat(99.99),
				},
				BundlePrices: map[string]decimal.Decimal{
					"business_bundle": decimal.NewFromFloat(49.99),
				},
				Currency: "USD",
			},
			wantErr: nil,
		},
		{
			name:     "Invalid currency",
			features: []string{"advanced_sso"},
			bundles:  []string{},
			currency: "INVALID",
			want:     nil,
			wantErr:  ErrInvalidCurrency,
		},
		{
			name:     "Free features only",
			features: []string{"basic_auth"},
			bundles:  []string{},
			currency: "USD",
			want: &PricingDetails{
				TotalAmount:   decimal.Zero,
				FeaturePrices: map[string]decimal.Decimal{},
				BundlePrices:  map[string]decimal.Decimal{},
				Currency:      "USD",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CalculatePricing(tt.features, tt.bundles, tt.currency)

			if err != tt.wantErr {
				t.Errorf("CalculatePricing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				return
			}

			if !got.TotalAmount.Equal(tt.want.TotalAmount) {
				t.Errorf("TotalAmount = %v, want %v", got.TotalAmount, tt.want.TotalAmount)
			}

			if got.Currency != tt.want.Currency {
				t.Errorf("Currency = %v, want %v", got.Currency, tt.want.Currency)
			}

			for feature, price := range tt.want.FeaturePrices {
				if !got.FeaturePrices[feature].Equal(price) {
					t.Errorf("FeaturePrice[%s] = %v, want %v", feature, got.FeaturePrices[feature], price)
				}
			}

			for bundle, price := range tt.want.BundlePrices {
				if !got.BundlePrices[bundle].Equal(price) {
					t.Errorf("BundlePrice[%s] = %v, want %v", bundle, got.BundlePrices[bundle], price)
				}
			}
		})
	}
}
