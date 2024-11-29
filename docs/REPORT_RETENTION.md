# Report Retention System

## Overview
The PasswordImmunity Report Retention System provides flexible, organization-specific control over how long different types of reports are retained in the system.

## Default Retention Periods
- Daily Reports: 7 days
- Weekly Reports: 30 days
- Monthly Reports: 365 days

## Minimum Retention Periods
To ensure proper system operation and compliance requirements:
- Daily Reports: 24 hours minimum
- Weekly Reports: 7 days minimum
- Monthly Reports: 30 days minimum

## Custom Retention Policies
Organizations can set their own retention periods that meet or exceed the minimum requirements. This allows for:
- Compliance with specific regulatory requirements
- Optimization of storage usage
- Organization-specific data retention needs

## Configuration Example
```go
policy := ReportRetentionPolicy{
    DailyReports:   2 * 24 * time.Hour,  // 2 days
    WeeklyReports:  14 * 24 * time.Hour, // 2 weeks
    MonthlyReports: 60 * 24 * time.Hour, // 60 days
}
```

## Automatic Cleanup
- Reports are automatically cleaned up based on their retention policies
- Cleanup runs daily by default
- Each organization's policy is applied independently
- Cleanup operations are logged for audit purposes

## API Usage
### Setting a Custom Policy
```http
POST /api/v1/organizations/{orgID}/retention-policy
Content-Type: application/json

{
    "daily_retention": "48h",
    "weekly_retention": "336h",
    "monthly_retention": "1440h"
}
```

### Getting Current Policy
```http
GET /api/v1/organizations/{orgID}/retention-policy
```

### Removing Custom Policy
```http
DELETE /api/v1/organizations/{orgID}/retention-policy
```

## Best Practices
1. Set retention periods based on actual business needs
2. Consider compliance requirements when setting retention periods
3. Monitor storage usage and adjust policies if needed
4. Regularly review and update retention policies
5. Keep audit logs of policy changes

## Enterprise Features
- Organization-specific retention policies
- Policy audit logging
- Storage usage reporting
- Policy change notifications
- Automated cleanup scheduling
- Retention policy templates
