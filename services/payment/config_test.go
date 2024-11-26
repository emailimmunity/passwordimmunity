package payment

import (
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestLoadMollieConfig(t *testing.T) {
    tests := []struct {
        name          string
        envVars      map[string]string
        expectedErr  string
        expectedConf *MollieConfig
    }{
        {
            name: "valid live mode config",
            envVars: map[string]string{
                "MOLLIE_API_KEY":     "live_key",
                "MOLLIE_TEST_API_KEY": "test_key",
                "MOLLIE_TEST_MODE":    "false",
            },
            expectedConf: &MollieConfig{
                APIKey:     "live_key",
                TestAPIKey: "test_key",
                IsTestMode: false,
            },
        },
        {
            name: "valid test mode config",
            envVars: map[string]string{
                "MOLLIE_API_KEY":     "live_key",
                "MOLLIE_TEST_API_KEY": "test_key",
                "MOLLIE_TEST_MODE":    "true",
            },
            expectedConf: &MollieConfig{
                APIKey:     "live_key",
                TestAPIKey: "test_key",
                IsTestMode: true,
            },
        },
        {
            name: "missing live key in live mode",
            envVars: map[string]string{
                "MOLLIE_TEST_API_KEY": "test_key",
                "MOLLIE_TEST_MODE":    "false",
            },
            expectedErr: "live mode enabled but MOLLIE_API_KEY not set",
        },
        {
            name: "missing test key in test mode",
            envVars: map[string]string{
                "MOLLIE_API_KEY":     "live_key",
                "MOLLIE_TEST_MODE":    "true",
            },
            expectedErr: "test mode enabled but MOLLIE_TEST_API_KEY not set",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Clear environment
            os.Clearenv()

            // Set test environment variables
            for k, v := range tt.envVars {
                os.Setenv(k, v)
            }

            // Load configuration
            config, err := LoadMollieConfig()

            if tt.expectedErr != "" {
                assert.EqualError(t, err, tt.expectedErr)
                assert.Nil(t, config)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedConf, config)
                assert.Equal(t, tt.expectedConf.GetActiveKey(), config.GetActiveKey())
            }
        })
    }
}
