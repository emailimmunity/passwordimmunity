# PasswordImmunity

A secure, enterprise-grade password management system with advanced features and Mollie payment integration.

## Features

### Community Edition (Free)
- Secure password vault with basic encryption
- Standard two-factor authentication
- Basic role management
- Standard API access
- Community support

### Enterprise Edition (Licensed)
The following features require an Enterprise license:

#### Security & Authentication
- Advanced SSO integrations (SAML, OIDC)
- Hardware security key support
- Advanced encryption options
- Custom security policies

#### Management & Control
- Enterprise policies and controls
- Advanced role and permission management
- Multi-tenant system capabilities
- Advanced group management
- Custom organizational units

#### Monitoring & Reporting
- Advanced audit logging
- Security compliance reporting
- Usage analytics dashboard
- Custom report generation
- Activity monitoring
- Configurable report retention policies
  - Customizable retention periods for daily, weekly, and monthly reports
  - Organization-specific retention settings
  - Automated cleanup based on retention policies

#### Integration & API
- Enterprise API access
- Custom webhook support
- Automated provisioning
- Third-party integrations
- API rate limit controls

#### Support & Services
- Priority technical support
- Custom feature development
- Deployment assistance
- Security assessments
- Training and documentation

## Getting Started
### Prerequisites
- Go 1.21 or higher
- PostgreSQL 13 or higher
- Redis (optional, for caching)

### Installation
```bash
git clone https://github.com/emailimmunity/passwordimmunity.git
cd passwordimmunity
make install
```

## Enterprise Features and Payment Integration

### Activating Enterprise Features
1. Sign up for a PasswordImmunity account
2. Navigate to the Enterprise Dashboard
3. Choose the features you want to activate
4. Complete the payment process through our secure Mollie integration
5. Features will be automatically activated upon successful payment

### Payment Integration
- Secure payment processing through Mollie
- Support for multiple currencies with minimum amounts:
  - EUR: €10.00
  - USD: $10.00
  - GBP: £10.00
- Strict payment validation:
  - Amount and currency validation
  - Payment status verification
  - Organization and feature validation
  - Duration format validation
- Real-time webhook processing with validation
- Comprehensive payment history and invoicing
- Automatic license activation upon verified payment
- Secure error handling and validation logging

### Enterprise License Management
- Secure license activation through validated payments
- Feature-specific activation controls:
  - Individual feature activation
  - Bundle-based licensing
  - Time-based access control
- Comprehensive validation:
  - Payment verification
  - License duration validation
  - Feature availability checks
  - Organization validation
- Real-time activation status monitoring
- Automated notification system for:
  - License activation
  - Expiration warnings
  - Payment failures
  - Renewal reminders

## License
This software is available under a dual license model:

1. Community Edition: GNU Affero General Public License v3.0 (AGPL-3.0)
   - Includes all core password management functionality
   - Source code available for security auditing
   - Community support through GitHub issues

2. Enterprise Edition: Commercial License
   - All Community Edition features
   - Enterprise features activated through license verification
   - Automatic feature enablement upon payment
   - Priority support and updates
   - Custom deployment options

#### Enterprise (Starting at €49.99/month)
- All Professional features
- Advanced SSO & security features
- Unlimited users and organizations
- Priority technical support
- Custom integrations
- Minimum commitment: 1 month

#### Custom Enterprise Solutions
- Custom feature selection
- Dedicated support team
- On-premise deployment
- Custom development
- Volume licensing

#### Payment Requirements
- Minimum payment amounts:
  - EUR: €10.00
  - USD: $10.00
  - GBP: £10.00
- Flexible duration options (1-12 months)
- Enterprise features instantly activated
- Automatic renewal available

While all source code is visible for transparency and security auditing, enterprise features require a valid license key for activation. License verification is handled through our secure licensing service integrated with Mollie payments.

#### Enterprise (€199.99/month)
- All Professional features
- Advanced SSO & security
- Unlimited users
- Priority support
- Custom integrations

#### Custom (Contact Sales)
- Custom feature selection
- Dedicated support team
- On-premise deployment
- Custom development
- Volume licensing

### Enterprise License Management
- Self-service license activation through secure API
- Automatic license verification and renewal
- Usage monitoring and analytics dashboard
- Feature-specific licensing controls
- Multi-tenant license management
- Grace period handling for renewals
- License audit logging and compliance tracking
- Automated notification system for license events

### Report Retention Management
- Configurable retention periods for different report types:
  - Daily reports (minimum 24 hours)
  - Weekly reports (minimum 7 days)
  - Monthly reports (minimum 30 days)
- Organization-specific retention policies
- Automated cleanup based on retention settings
- Default retention policy for new organizations
- Real-time policy updates and enforcement
- Retention policy audit logging

### Available Enterprise Plans
1. Professional
   - Basic enterprise features
   - Standard support
   - Up to 100 users

2. Enterprise
   - All enterprise features
   - Priority support
   - Unlimited users
   - Custom integrations

## Documentation
- [Installation Guide](docs/installation/README.md)
- [API Documentation](docs/api/README.md)
- [Deployment Guide](docs/deployment/README.md)
- [Payment Integration](docs/payment/PAYMENT_INTEGRATION.md)
- [Report Retention](docs/REPORT_RETENTION.md)

3. Custom
   - Tailored feature set
   - Dedicated support
   - Custom development
   - On-premise deployment options

## Documentation
- [Installation Guide](docs/installation/README.md)
- [API Documentation](docs/api/README.md)
- [Deployment Guide](docs/deployment/README.md)
- [Payment Integration](docs/payment/PAYMENT_INTEGRATION.md)

## Development
See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## Security
For security concerns, please see our [Security Policy](SECURITY.md).

## License
This software is available under a dual license:

1. Community Edition: GNU Affero General Public License v3.0 (AGPL-3.0)
2. Enterprise Edition: Commercial license for enterprise features

While all source code is visible for transparency and security auditing, enterprise features require a valid license for activation and use. See LICENSE.enterprise for details.

## Support
For enterprise support, licensing, and custom features, contact our support team.
