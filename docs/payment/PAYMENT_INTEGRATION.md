# Payment Integration Guide

## Overview
This guide covers the integration of the Mollie payment system into PasswordImmunity.

## Configuration
1. Set up Mollie API keys in environment variables:
```env
MOLLIE_API_KEY=your_api_key
MOLLIE_WEBHOOK_SECRET=your_webhook_secret
```

## Implementation
### Payment Flow
1. User selects a subscription plan
2. System creates payment request via Mollie API
3. User completes payment on Mollie's platform
4. Webhook receives payment confirmation
5. System activates corresponding license

### Code Examples
```go
// Create payment
payment, err := paymentService.CreatePayment(ctx, &CreatePaymentRequest{
    Amount: "29.99",
    Currency: "EUR",
    Description: "Enterprise License - Annual",
    RedirectURL: "https://passwordimmunity.com/payment/success",
    WebhookURL: "https://api.passwordimmunity.com/webhooks/payment",
})

// Handle webhook
func (s *PaymentService) HandleWebhook(ctx context.Context, req *WebhookRequest) error {
    payment, err := s.mollieClient.GetPayment(ctx, req.PaymentID)
    if err != nil {
        return err
    }

    if payment.Status == "paid" {
        return s.activateLicense(ctx, payment.OrderID)
    }
    return nil
}
```

## Testing
1. Use Mollie's test API key for development
2. Test webhook handling using Mollie's test payments
3. Verify license activation flow with test payments

## Security Considerations
1. Validate webhook signatures
2. Store API keys securely
3. Use HTTPS for all payment endpoints
4. Implement proper error handling
5. Log all payment-related activities

## Error Handling
1. Payment failures
2. Network issues
3. Invalid configurations
4. Webhook processing errors

## Monitoring
1. Track payment success rates
2. Monitor webhook processing
3. Alert on critical failures
4. Track license activation status
