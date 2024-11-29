# Payment Integration Guide

## Overview
PasswordImmunity uses Mollie for secure payment processing and enterprise feature activation. This guide explains how to set up and use the payment integration system.

## Configuration

### Environment Variables
```env
MOLLIE_API_KEY=your_mollie_api_key_here
WEBHOOK_BASE_URL=https://your-domain.com/api/enterprise/activation/webhook
PAYMENT_REDIRECT_URL=https://your-domain.com/dashboard/subscription/complete
```

### Mollie Setup
1. Create a Mollie account at https://www.mollie.com
2. Obtain your API key from the Mollie dashboard
3. Configure webhook URL in your Mollie dashboard
4. Set up your payment methods and currencies

## Payment Flow

1. Feature Activation Request
   ```http
   POST /api/enterprise/activation/initiate
   {
     "feature": "advanced_sso",
     "plan": "monthly"
   }
   ```

2. Payment Validation
   - Amount validation per currency:
     - EUR: Minimum €10.00
     - USD: Minimum $10.00
     - GBP: Minimum £10.00
   - Organization validation
   - Feature/bundle validation
   - Duration format validation (e.g., "720h" for 30 days)

3. Payment Processing
   - Secure redirect to Mollie payment page
   - Payment method selection
   - Amount and currency verification
   - Metadata validation for features and duration

4. Webhook Processing
   - Signature verification
   - Payment status validation (must be "paid")
   - Response validation:
     - Valid organization ID
     - Valid feature/bundle selection
     - Valid payment amount
     - Valid duration format
   - Feature activation upon all validations passing

## Feature Activation

### Available Features and Bundles

#### Individual Features
- advanced_sso
- custom_roles
- advanced_reporting
- multi_tenant
- api_automation

#### Feature Bundles
1. Security Bundle
   - Advanced SSO
   - Hardware key support
   - Custom security policies
2. Management Bundle
   - Custom roles
   - Advanced reporting
   - Multi-tenant support

#### Validation Requirements
- Features must exist in system
- Bundles must contain valid features
- Organization must be eligible
- Payment amount must meet minimums

## Error Handling

### Validation Errors
1. Payment Amount
   - Below minimum amount for currency
   - Invalid currency specified
   - Negative amount provided
2. Organization Validation
   - Invalid organization ID
   - Organization not found
   - Missing permissions
3. Feature/Bundle Validation
   - Invalid feature selection
   - Unavailable features
   - Invalid bundle combination
4. Duration Validation
   - Invalid duration format
   - Unsupported duration
   - Duration parsing errors

### Troubleshooting Steps
1. Payment Validation
   - Verify payment amount meets minimum requirements
   - Check currency is supported (EUR, USD, GBP)
   - Confirm organization ID is valid
   - Validate feature/bundle selection
2. Webhook Processing
   - Check webhook signature
   - Verify payment status is "paid"
   - Confirm all required metadata present
   - Review validation error logs
3. Feature Activation
   - Check license service logs
   - Verify feature flag status
   - Confirm activation timestamps

### Security Considerations
- Always use HTTPS for all endpoints
- Implement comprehensive validation:
  - Payment amount and currency
  - Organization and feature verification
  - Duration format validation
  - Webhook signature verification
- Secure storage of:
  - API keys
  - Payment metadata
  - License information
- Payment validation logging:
  - Record all validation failures
  - Monitor unusual patterns
  - Track activation attempts
   - Monitor webhook processing status
   - Confirm all required metadata present
   - Review validation error logs
3. Feature Activation
   - Check license service logs
   - Verify feature flag status
   - Confirm activation timestamps
   - Monitor webhook processing status

## Error Handling

### Common Issues
1. Invalid API Key
2. Webhook Configuration
3. Payment Verification
4. Feature Activation

### Troubleshooting
- Check Mollie dashboard for payment status
- Verify webhook logs
- Confirm environment variables
- Check system logs for activation status

## Development Integration

### Testing
1. Use Mollie test API key
2. Test webhook handling
3. Verify feature activation
4. Check payment status updates

### Security Considerations
- Always use HTTPS
- Validate webhook signatures
- Secure API key storage
- Monitor payment activities

## Support
For integration support:
1. Check Mollie documentation
2. Review system logs
3. Contact support team
