package licensing

import "errors"

var (
	// ErrLicenseNotFound is returned when a license operation is attempted on a non-existent license
	ErrLicenseNotFound = errors.New("license not found")

	// ErrInvalidCurrency is returned when an unsupported currency is provided
	ErrInvalidCurrency = errors.New("invalid currency")

	// ErrInvalidAmount is returned when the amount is zero or negative
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrInvalidDuration is returned when the duration is zero or negative
	ErrInvalidDuration = errors.New("invalid duration")

	// ErrNoFeaturesOrBundles is returned when no features or bundles are specified
	ErrNoFeaturesOrBundles = errors.New("no features or bundles specified")

	// ErrInvalidOrganizationID is returned when the organization ID is empty
	ErrInvalidOrganizationID = errors.New("invalid organization ID")

	// ErrInvalidPaymentID is returned when the payment ID is empty
	ErrInvalidPaymentID = errors.New("invalid payment ID")

	// ErrLicenseExpired is returned when attempting to use an expired license
	ErrLicenseExpired = errors.New("license expired")

	// ErrLicenseInactive is returned when attempting to use an inactive license
	ErrLicenseInactive = errors.New("license inactive")

	// ErrInsufficientPayment is returned when the payment amount is less than required for the requested features/bundles
	ErrInsufficientPayment = errors.New("insufficient payment amount for requested features")
)
