package config

import (
    "fmt"
    "os"
)

type PaymentConfig struct {
    MollieAPIKey      string
    WebhookBaseURL    string
    RedirectBaseURL   string
    Environment       string
    MinimumAmounts    map[string]float64
    SupportedFeatures map[string]float64
    FeatureBundles    map[string][]string
    BundlePricing     map[string]float64
}

func LoadPaymentConfig() (*PaymentConfig, error) {
    apiKey := os.Getenv("MOLLIE_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("MOLLIE_API_KEY environment variable is required")
    }

    webhookBase := os.Getenv("WEBHOOK_BASE_URL")
    if webhookBase == "" {
        return nil, fmt.Errorf("WEBHOOK_BASE_URL environment variable is required")
    }

    redirectBase := os.Getenv("REDIRECT_BASE_URL")
    if redirectBase == "" {
        return nil, fmt.Errorf("REDIRECT_BASE_URL environment variable is required")
    }

    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    return &PaymentConfig{
        MollieAPIKey:    apiKey,
        WebhookBaseURL:  webhookBase,
        RedirectBaseURL: redirectBase,
        Environment:     env,
        MinimumAmounts: map[string]float64{
            "EUR": 1.00,
            "USD": 1.00,
            "GBP": 1.00,
        },
        SupportedFeatures: map[string]float64{
            "advanced_sso":       10.00,
            "directory_sync":     15.00,
            "custom_roles":       20.00,
            "advanced_reporting": 25.00,
        },
        FeatureBundles: map[string][]string{
            "enterprise": {
                "advanced_sso",
                "directory_sync",
                "custom_roles",
                "advanced_reporting",
            },
            "security": {
                "advanced_sso",
                "directory_sync",
            },
        },
        BundlePricing: map[string]float64{
            "enterprise": 50.00,
            "security":   30.00,
        },
    }, nil
}

func (c *PaymentConfig) GetFeaturePrice(featureID string) (float64, error) {
    price, ok := c.SupportedFeatures[featureID]
    if !ok {
        return 0, fmt.Errorf("unsupported feature: %s", featureID)
    }
    return price, nil
}

func (c *PaymentConfig) GetBundlePrice(bundleID string) (float64, error) {
    price, ok := c.BundlePricing[bundleID]
    if !ok {
        return 0, fmt.Errorf("unsupported bundle: %s", bundleID)
    }
    return price, nil
}

func (c *PaymentConfig) GetWebhookURL(paymentID string) string {
    return fmt.Sprintf("%s/api/payments/webhook/%s", c.WebhookBaseURL, paymentID)
}

func (c *PaymentConfig) GetRedirectURL(paymentID string) string {
    return fmt.Sprintf("%s/payments/complete/%s", c.RedirectBaseURL, paymentID)
}

func (c *PaymentConfig) IsSupportedCurrency(currency string) bool {
    _, ok := c.MinimumAmounts[currency]
    return ok
}

func (c *PaymentConfig) GetMinimumAmount(currency string) (float64, error) {
    amount, ok := c.MinimumAmounts[currency]
    if !ok {
        return 0, fmt.Errorf("unsupported currency: %s", currency)
    }
    return amount, nil
}
