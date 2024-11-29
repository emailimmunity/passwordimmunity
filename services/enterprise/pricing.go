package enterprise

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/emailimmunity/passwordimmunity/config"
)

type PricingManager struct {
	baseCurrency string
	rates        map[string]decimal.Decimal
}

func NewPricingManager(baseCurrency string) *PricingManager {
	return &PricingManager{
		baseCurrency: baseCurrency,
		rates: map[string]decimal.Decimal{
			"USD": decimal.NewFromFloat(1.0),
			"EUR": decimal.NewFromFloat(0.9),
		},
	}
}

func (p *PricingManager) ConvertPrice(amount decimal.Decimal, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	fromRate, exists := p.rates[fromCurrency]
	if !exists {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", fromCurrency)
	}

	toRate, exists := p.rates[toCurrency]
	if !exists {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", toCurrency)
	}

	// Convert to base currency first, then to target currency
	inBase := amount.Div(fromRate)
	return inBase.Mul(toRate), nil
}

func (p *PricingManager) GetFeaturePrice(featureID, currency string) (decimal.Decimal, error) {
	price, features, err := config.GetBundlePrice("enterprise_starter", p.baseCurrency)
	if err != nil {
		return decimal.Zero, err
	}

	// Calculate per-feature price (simplified)
	perFeature := price.Div(decimal.NewFromInt(int64(len(features))))

	return p.ConvertPrice(perFeature, p.baseCurrency, currency)
}

func (p *PricingManager) ValidatePayment(amount decimal.Decimal, currency string, bundleID string) error {
	expectedPrice, _, err := config.GetBundlePrice(bundleID, currency)
	if err != nil {
		return err
	}

	if !amount.Equal(expectedPrice) {
		return fmt.Errorf("invalid payment amount for bundle %s: expected %s %s, got %s %s",
			bundleID, expectedPrice, currency, amount, currency)
	}

	return nil
}
