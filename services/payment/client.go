package payment

import (
    "fmt"
    "github.com/mollie/mollie-api-go/v2/mollie"
)

type MollieClient interface {
    CreatePayment(amount, currency, description string) (*mollie.Payment, error)
    GetPayment(id string) (*mollie.Payment, error)
}

type mollieClient struct {
    client *mollie.Client
    config *MollieConfig
}

func NewMollieClient(config *MollieConfig) (MollieClient, error) {
    if config == nil {
        return nil, fmt.Errorf("mollie configuration is required")
    }

    client := mollie.NewClient(config.GetActiveKey())

    return &mollieClient{
        client: client,
        config: config,
    }, nil
}

func (m *mollieClient) CreatePayment(amount, currency, description string) (*mollie.Payment, error) {
    payment, err := m.client.Payments.Create(&mollie.Payment{
        Amount: &mollie.Amount{
            Currency: currency,
            Value:    amount,
        },
        Description: description,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create Mollie payment: %w", err)
    }
    return payment, nil
}

func (m *mollieClient) GetPayment(id string) (*mollie.Payment, error) {
    payment, err := m.client.Payments.Get(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get Mollie payment: %w", err)
    }
    return payment, nil
}
