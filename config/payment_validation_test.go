package config

import (
	"testing"
)

func TestValidatePaymentAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		want     bool
	}{
		{"Valid EUR amount", 15.00, "EUR", true},
		{"Valid USD amount", 20.00, "USD", true},
		{"Valid GBP amount", 12.50, "GBP", true},
		{"Below minimum EUR", 9.99, "EUR", false},
		{"Below minimum USD", 5.00, "USD", false},
		{"Below minimum GBP", 8.00, "GBP", false},
		{"Invalid currency", 50.00, "JPY", false},
		{"Zero amount", 0.00, "EUR", false},
		{"Negative amount", -10.00, "USD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePaymentAmount(tt.amount, tt.currency); got != tt.want {
				t.Errorf("ValidatePaymentAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSupportedCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		want     bool
	}{
		{"Valid EUR", "EUR", true},
		{"Valid USD", "USD", true},
		{"Valid GBP", "GBP", true},
		{"Invalid JPY", "JPY", false},
		{"Empty string", "", false},
		{"Lowercase eur", "eur", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedCurrency(tt.currency); got != tt.want {
				t.Errorf("IsSupportedCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMinimumAmount(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		want     float64
	}{
		{"EUR minimum", "EUR", 10.00},
		{"USD minimum", "USD", 10.00},
		{"GBP minimum", "GBP", 10.00},
		{"Invalid currency", "JPY", 0.00},
		{"Empty currency", "", 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMinimumAmount(tt.currency); got != tt.want {
				t.Errorf("GetMinimumAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}
