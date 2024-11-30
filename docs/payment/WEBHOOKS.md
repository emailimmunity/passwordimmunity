# Payment System Webhooks

## Overview
The webhook system handles asynchronous payment status updates from Mollie.

## Endpoints
```
POST /webhooks/payment
```

## Authentication
- Validate webhook signature using MOLLIE_WEBHOOK_SECRET
- Verify payment ID format
- Validate request origin

## Request Format
```json
{
  "id": "tr_xxx",
  "status": "paid",
  "amount": {
    "value": "29.99",
    "currency": "EUR"
  },
  "metadata": {
    "order_id": "order_xxx",
    "license_type": "enterprise"
  }
}
```

## Response Codes
- 200: Successfully processed
- 400: Invalid request format
- 401: Invalid signature
- 500: Processing error

## Processing Flow
1. Validate request signature
2. Verify payment status
3. Update license status
4. Send confirmation email
5. Log transaction

## Error Handling
1. Invalid signatures
2. Duplicate webhooks
3. Network timeouts
4. Database errors
5. Email failures

## Monitoring
1. Track webhook success rate
2. Monitor processing time
3. Alert on failures
4. Log all webhook events

## Testing
1. Use test webhooks
2. Simulate failures
3. Verify retry logic
4. Test all status types
