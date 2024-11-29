package config

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestParseBillingPeriod(t *testing.T) {
	tests := []struct {
		name    string
		period  string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "monthly period",
			period:  "monthly",
			want:    30 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "quarterly period",
			period:  "quarterly",
			want:    90 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "yearly period",
			period:  "yearly",
			want:    365 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid period",
			period:  "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBillingPeriod(tt.period)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetBillingPeriods(t *testing.T) {
	periods := GetBillingPeriods()
	assert.Equal(t, []string{"monthly", "quarterly", "yearly"}, periods)
}

func TestValidateBillingPeriod(t *testing.T) {
	assert.True(t, ValidateBillingPeriod("monthly"))
	assert.True(t, ValidateBillingPeriod("quarterly"))
	assert.True(t, ValidateBillingPeriod("yearly"))
	assert.False(t, ValidateBillingPeriod("invalid"))
}
