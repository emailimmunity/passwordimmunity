package config

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestConvertCurrency(t *testing.T) {
	tests := []struct {
		name         string
		amount       decimal.Decimal
		fromCurrency string
		toCurrency   string
		want         decimal.Decimal
		wantErr      bool
	}{
		{
			name:         "Same currency",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "USD",
			toCurrency:   "USD",
			want:         decimal.NewFromFloat(100),
			wantErr:      false,
		},
		{
			name:         "Different currency",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "USD",
			toCurrency:   "EUR",
			want:         decimal.NewFromFloat(100), // Currently 1:1 as per TODO
			wantErr:      false,
		},
		{
			name:         "Unsupported currency",
			amount:       decimal.NewFromFloat(100),
			fromCurrency: "USD",
			toCurrency:   "XXX",
			want:         decimal.Zero,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertCurrency(tt.amount, tt.fromCurrency, tt.toCurrency)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}
}
