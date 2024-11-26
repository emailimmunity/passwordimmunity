# Payment System Webhooks

## Overview
The payment system uses webhooks to receive real-time payment status updates from Mollie. These updates trigger license management actions and notifications.

## Webhook Endpoints

### Payment Status Webhook
- URL: `/api/v1/payments/webhook`
- Method: `POST`
- Authentication: Mollie signature verification

## Payment Status Handling

### Status Types
- `paid`: Payment successful, license activated
- `failed`: Payment failed, notification sent
- `canceled`: Payment canceled, license deactivated
- `expired`: Payment expired, license deactivated

### Example Webhook Payload
```json
{
  "id": "tr_WDqYK6vllg",
  "status": "paid"
}
```

## Implementation

### Webhook Processing
```go
func HandlePaymentWebhook(ctx context.Context, providerID string, status string) error {
    // Update payment status
    // Handle license activation/deactivation
    // Send notifications if needed
}
```

## Security

### Webhook Verification
- Verify Mollie signature
- Validate webhook payload
- Check payment existence
- Verify organization details

### Error Handling
- Log all webhook attempts
- Track failed processing
- Send notifications for errors
- Maintain audit trail

## Testing

### Test Webhooks
1. Use test mode configuration
2. Generate test payments
3. Verify webhook processing
4. Check notification delivery

## Troubleshooting

### Common Issues
1. Invalid signature
2. Missing payment record
3. Invalid status transition
4. Notification failure

### Resolution Steps
1. Check configuration
2. Verify webhook URL
3. Review error logs
4. Test notification system
