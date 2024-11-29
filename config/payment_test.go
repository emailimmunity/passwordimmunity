package config

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLoadPaymentConfig(t *testing.T) {
    t.Run("successful config loading", func(t *testing.T) {
        // Set required environment variables
        os.Setenv("MOLLIE_API_KEY", "test_key")
        os.Setenv("WEBHOOK_BASE_URL", "https://api.example.com")
        os.Setenv("REDIRECT_BASE_URL", "https://app.example.com")
        defer func() {
            os.Unsetenv("MOLLIE_API_KEY")
            os.Unsetenv("WEBHOOK_BASE_URL")
            os.Unsetenv("REDIRECT_BASE_URL")
        }()

        config, err := LoadPaymentConfig()
        assert.NoError(t, err)
        assert.NotNil(t, config)
        assert.Equal(t, "test_key", config.MollieAPIKey)
        assert.Equal(t, "https://api.example.com", config.WebhookBaseURL)
        assert.Equal(t, "https://app.example.com", config.RedirectBaseURL)
    })

    t.Run("missing required environment variables", func(t *testing.T) {
        os.Unsetenv("MOLLIE_API_KEY")
        os.Unsetenv("WEBHOOK_BASE_URL")
        os.Unsetenv("REDIRECT_BASE_URL")

        config, err := LoadPaymentConfig()
        assert.Error(t, err)
        assert.Nil(t, config)
        assert.Contains(t, err.Error(), "MOLLIE_API_KEY")
    })
}

func TestPaymentConfig_GetFeaturePrice(t *testing.T) {
    config := &PaymentConfig{
        SupportedFeatures: map[string]float64{
            "advanced_sso": 10.00,
        },
    }

    t.Run("valid feature", func(t *testing.T) {
        price, err := config.GetFeaturePrice("advanced_sso")
        assert.NoError(t, err)
        assert.Equal(t, 10.00, price)
    })

    t.Run("invalid feature", func(t *testing.T) {
        price, err := config.GetFeaturePrice("nonexistent")
        assert.Error(t, err)
        assert.Equal(t, 0.0, price)
    })
}

func TestPaymentConfig_GetBundlePrice(t *testing.T) {
    config := &PaymentConfig{
        BundlePricing: map[string]float64{
            "enterprise": 50.00,
        },
    }

    t.Run("valid bundle", func(t *testing.T) {
        price, err := config.GetBundlePrice("enterprise")
        assert.NoError(t, err)
        assert.Equal(t, 50.00, price)
    })

    t.Run("invalid bundle", func(t *testing.T) {
        price, err := config.GetBundlePrice("nonexistent")
        assert.Error(t, err)
        assert.Equal(t, 0.0, price)
    })
}

func TestPaymentConfig_GetWebhookURL(t *testing.T) {
    config := &PaymentConfig{
        WebhookBaseURL: "https://api.example.com",
    }

    url := config.GetWebhookURL("payment123")
    assert.Equal(t, "https://api.example.com/api/payments/webhook/payment123", url)
}

func TestPaymentConfig_GetRedirectURL(t *testing.T) {
    config := &PaymentConfig{
        RedirectBaseURL: "https://app.example.com",
    }

    url := config.GetRedirectURL("payment123")
    assert.Equal(t, "https://app.example.com/payments/complete/payment123", url)
}

func TestPaymentConfig_IsSupportedCurrency(t *testing.T) {
    config := &PaymentConfig{
        MinimumAmounts: map[string]float64{
            "EUR": 1.00,
            "USD": 1.00,
        },
    }

    assert.True(t, config.IsSupportedCurrency("EUR"))
    assert.True(t, config.IsSupportedCurrency("USD"))
    assert.False(t, config.IsSupportedCurrency("GBP"))
}

func TestPaymentConfig_GetMinimumAmount(t *testing.T) {
    config := &PaymentConfig{
        MinimumAmounts: map[string]float64{
            "EUR": 1.00,
            "USD": 1.00,
        },
    }

    t.Run("supported currency", func(t *testing.T) {
        amount, err := config.GetMinimumAmount("EUR")
        assert.NoError(t, err)
        assert.Equal(t, 1.00, amount)
    })

    t.Run("unsupported currency", func(t *testing.T) {
        amount, err := config.GetMinimumAmount("GBP")
        assert.Error(t, err)
        assert.Equal(t, 0.0, amount)
    })
}
