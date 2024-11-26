package services

import (
	"context"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type ShareType string

const (
	ShareTypeView    ShareType = "view"
	ShareTypeEdit    ShareType = "edit"
	ShareTypeManage  ShareType = "manage"
)

type SharingService interface {
	ShareItem(ctx context.Context, itemID uuid.UUID, shareConfig models.ShareConfig) error
	UpdateSharing(ctx context.Context, shareID uuid.UUID, shareConfig models.ShareConfig) error
	RevokeSharing(ctx context.Context, shareID uuid.UUID) error
	GetSharing(ctx context.Context, shareID uuid.UUID) (*models.ShareConfig, error)
	ListSharedItems(ctx context.Context, userID uuid.UUID) ([]models.SharedItem, error)
	ListItemShares(ctx context.Context, itemID uuid.UUID) ([]models.ShareConfig, error)
	ValidateShareAccess(ctx context.Context, shareID uuid.UUID, accessType ShareType) error
}

type sharingService struct {
	repo        repository.Repository
	audit       AuditService
	encryption  EncryptionService
	notification NotificationService
}

func NewSharingService(
	repo repository.Repository,
	audit AuditService,
	encryption EncryptionService,
	notification NotificationService,
) SharingService {
	return &sharingService{
		repo:         repo,
		audit:        audit,
		encryption:   encryption,
		notification: notification,
	}
}

func (s *sharingService) ShareItem(ctx context.Context, itemID uuid.UUID, shareConfig models.ShareConfig) error {
	shareConfig.ID = uuid.New()
	shareConfig.ItemID = itemID
	shareConfig.Status = "active"
	shareConfig.CreatedAt = time.Now()
	shareConfig.UpdatedAt = time.Now()

	// Re-encrypt item key for recipient
	encryptedKey, err := s.encryption.ReEncryptKey(ctx, shareConfig.ItemKey, shareConfig.RecipientID)
	if err != nil {
		return err
	}
	shareConfig.EncryptedKey = encryptedKey

	if err := s.repo.CreateShareConfig(ctx, &shareConfig); err != nil {
		return err
	}

	// Notify recipient
	if err := s.notification.SendShareNotification(ctx, shareConfig); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("item_shared", "Item shared")
	metadata["share_type"] = string(shareConfig.Type)
	metadata["recipient_id"] = shareConfig.RecipientID.String()
	if err := s.createAuditLog(ctx, "sharing.created", shareConfig.OwnerID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sharingService) UpdateSharing(ctx context.Context, shareID uuid.UUID, shareConfig models.ShareConfig) error {
	existing, err := s.GetSharing(ctx, shareID)
	if err != nil {
		return err
	}

	shareConfig.ID = existing.ID
	shareConfig.ItemID = existing.ItemID
	shareConfig.CreatedAt = existing.CreatedAt
	shareConfig.UpdatedAt = time.Now()

	if err := s.repo.UpdateShareConfig(ctx, &shareConfig); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("share_updated", "Share configuration updated")
	metadata["share_type"] = string(shareConfig.Type)
	if err := s.createAuditLog(ctx, "sharing.updated", shareConfig.OwnerID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sharingService) RevokeSharing(ctx context.Context, shareID uuid.UUID) error {
	share, err := s.GetSharing(ctx, shareID)
	if err != nil {
		return err
	}

	share.Status = "revoked"
	share.RevokedAt = &time.Time{}
	*share.RevokedAt = time.Now()
	share.UpdatedAt = time.Now()

	if err := s.repo.UpdateShareConfig(ctx, share); err != nil {
		return err
	}

	// Notify recipient
	if err := s.notification.SendShareRevocationNotification(ctx, *share); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("share_revoked", "Share access revoked")
	metadata["recipient_id"] = share.RecipientID.String()
	if err := s.createAuditLog(ctx, "sharing.revoked", share.OwnerID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sharingService) GetSharing(ctx context.Context, shareID uuid.UUID) (*models.ShareConfig, error) {
	return s.repo.GetShareConfig(ctx, shareID)
}

func (s *sharingService) ListSharedItems(ctx context.Context, userID uuid.UUID) ([]models.SharedItem, error) {
	return s.repo.ListSharedItems(ctx, userID)
}

func (s *sharingService) ListItemShares(ctx context.Context, itemID uuid.UUID) ([]models.ShareConfig, error) {
	return s.repo.ListItemShares(ctx, itemID)
}

func (s *sharingService) ValidateShareAccess(ctx context.Context, shareID uuid.UUID, accessType ShareType) error {
	share, err := s.GetSharing(ctx, shareID)
	if err != nil {
		return err
	}

	if share.Status != "active" {
		return errors.New("share is not active")
	}

	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return errors.New("share has expired")
	}

	// Validate access type
	switch accessType {
	case ShareTypeView:
		// View access is allowed for all share types
		return nil
	case ShareTypeEdit:
		if share.Type != string(ShareTypeEdit) && share.Type != string(ShareTypeManage) {
			return errors.New("insufficient share permissions for edit access")
		}
	case ShareTypeManage:
		if share.Type != string(ShareTypeManage) {
			return errors.New("insufficient share permissions for manage access")
		}
	default:
		return errors.New("invalid access type")
	}

	return nil
}
