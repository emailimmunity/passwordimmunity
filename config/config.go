package config

import (
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type Config struct {
	// Server settings
	Port string
	Environment string

	// Database settings
	DatabaseURL string

	// Payment settings
	MollieAPIKey       string
	WebhookBaseURL     string
	PaymentRedirectURL string
	PaymentCurrency    string
	PaymentLocale      string
	PaymentRetryLimit  int
	PaymentExpiryHours int

	// Enterprise feature settings
	EnterpriseFeatures map[string]FeatureConfig
	LicensePublicKey   string
	LicensePrivateKey  string
	LicenseExpiryDays  int
	LicenseGraceDays   int

	// Feature flag settings
	DefaultFeatures    []string
	FeatureGracePeriod int // days
}

type FeatureConfig struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Tier        string  `json:"tier"`
	GracePeriod int     `json:"grace_period"` // days
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv("DATABASE_URL", "postgresql://localhost:5432/passwordimmunity?sslmode=disable"),

		MollieAPIKey:       requireEnv("MOLLIE_API_KEY"),
		WebhookBaseURL:     requireEnv("WEBHOOK_BASE_URL"),
		PaymentRedirectURL: requireEnv("PAYMENT_REDIRECT_URL"),
		PaymentCurrency:    getEnv("PAYMENT_CURRENCY", "EUR"),
		PaymentLocale:      getEnv("PAYMENT_LOCALE", "en_US"),
		PaymentRetryLimit:  getEnvInt("PAYMENT_RETRY_LIMIT", 3),
		PaymentExpiryHours: getEnvInt("PAYMENT_EXPIRY_HOURS", 24),

		LicensePublicKey:   requireEnv("LICENSE_PUBLIC_KEY"),
		LicensePrivateKey:  requireEnv("LICENSE_PRIVATE_KEY"),
		LicenseExpiryDays:  getEnvInt("LICENSE_EXPIRY_DAYS", 365),
		LicenseGraceDays:   getEnvInt("LICENSE_GRACE_DAYS", 14),
		FeatureGracePeriod: getEnvInt("FEATURE_GRACE_PERIOD", 7),

		EnterpriseFeatures: loadEnterpriseFeatures(),

		DefaultFeatures: []string{
			"basic_vault",
			"two_factor_auth",
			"password_generator",
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable not set: " + key)
	}
	return value
}

func loadEnterpriseFeatures() map[string]FeatureConfig {
	return map[string]FeatureConfig{
		"advanced_sso": {
			Name:        "Advanced SSO Integration",
			Description: "Enterprise-grade Single Sign-On capabilities with SAML and OIDC support",
			Price:       49.99,
			Tier:        "enterprise",
			GracePeriod: 14,
		},
		"multi_tenant": {
			Name:        "Multi-Tenant Management",
			Description: "Advanced organization and tenant management with hierarchical controls",
			Price:       79.99,
			Tier:        "enterprise",
			GracePeriod: 14,
		},
		"advanced_audit": {
			Name:        "Advanced Audit Logs",
			Description: "Comprehensive audit logging with advanced reporting and retention",
			Price:       29.99,
			Tier:        "business",
			GracePeriod: 7,
		},
		"advanced_policy": {
			Name:        "Advanced Policy Controls",
			Description: "Enterprise policy management and enforcement capabilities",
			Price:       39.99,
			Tier:        "enterprise",
			GracePeriod: 14,
		},
		"directory_sync": {
			Name:        "Directory Synchronization",
			Description: "Advanced directory integration and user provisioning",
			Price:       59.99,
			Tier:        "enterprise",
			GracePeriod: 14,
		},
		"emergency_access": {
			Name:        "Emergency Access",
			Description: "Secure emergency access protocols and management",
			Price:       19.99,
			Tier:        "business",
			GracePeriod: 7,
		},
	}
}
