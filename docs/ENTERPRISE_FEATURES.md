# Enterprise Features and Licensing

PasswordImmunity follows a dual-licensing model with enterprise features available through paid licenses.

## Licensing Model

### Community Edition (AGPL-3.0)
- Basic password management
- Standard two-factor authentication
- Password generator
- Basic audit logs
- Standard user management

### Enterprise Edition (Commercial License)
All features from Community Edition plus:

#### Business Tier ($29.99-39.99/month)
- Advanced Audit Logs
  - Detailed activity tracking
  - Custom retention policies
  - Advanced reporting
- Emergency Access
  - Secure account recovery
  - Designated emergency contacts
  - Time-based access controls

#### Enterprise Tier ($49.99-79.99/month)
- Advanced SSO Integration
  - SAML 2.0 support
  - OIDC integration
  - Custom identity provider support
- Multi-Tenant Management
  - Hierarchical organization structure
  - Cross-organization management
  - Custom role definitions
- Advanced Policy Controls
  - Custom password policies
  - Access control rules
  - Security enforcement
- Directory Synchronization
  - Active Directory integration
  - LDAP synchronization
  - Automated user provisioning

## Feature Activation

1. Purchase enterprise features through the admin panel
2. Complete payment via Mollie integration
3. Features are automatically activated upon payment confirmation
4. Grace period provided for testing and evaluation

## Payment Integration

PasswordImmunity uses Mollie for secure payment processing:
- Supports multiple payment methods
- Automatic license activation
- Secure webhook integration
- Payment status tracking

## License Management

Enterprise features are managed through:
- Cryptographic license verification
- Automatic renewal handling
- Grace period management
- Feature-specific activation

## Configuration

Enterprise features can be configured through:
- Environment variables
- Configuration files
- Admin panel settings

See `.env.example` for available configuration options.
