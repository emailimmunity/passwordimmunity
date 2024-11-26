# Installation Guide

## Prerequisites

- Rust 1.70.0 or later
- PostgreSQL 12.0 or later
- Node.js 18.0 or later (for web vault)
- Docker (optional, for containerized deployment)

## Quick Start

### Using Docker

```bash
docker pull passwordimmunity/server
docker run -d --name passwordimmunity \
  -e DATABASE_URL=postgresql://user:password@db:5432/passwordimmunity \
  -p 8000:80 \
  passwordimmunity/server
```

### Manual Installation

1. Clone the repository:
```bash
git clone https://github.com/emailimmunity/passwordimmunity
cd passwordimmunity
```

2. Set up the database:
```bash
# Create PostgreSQL database
createdb passwordimmunity
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Build and run:
```bash
cargo build --release
./target/release/passwordimmunity
```

## Configuration

### Environment Variables

- `DATABASE_URL`: PostgreSQL connection string
- `DOMAIN`: Your domain name
- `SMTP_HOST`: SMTP server for email notifications
- `SMTP_PORT`: SMTP port
- `SMTP_SSL`: Enable/disable SSL for SMTP
- `SMTP_USERNAME`: SMTP authentication username
- `SMTP_PASSWORD`: SMTP authentication password

### Enterprise Features

Enterprise features can be enabled by setting the following environment variables:

- `ENABLE_SSO`: Enable SSO integration
- `ENABLE_ADVANCED_ROLES`: Enable advanced role management
- `ENABLE_AUDIT_LOGS`: Enable detailed audit logging
- `ENABLE_API_ACCESS`: Enable enterprise API access

## Security Considerations

1. Always use HTTPS in production
2. Set up proper firewall rules
3. Configure secure database access
4. Implement regular backups
5. Enable audit logging

## Troubleshooting

Common issues and their solutions:

1. Database connection errors:
   - Verify PostgreSQL is running
   - Check connection string
   - Ensure database user has proper permissions

2. Web vault access issues:
   - Verify domain configuration
   - Check SSL/TLS setup
   - Confirm firewall settings

3. Authentication problems:
   - Verify SMTP configuration
   - Check SSO settings if enabled
   - Confirm user permissions

## Support

For enterprise support and feature requests, please contact support@emailimmunity.com

For community support:
- GitHub Issues
- Community Forums
- Documentation Wiki
