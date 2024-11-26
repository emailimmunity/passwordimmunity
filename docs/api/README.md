# API Documentation

## Authentication

All API requests require authentication using a bearer token:
```http
Authorization: Bearer your_api_token
```

## API Endpoints

### Authentication

```http
POST /api/auth/login
POST /api/auth/register
POST /api/auth/2fa
GET /api/auth/profile
```

### Password Management

```http
GET /api/vault/items
POST /api/vault/items
PUT /api/vault/items/{id}
DELETE /api/vault/items/{id}
```

### Organization Management

```http
GET /api/organizations
POST /api/organizations
PUT /api/organizations/{id}
DELETE /api/organizations/{id}
```

### Role Management

```http
GET /api/roles
POST /api/roles
PUT /api/roles/{id}
DELETE /api/roles/{id}
```

### Enterprise Features

```http
POST /api/sso/configure
GET /api/audit-logs
POST /api/policies
GET /api/reports/security
```

## Response Format

All responses follow the format:
```json
{
  "success": true,
  "data": {},
  "message": "Operation successful"
}
```

## Error Handling

Errors follow the format:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

## Rate Limiting

API requests are limited to:
- 100 requests per minute for standard users
- 1000 requests per minute for enterprise users

## Webhooks

Enterprise features include webhook support for events:
```http
POST /api/webhooks/configure
```

## Examples

### Creating a New Vault Item
```bash
curl -X POST https://your-domain.com/api/vault/items \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "login",
    "name": "Example Login",
    "login": {
      "username": "user@example.com",
      "password": "encrypted_password"
    }
  }'
```

### Managing Roles
```bash
curl -X POST https://your-domain.com/api/roles \
  -H "Authorization: Bearer your_api_token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Team Lead",
    "permissions": ["manage_users", "view_reports"]
  }'
```

## SDK Support

Official SDKs are available for:
- Python
- JavaScript
- Go
- Ruby
- PHP

## Best Practices

1. Always use HTTPS
2. Implement proper error handling
3. Cache responses when appropriate
4. Use pagination for large datasets
5. Implement retry logic with exponential backoff

## Migration Guide

For users migrating from Bitwarden/Vaultwarden API:
1. Update authentication headers
2. Review endpoint changes
3. Update response handling
4. Test enterprise features
