package services

import (
	"context"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type EmergencyAccessType string

const (
	EmergencyAccessView    EmergencyAccessType = "view"
	EmergencyAccessTakeover EmergencyAccessType = "takeover"
)

type EmergencyAccessService interface {
	GrantEmergencyAccess(ctx context.Context, granterID, granteeID uuid.UUID, accessType EmergencyAccessType, waitTime int) error
	RevokeEmergencyAccess(ctx context.Context, accessID uuid.UUID) error
	InitiateEmergencyAccess(ctx context.Context, accessID uuid.UUID) error
	ApproveEmergencyAccess(ctx context.Context, accessID uuid.UUID) error
	RejectEmergencyAccess(ctx context.Context, accessID uuid.UUID) error
	GetEmergencyAccess(ctx context.Context, accessID uuid.UUID) (*models.EmergencyAccess, error)
	ListGrantedAccess(ctx context.Context, userID uuid.UUID) ([]models.EmergencyAccess, error)
	ListTrustedAccess(ctx context.Context, userID uuid.UUID) ([]models.EmergencyAccess, error)
}

type emergencyAccessService struct {
	repo        repository.Repository
	audit       AuditService
	vault       VaultService
	notification NotificationService
}

func NewEmergencyAccessService(
	repo repository.Repository,
	audit AuditService,
	vault VaultService,
	notification NotificationService,
) EmergencyAccessService {
	return &emergencyAccessService{
		repo:         repo,
		audit:        audit,
		vault:        vault,
		notification: notification,
	}
}

func (s *emergencyAccessService) GrantEmergencyAccess(
	ctx context.Context,
	granterID, granteeID uuid.UUID,
	accessType EmergencyAccessType,
	waitTime int,
) error {
	access := &models.EmergencyAccess{
		ID:        uuid.New(),
		GranterID: granterID,
		GranteeID: granteeID,
		Type:      string(accessType),
		Status:    "invited",
		WaitTime:  waitTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateEmergencyAccess(ctx, access); err != nil {
		return err
	}

	// Notify grantee
	if err := s.notification.SendEmergencyAccessInvite(ctx, granterID, granteeID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("emergency_access_granted", "Emergency access granted")
	metadata["access_type"] = string(accessType)
	if err := s.createAuditLog(ctx, "emergency.access.granted", granterID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *emergencyAccessService) RevokeEmergencyAccess(ctx context.Context, accessID uuid.UUID) error {
	access, err := s.GetEmergencyAccess(ctx, accessID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteEmergencyAccess(ctx, accessID); err != nil {
		return err
	}

	// Notify grantee
	if err := s.notification.SendEmergencyAccessRevoked(ctx, access.GranterID, access.GranteeID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("emergency_access_revoked", "Emergency access revoked")
	if err := s.createAuditLog(ctx, "emergency.access.revoked", access.GranterID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *emergencyAccessService) InitiateEmergencyAccess(ctx context.Context, accessID uuid.UUID) error {
	access, err := s.GetEmergencyAccess(ctx, accessID)
	if err != nil {
		return err
	}

	access.Status = "initiated"
	access.InitiatedAt = &time.Time{}
	*access.InitiatedAt = time.Now()
	access.UpdatedAt = time.Now()

	if err := s.repo.UpdateEmergencyAccess(ctx, access); err != nil {
		return err
	}

	// Notify granter
	if err := s.notification.SendEmergencyAccessInitiated(ctx, access.GranterID, access.GranteeID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("emergency_access_initiated", "Emergency access initiated")
	if err := s.createAuditLog(ctx, "emergency.access.initiated", access.GranteeID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *emergencyAccessService) ApproveEmergencyAccess(ctx context.Context, accessID uuid.UUID) error {
	access, err := s.GetEmergencyAccess(ctx, accessID)
	if err != nil {
		return err
	}

	access.Status = "approved"
	access.ApprovedAt = &time.Time{}
	*access.ApprovedAt = time.Now()
	access.UpdatedAt = time.Now()

	if err := s.repo.UpdateEmergencyAccess(ctx, access); err != nil {
		return err
	}

	// Grant vault access based on access type
	if err := s.grantVaultAccess(ctx, access); err != nil {
		return err
	}

	// Notify grantee
	if err := s.notification.SendEmergencyAccessApproved(ctx, access.GranterID, access.GranteeID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("emergency_access_approved", "Emergency access approved")
	if err := s.createAuditLog(ctx, "emergency.access.approved", access.GranterID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *emergencyAccessService) RejectEmergencyAccess(ctx context.Context, accessID uuid.UUID) error {
	access, err := s.GetEmergencyAccess(ctx, accessID)
	if err != nil {
		return err
	}

	access.Status = "rejected"
	access.UpdatedAt = time.Now()

	if err := s.repo.UpdateEmergencyAccess(ctx, access); err != nil {
		return err
	}

	// Notify grantee
	if err := s.notification.SendEmergencyAccessRejected(ctx, access.GranterID, access.GranteeID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("emergency_access_rejected", "Emergency access rejected")
	if err := s.createAuditLog(ctx, "emergency.access.rejected", access.GranterID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *emergencyAccessService) GetEmergencyAccess(ctx context.Context, accessID uuid.UUID) (*models.EmergencyAccess, error) {
	return s.repo.GetEmergencyAccess(ctx, accessID)
}

func (s *emergencyAccessService) ListGrantedAccess(ctx context.Context, userID uuid.UUID) ([]models.EmergencyAccess, error) {
	return s.repo.ListGrantedEmergencyAccess(ctx, userID)
}

func (s *emergencyAccessService) ListTrustedAccess(ctx context.Context, userID uuid.UUID) ([]models.EmergencyAccess, error) {
	return s.repo.ListTrustedEmergencyAccess(ctx, userID)
}

func (s *emergencyAccessService) grantVaultAccess(ctx context.Context, access *models.EmergencyAccess) error {
	switch EmergencyAccessType(access.Type) {
	case EmergencyAccessView:
		return s.vault.GrantViewAccess(ctx, access.GranterID, access.GranteeID)
	case EmergencyAccessTakeover:
		return s.vault.GrantTakeoverAccess(ctx, access.GranterID, access.GranteeID)
	default:
		return errors.New("invalid emergency access type")
	}
}
