# Deployment Guide

## Deployment Options

### Docker Deployment (Recommended)

1. Pull the official image:
```bash
docker pull passwordimmunity/server
```

2. Create a docker-compose.yml:
```yaml
version: '3'
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_DB: passwordimmunity
      POSTGRES_USER: passwordimmunity
      POSTGRES_PASSWORD: your_secure_password
    volumes:
      - ./data/db:/var/lib/postgresql/data
    restart: always

  server:
    image: passwordimmunity/server
    depends_on:
      - db
    environment:
      DATABASE_URL: postgresql://passwordimmunity:your_secure_password@db:5432/passwordimmunity
      DOMAIN: https://your-domain.com
    ports:
      - "8000:80"
    volumes:
      - ./data/attachments:/data/attachments
    restart: always

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - server
    restart: always
```

### Manual Deployment

1. System Requirements:
   - 2+ CPU cores
   - 4GB+ RAM
   - 20GB+ storage
   - PostgreSQL 12+
   - Nginx/Apache

2. Database Setup:
```sql
CREATE DATABASE passwordimmunity;
CREATE USER passwordimmunity WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE passwordimmunity TO passwordimmunity;
```

3. Application Setup:
```bash
# Create service user
sudo useradd -r -s /bin/false passwordimmunity

# Create directories
sudo mkdir -p /opt/passwordimmunity
sudo chown passwordimmunity:passwordimmunity /opt/passwordimmunity

# Copy application files
sudo cp -r /path/to/build/* /opt/passwordimmunity/

# Create systemd service
sudo nano /etc/systemd/system/passwordimmunity.service
```

Example systemd service file:
```ini
[Unit]
Description=PasswordImmunity Server
After=network.target postgresql.service

[Service]
Type=simple
User=passwordimmunity
Group=passwordimmunity
WorkingDirectory=/opt/passwordimmunity
Environment=DATABASE_URL=postgresql://passwordimmunity:your_secure_password@localhost/passwordimmunity
ExecStart=/opt/passwordimmunity/passwordimmunity
Restart=always

[Install]
WantedBy=multi-user.target
```

4. Web Server Configuration:

Example Nginx configuration:
```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Security Hardening

1. Firewall Configuration:
```bash
# Allow only necessary ports
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

2. SSL/TLS Setup:
```bash
# Using certbot for Let's Encrypt
sudo certbot --nginx -d your-domain.com
```

3. Database Security:
- Enable SSL connections
- Regular security updates
- Automated backups
- Connection pooling

4. Application Security:
- Rate limiting
- Failed login protection
- IP filtering
- Regular security audits

## Monitoring

1. System Monitoring:
- CPU usage
- Memory usage
- Disk space
- Network traffic

2. Application Monitoring:
- Request latency
- Error rates
- Active users
- Authentication attempts

3. Database Monitoring:
- Connection count
- Query performance
- Lock waiting
- Index usage

## Backup Strategy

1. Database Backups:
```bash
# Automated daily backups
pg_dump -U passwordimmunity passwordimmunity > backup_$(date +%Y%m%d).sql
```

2. Application Data:
- Regular backups of /data directory
- Attachment storage backups
- Configuration backups

3. Backup Verification:
- Regular restore testing
- Integrity checks
- Backup rotation

## Scaling

1. Horizontal Scaling:
- Load balancer configuration
- Multiple application instances
- Read replicas for database

2. Vertical Scaling:
- CPU optimization
- Memory optimization
- Database tuning

## Troubleshooting

1. Common Issues:
- Database connection errors
- Memory issues
- CPU bottlenecks
- Network latency

2. Logging:
- Application logs
- Database logs
- Web server logs
- System logs

## Maintenance

1. Regular Tasks:
- Security updates
- Database optimization
- Log rotation
- Backup verification

2. Emergency Procedures:
- Failover process
- Data recovery
- Incident response
- Service restoration
