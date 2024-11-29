package licensing

import (
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/shopspring/decimal"
)

// PricingDetails contains the breakdown of costs for features and bundles
type PricingDetails struct {
	TotalAmount     decimal.Decimal            `json:"total_amount"`
	FeaturePrices   map[string]decimal.Decimal `json:"feature_prices"`
	BundlePrices    map[string]decimal.Decimal `json:"bundle_prices"`
	Currency        string                     `json:"currency"`
}

// CalculatePricing returns the total cost and pricing details for requested features and bundles
func (s *Service) CalculatePricing(features []string, bundles []string, currency string) (*PricingDetails, error) {
	if !config.IsSupportedCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	details := &PricingDetails{
		TotalAmount:   decimal.Zero,
		FeaturePrices: make(map[string]decimal.Decimal),
		BundlePrices:  make(map[string]decimal.Decimal),
		Currency:      currency,
	}

	// Calculate feature prices
	for _, featureID := range features {
		tier := s.GetFeatureTier(featureID)
		if tier.RequiresPayment {
			details.FeaturePrices[featureID] = tier.MinimumAmount
			details.TotalAmount = details.TotalAmount.Add(tier.MinimumAmount)
		}
	}

	// Calculate bundle prices
	for _, bundleID := range bundles {
		price := config.GetBundlePrice(bundleID)
		details.BundlePrices[bundleID] = price
		details.TotalAmount = details.TotalAmount.Add(price)
	}

	return details, nil
}
