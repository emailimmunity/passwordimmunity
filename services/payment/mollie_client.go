package payment

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/mollie/mollie-api-go/v2/mollie"
)

type MollieClient struct {
    client *mollie.Client
    config *Config
}

type Config struct {
    APIKey      string
    WebhookURL  string
    RedirectURL string
}

func NewMollieClient(config *Config) (*MollieClient, error) {
    client, err := mollie.NewClient(config.APIKey, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create Mollie client: %w", err)
    }

    return &MollieClient{
        client: client,
        config: config,
    }, nil
}

type CreatePaymentRequest struct {
    Amount      float64
    Currency    string
    Description string
    FeatureID   string
    BundleID    string
    OrgID       string
}

func (c *MollieClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*mollie.Payment, error) {
    payment, err := c.client.Payments.Create(&mollie.Payment{
        Amount: &mollie.Amount{
            Currency: req.Currency,
            Value:    fmt.Sprintf("%.2f", req.Amount),
        },
        Description: req.Description,
        RedirectURL: c.config.RedirectURL,
        WebhookURL:  c.config.WebhookURL,
        Metadata: map[string]interface{}{
            "feature_id": req.FeatureID,
            "bundle_id": req.BundleID,
            "org_id":    req.OrgID,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    return payment, nil
}

func (c *MollieClient) GetPayment(ctx context.Context, paymentID string) (*mollie.Payment, error) {
    payment, err := c.client.Payments.Get(paymentID)
    if err != nil {
        return nil, fmt.Errorf("failed to get payment: %w", err)
    }

    return payment, nil
}

func (c *MollieClient) HandleWebhook(ctx context.Context, paymentID string) error {
    payment, err := c.GetPayment(ctx, paymentID)
    if err != nil {
        return err
    }

    if payment.Status != "paid" {
        return fmt.Errorf("payment status is %s", payment.Status)
    }

    // Extract metadata
    metadata := payment.Metadata.(map[string]interface{})
    featureID, _ := metadata["feature_id"].(string)
    bundleID, _ := metadata["bundle_id"].(string)
    orgID, _ := metadata["org_id"].(string)

    // TODO: Trigger feature activation based on the payment
    // This will be handled by the payment service

    return nil
}

func (c *MollieClient) ValidateWebhook(r *http.Request) (string, error) {
    if err := r.ParseForm(); err != nil {
        return "", fmt.Errorf("failed to parse form: %w", err)
    }

    paymentID := r.PostForm.Get("id")
    if paymentID == "" {
        return "", fmt.Errorf("missing payment ID in webhook")
    }

    return paymentID, nil
}
