# Webhook Integration Guide

## Webhook Setup

### Endpoint Configuration
```
POST /api/enterprise/activation/webhook
```

The webhook endpoint processes payment status updates from Mollie and manages feature activation.

### Security Requirements
1. HTTPS only
2. IP whitelisting (Mollie IPs)
3. Signature verification
4. Request validation

## Webhook Events

### Supported Events
- `payment.status_changed`
- `subscription.created`
- `subscription.canceled`
- `subscription.suspended`

### Event Processing
1. Verify webhook signature
2. Validate payment status
3. Update subscription status
4. Activate/deactivate features
5. Send notifications

## Error Handling

### Retry Mechanism
- Exponential backoff
- Maximum 5 retries
- Error logging
- Alert system

### Common Issues
1. Invalid signatures
2. Timeout errors
3. Database conflicts
4. Network issues

## Testing

### Test Webhook Flow
1. Use Mollie test environment
2. Generate test events
3. Verify processing
4. Check feature states

## Monitoring

### Key Metrics
1. Webhook success rate
2. Processing time
3. Error frequency
4. Feature activation time

## Best Practices

### Implementation Guidelines
1. Idempotent processing
2. Async event handling
3. Comprehensive logging
4. Status tracking

### Security Measures
1. TLS 1.2+
2. Request validation
3. Rate limiting
4. Access control
