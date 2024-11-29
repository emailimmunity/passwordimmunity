package licensing

import (
	"time"
)

// RenewalNotification represents a license renewal notification
type RenewalNotification struct {
	OrganizationID string    `json:"organization_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	DaysRemaining  int       `json:"days_remaining"`
	Priority       string    `json:"priority"` // "low", "medium", "high", "critical"
	Message        string    `json:"message"`
}

// GetRenewalNotifications returns notifications for licenses needing renewal
func (s *Service) GetRenewalNotifications() []RenewalNotification {
	var notifications []RenewalNotification
	now := time.Now()

	for orgID, license := range s.licenses {
		daysRemaining := int(license.ExpiresAt.Sub(now).Hours() / 24)

		switch {
		case daysRemaining < 0:
			notifications = append(notifications, RenewalNotification{
				OrganizationID: orgID,
				ExpiresAt:      license.ExpiresAt,
				DaysRemaining:  0,
				Priority:       "critical",
				Message:        "License has expired. Renewal required immediately.",
			})
		case daysRemaining <= 7:
			notifications = append(notifications, RenewalNotification{
				OrganizationID: orgID,
				ExpiresAt:      license.ExpiresAt,
				DaysRemaining:  daysRemaining,
				Priority:       "high",
				Message:        "License expires in less than a week.",
			})
		case daysRemaining <= 14:
			notifications = append(notifications, RenewalNotification{
				OrganizationID: orgID,
				ExpiresAt:      license.ExpiresAt,
				DaysRemaining:  daysRemaining,
				Priority:       "medium",
				Message:        "License expires in less than two weeks.",
			})
		case daysRemaining <= 30:
			notifications = append(notifications, RenewalNotification{
				OrganizationID: orgID,
				ExpiresAt:      license.ExpiresAt,
				DaysRemaining:  daysRemaining,
				Priority:       "low",
				Message:        "License expires in less than a month.",
			})
		}
	}

	return notifications
}
