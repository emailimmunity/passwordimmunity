package payment

import (
    "fmt"
    "os"
)

type MollieConfig struct {
    APIKey     string
    TestAPIKey string
    IsTestMode bool
}

func LoadMollieConfig() (*MollieConfig, error) {
    config := &MollieConfig{
        APIKey:     os.Getenv("MOLLIE_API_KEY"),
        TestAPIKey: os.Getenv("MOLLIE_TEST_API_KEY"),
        IsTestMode: os.Getenv("MOLLIE_TEST_MODE") == "true",
    }

    if err := config.Validate(); err != nil {
        return nil, err
    }

    return config, nil
}

func (c *MollieConfig) Validate() error {
    if c.IsTestMode && c.TestAPIKey == "" {
        return fmt.Errorf("test mode enabled but MOLLIE_TEST_API_KEY not set")
    }

    if !c.IsTestMode && c.APIKey == "" {
        return fmt.Errorf("live mode enabled but MOLLIE_API_KEY not set")
    }

    return nil
}

func (c *MollieConfig) GetActiveKey() string {
    if c.IsTestMode {
        return c.TestAPIKey
    }
    return c.APIKey
}
