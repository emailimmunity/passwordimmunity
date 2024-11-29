package config

import (
	"fmt"
	"time"
)

// ParseBillingPeriod converts a billing period string to time.Duration
func ParseBillingPeriod(period string) (time.Duration, error) {
	switch period {
	case "monthly":
		return 30 * 24 * time.Hour, nil
	case "quarterly":
		return 90 * 24 * time.Hour, nil
	case "yearly":
		return 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid billing period: %s", period)
	}
}

// GetBillingPeriods returns all valid billing periods
func GetBillingPeriods() []string {
	return []string{"monthly", "quarterly", "yearly"}
}

// ValidateBillingPeriod checks if a billing period is valid
func ValidateBillingPeriod(period string) bool {
	_, err := ParseBillingPeriod(period)
	return err == nil
}
