package payment

import "errors"

var (
	// ErrInvalidPaymentID is returned when payment ID is empty or invalid
	ErrInvalidPaymentID = errors.New("invalid payment ID")

	// ErrNoFeaturesSelected is returned when no features or bundles are selected
	ErrNoFeaturesSelected = errors.New("no features or bundles selected")

	// ErrInvalidAmount is returned when payment amount is zero or negative
	ErrInvalidAmount = errors.New("invalid payment amount")

	// ErrInvalidDateRange is returned when activation date range is invalid
	ErrInvalidDateRange = errors.New("invalid date range")

	// ErrPaymentNotFound is returned when payment record is not found
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrPaymentNotCompleted is returned when payment status is not completed
	ErrPaymentNotCompleted = errors.New("payment not completed")

	// ErrFeatureAlreadyActive is returned when attempting to activate an already active feature
	ErrFeatureAlreadyActive = errors.New("feature already active")

	// ErrUnsupportedCurrency is returned when currency is not supported
	ErrUnsupportedCurrency = errors.New("unsupported currency")

	// ErrInvalidBillingPeriod is returned when billing period is not supported
	ErrInvalidBillingPeriod = errors.New("invalid billing period")

	// ErrNoFeaturesOrBundles is returned when neither features nor bundles are specified
	ErrNoFeaturesOrBundles = errors.New("no features or bundles specified")

	// ErrInvalidMetadata is returned when payment metadata is invalid or incomplete
	ErrInvalidMetadata = errors.New("invalid payment metadata")
)
