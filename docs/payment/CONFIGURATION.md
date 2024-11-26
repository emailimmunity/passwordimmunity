# Payment System Configuration

## Environment Variables

### Required Variables
```
# Mollie Configuration
MOLLIE_API_KEY=your_live_api_key
MOLLIE_TEST_API_KEY=your_test_api_key
MOLLIE_TEST_MODE=true/false

# SMTP Configuration for Notifications
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASS=your_smtp_password
SMTP_FROM=notifications@passwordimmunity.com
```

## License Types and Pricing

### Enterprise License
- Monthly: €99.00
- Yearly: €999.00 (Save ~16%)

### Features Included
- All enterprise features
- Priority support
- Advanced role management
- Custom SSO integrations
- Advanced audit logging
- API access

## Payment Processing

### Test Mode
1. Set `MOLLIE_TEST_MODE=true`
2. Use test API key for development
3. Test cards available in Mollie documentation

### Production Mode
1. Set `MOLLIE_TEST_MODE=false`
2. Use live API key
3. Ensure proper SSL configuration
4. Configure production SMTP settings

## Notification System

### Email Templates
Notifications are sent for:
- Payment failures
- License cancellations
- License activations
- Payment processing issues

### Monitoring
- All notifications are logged
- Failed notifications tracked
- Email delivery status monitored

## Security Considerations
- API keys stored securely in environment
- SMTP credentials protected
- All transactions logged
- PCI compliance maintained
