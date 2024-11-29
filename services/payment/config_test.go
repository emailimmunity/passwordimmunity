package payment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configData := `{
		"mollieApiKey": "test_key",
		"webhookBaseURL": "https://test.com",
		"currency": "USD",
		"featurePricing": {
			"test_feature": {
				"monthly": 9.99,
				"yearly": 99.99,
				"description": "Test Feature"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		envVars   map[string]string
		wantErr   bool
		checkFunc func(*PaymentConfig) error
	}{
		{
			name: "valid config file",
			envVars: map[string]string{
				"MOLLIE_API_KEY":   "env_test_key",
				"WEBHOOK_BASE_URL": "https://env-test.com",
			},
			wantErr: false,
			checkFunc: func(c *PaymentConfig) error {
				if c.MollieAPIKey != "env_test_key" {
					t.Errorf("expected API key from env, got %s", c.MollieAPIKey)
				}
				return nil
			},
		},
		{
			name:    "missing required fields",
			envVars: map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				if err := tt.checkFunc(config); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  PaymentConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: PaymentConfig{
				MollieAPIKey:   "test_key",
				WebhookBaseURL: "https://test.com",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: PaymentConfig{
				WebhookBaseURL: "https://test.com",
			},
			wantErr: true,
		},
		{
			name: "missing webhook URL",
			config: PaymentConfig{
				MollieAPIKey: "test_key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(&tt.config); (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
