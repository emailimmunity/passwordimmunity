package payment

import (
	"encoding/json"
	"fmt"
	"os"
)

// FeaturePricing defines pricing for a specific feature
type FeaturePricing struct {
	Monthly     float64 `json:"monthly"`
	Yearly      float64 `json:"yearly"`
	Description string  `json:"description"`
}

// PaymentConfig holds all payment-related configuration
type PaymentConfig struct {
	MollieAPIKey     string                     `json:"mollieApiKey"`
	WebhookBaseURL   string                     `json:"webhookBaseURL"`
	Currency         string                     `json:"currency"`
	FeaturePricing   map[string]FeaturePricing `json:"featurePricing"`
	DefaultLocale    string                     `json:"defaultLocale"`
	PaymentMethods   []string                   `json:"paymentMethods"`
}

var defaultConfig = PaymentConfig{
	Currency:      "EUR",
	DefaultLocale: "en_US",
	PaymentMethods: []string{
		"ideal",
		"creditcard",
		"paypal",
	},
	FeaturePricing: map[string]FeaturePricing{
		"advanced_sso": {
			Monthly:     49.99,
			Yearly:      499.99,
			Description: "Advanced SSO Integration",
		},
		"custom_roles": {
			Monthly:     29.99,
			Yearly:      299.99,
			Description: "Custom Role Management",
		},
		"multi_tenant": {
			Monthly:     99.99,
			Yearly:      999.99,
			Description: "Multi-tenant System",
		},
	},
}

// LoadConfig loads payment configuration from file or environment
func LoadConfig(configPath string) (*PaymentConfig, error) {
	config := defaultConfig

	// Override from environment variables
	if apiKey := os.Getenv("MOLLIE_API_KEY"); apiKey != "" {
		config.MollieAPIKey = apiKey
	}

	if webhookURL := os.Getenv("WEBHOOK_BASE_URL"); webhookURL != "" {
		config.WebhookBaseURL = webhookURL
	}

	if currency := os.Getenv("PAYMENT_CURRENCY"); currency != "" {
		config.Currency = currency
	}

	// Load from config file if provided
	if configPath != "" {
		if err := loadConfigFile(&config, configPath); err != nil {
			return nil, err
		}
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadConfigFile(config *PaymentConfig, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func validateConfig(config *PaymentConfig) error {
	if config.MollieAPIKey == "" {
		return fmt.Errorf("Mollie API key is required")
	}

	if config.WebhookBaseURL == "" {
		return fmt.Errorf("webhook base URL is required")
	}

	return nil
}
