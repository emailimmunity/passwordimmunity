package licensing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/audit"
)

type RetentionPolicyAuditEvent struct {
	OrganizationID string    `json:"organization_id"`
	Action         string    `json:"action"` // "set", "remove"
	OldPolicy      *Policy   `json:"old_policy,omitempty"`
	NewPolicy      *Policy   `json:"new_policy,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	UserID         string    `json:"user_id"`
}

type Policy struct {
	DailyReports   time.Duration `json:"daily_reports"`
	WeeklyReports  time.Duration `json:"weekly_reports"`
	MonthlyReports time.Duration `json:"monthly_reports"`
}

func (s *Service) logRetentionPolicyChange(ctx context.Context, orgID string, action string, oldPolicy, newPolicy *Policy) error {
	userID := ctx.Value("user_id").(string)

	event := RetentionPolicyAuditEvent{
		OrganizationID: orgID,
		Action:         action,
		OldPolicy:      oldPolicy,
		NewPolicy:      newPolicy,
		Timestamp:      time.Now(),
		UserID:         userID,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.auditLogger.Log(ctx, "retention_policy_change", string(eventJSON))
}
