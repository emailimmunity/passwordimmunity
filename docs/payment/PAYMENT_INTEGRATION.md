# Payment System Integration

## Overview
The payment system integrates with Mollie for payment processing and includes a notification system for payment failures and license cancellations.

## Components

### Payment Service
- Handles payment creation and processing
- Manages payment status updates via webhooks
- Integrates with license management
- Sends notifications for payment failures

### Notification Service
- Sends email notifications for payment failures
- Sends email notifications for license cancellations
- Uses configurable email service implementation

## Configuration

### Environment Variables
```
MOLLIE_API_KEY=your_live_api_key
MOLLIE_TEST_API_KEY=your_test_api_key
MOLLIE_TEST_MODE=true/false
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASS=your_smtp_password
SMTP_FROM=notifications@passwordimmunity.com
```

## Payment Workflow

1. Payment Creation
   - Client initiates payment
   - System creates payment record
   - Redirects to Mollie payment page

2. Payment Processing
   - Mollie processes payment
   - Sends webhook with payment status
   - System updates payment status

3. License Management
   - Successful payment activates license
   - Failed payment triggers notifications
   - Failed payment cancels associated license

4. Notification System
   - Sends email notifications for payment failures
   - Sends email notifications for license cancellations
   - Logs all notification attempts

## Error Handling
- Payment failures trigger notifications
- License cancellations trigger notifications
- All errors are logged with context
- Failed notifications are logged but don't block payment processing

## Testing
- Test mode available via MOLLIE_TEST_MODE
- Comprehensive test suite for payment workflow
- Mock implementations for testing
- Integration tests for full workflow

## Security Considerations
- API keys stored in environment variables
- SMTP credentials stored in environment variables
- All sensitive data logged securely
- Payment data handled according to PCI compliance

## Monitoring
- Payment status changes logged
- Notification attempts logged
- Error conditions logged
- Integration status monitored

## Troubleshooting
1. Payment Creation Issues
   - Check API key configuration
   - Verify Mollie API status
   - Check payment creation logs

2. Webhook Issues
   - Verify webhook URL configuration
   - Check webhook processing logs
   - Verify payment status updates

3. Notification Issues
   - Check SMTP configuration
   - Verify email service status
   - Check notification logs

## Development Guidelines
1. Payment Integration
   ```go
   // Create a payment
   payment, err := paymentService.CreatePayment(ctx, orgID, "enterprise", "yearly")
   if err != nil {
       // Handle error
   }
   ```

2. Webhook Handling
   ```go
   // Handle webhook
   err := paymentService.HandlePaymentWebhook(ctx, providerID, "paid")
   if err != nil {
       // Handle error
   }
   ```

3. Notification Implementation
   ```go
   // Send notification
   err := notificationService.NotifyPaymentFailed(ctx, payment, "Payment declined")
   if err != nil {
       // Handle error
   }
   ```
