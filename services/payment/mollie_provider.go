package payment

import (
    "context"
    "fmt"
    "strconv"
)

type MollieProvider struct {
    client *MollieClient
}

func NewMollieProvider(client *MollieClient) *MollieProvider {
    return &MollieProvider{
        client: client,
    }
}

func (p *MollieProvider) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
    // Convert PaymentRequest to Mollie-specific request
    mollieReq := &CreatePaymentRequest{
        Amount:      parseAmount(req.Amount.Value),
        Currency:    req.Amount.Currency,
        Description: req.Description,
        FeatureID:   getFirstFeature(req.Metadata.Features),
        BundleID:    getFirstBundle(req.Metadata.Bundles),
        OrgID:       req.Metadata.OrganizationID,
    }

    // Create payment via Mollie
    payment, err := p.client.CreatePayment(ctx, mollieReq)
    if err != nil {
        return nil, fmt.Errorf("mollie payment creation failed: %w", err)
    }

    // Convert Mollie payment to PaymentResponse
    return &PaymentResponse{
        ID:          payment.ID,
        Status:      payment.Status,
        Amount:      req.Amount,
        RedirectURL: payment.RedirectURL,
        WebhookURL:  payment.WebhookURL,
        Metadata:    req.Metadata,
    }, nil
}

func (p *MollieProvider) VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
    // Get payment details from Mollie
    payment, err := p.client.GetPayment(ctx, paymentID)
    if err != nil {
        return nil, fmt.Errorf("failed to verify payment: %w", err)
    }

    // Extract metadata
    metadata := payment.Metadata.(map[string]interface{})

    // Convert Mollie payment to PaymentResponse
    return &PaymentResponse{
        ID:     payment.ID,
        Status: payment.Status,
        Amount: Amount{
            Currency: payment.Amount.Currency,
            Value:    payment.Amount.Value,
        },
        RedirectURL: payment.RedirectURL,
        WebhookURL:  payment.WebhookURL,
        Metadata: PaymentMetadata{
            OrganizationID: metadata["org_id"].(string),
            Features:       extractFeatures(metadata),
            Bundles:       extractBundles(metadata),
            Duration:      metadata["duration"].(string),
        },
    }, nil
}

// Helper functions

func parseAmount(value string) float64 {
    amount, _ := strconv.ParseFloat(value, 64)
    return amount
}

func getFirstFeature(features []string) string {
    if len(features) > 0 {
        return features[0]
    }
    return ""
}

func getFirstBundle(bundles []string) string {
    if len(bundles) > 0 {
        return bundles[0]
    }
    return ""
}

func extractFeatures(metadata map[string]interface{}) []string {
    if featureID, ok := metadata["feature_id"].(string); ok && featureID != "" {
        return []string{featureID}
    }
    return nil
}

func extractBundles(metadata map[string]interface{}) []string {
    if bundleID, ok := metadata["bundle_id"].(string); ok && bundleID != "" {
        return []string{bundleID}
    }
    return nil
}
