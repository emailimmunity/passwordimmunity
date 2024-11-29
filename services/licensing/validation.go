package licensing

import (
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/shopspring/decimal"
)

// ValidateLicensePayment checks if the license payment details are valid
func ValidateLicensePayment(license *License) error {
	// Basic validation
	if license == nil {
		return ErrLicenseNotFound
	}

	if license.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	if !config.IsSupportedCurrency(license.Currency) {
		return ErrInvalidCurrency
	}

	// Calculate expected amount based on features and bundles
	expectedAmount := calculateExpectedAmount(license.Features, license.Bundles, license.Currency)

	// Allow for small decimal differences due to currency conversion
	difference := license.Amount.Sub(expectedAmount).Abs()
	tolerance := decimal.NewFromFloat(0.01)

	if difference.GreaterThan(tolerance) {
		return ErrInvalidAmount
	}

	return nil
}

func calculateExpectedAmount(features []string, bundles []string, currency string) decimal.Decimal {
	total := decimal.Zero

	// Add up feature costs
	for _, feature := range features {
		price := config.GetFeaturePrice(feature, currency)
		total = total.Add(price)
	}

	// Add up bundle costs
	for _, bundle := range bundles {
		price := config.GetBundlePrice(bundle, currency)
		total = total.Add(price)
	}

	return total
}
