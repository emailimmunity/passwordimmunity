# Payment System Configuration

## Environment Variables
```env
# Required
MOLLIE_API_KEY=live_xxx or test_xxx
MOLLIE_WEBHOOK_SECRET=your_webhook_secret

# Optional
MOLLIE_API_ENDPOINT=https://api.mollie.com/v2
PAYMENT_WEBHOOK_PATH=/webhooks/payment
PAYMENT_SUCCESS_URL=https://passwordimmunity.com/payment/success
PAYMENT_FAILURE_URL=https://passwordimmunity.com/payment/failure
```

## License Tiers
```yaml
tiers:
  enterprise:
    monthly: 29.99
    annual: 299.99
    features:
      - Advanced SSO
      - Custom Roles
      - Priority Support

  premium:
    monthly: 9.99
    annual: 99.99
    features:
      - Advanced Security
      - API Access
      - Extended History
```

## Webhook Configuration
1. Configure webhook URL in Mollie dashboard
2. Set webhook secret in environment
3. Enable webhook validation
4. Configure retry policy

## Security Settings
1. TLS requirements
2. API key rotation policy
3. Webhook signature validation
4. Rate limiting configuration

## Error Handling
1. Payment timeout settings
2. Retry configurations
3. Notification thresholds
4. Alert configurations

## Monitoring Configuration
1. Metrics collection
2. Alert thresholds
3. Dashboard configuration
4. Log levels and retention
