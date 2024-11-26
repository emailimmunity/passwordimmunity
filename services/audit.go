package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type AuditEventType string

const (
	AuditEventUserCreated           AuditEventType = "user.created"
	AuditEventUserLogin             AuditEventType = "user.login"
	AuditEventUserPasswordChanged   AuditEventType = "user.password_changed"
	AuditEventOrganizationCreated   AuditEventType = "organization.created"
	AuditEventOrganizationModified  AuditEventType = "organization.modified"
	AuditEventRoleCreated           AuditEventType = "role.created"
	AuditEventRoleModified          AuditEventType = "role.modified"
	AuditEventVaultItemCreated      AuditEventType = "vault.item_created"
	AuditEventVaultItemAccessed     AuditEventType = "vault.item_accessed"
	AuditEventVaultItemModified     AuditEventType = "vault.item_modified"
	AuditEventPermissionAssigned    AuditEventType = "permission.assigned"
	AuditEventUserAddedToOrg        AuditEventType = "organization.user_added"
	AuditEventUserRemovedFromOrg    AuditEventType = "organization.user_removed"
)

type AuditMetadata map[string]interface{}

func (s *service) createAuditLog(ctx context.Context, eventType AuditEventType, userID, orgID uuid.UUID, metadata AuditMetadata) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	log := &models.AuditLog{
		EventType:      string(eventType),
		UserID:        userID,
		OrganizationID: orgID,
		Metadata:      metadataBytes,
		Timestamp:     time.Now(),
	}

	return s.repo.CreateAuditLog(ctx, log)
}

func (s *service) GetAuditLogs(ctx context.Context, userID, orgID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	// Verify user has permission to view audit logs
	hasAccess, err := s.hasPermission(ctx, userID, orgID, "view_audit_logs")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	return s.repo.GetAuditLogs(ctx, userID, orgID, limit, offset)
}

// Helper function to create audit log with basic metadata
func createBasicMetadata(action, detail string) AuditMetadata {
	return AuditMetadata{
		"action":     action,
		"detail":     detail,
		"timestamp":  time.Now().Format(time.RFC3339),
		"ip_address": "", // To be filled by the API layer
	}
}
