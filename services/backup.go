package services

import (
	"context"
	"encoding/json"
	"time"
	"archive/zip"
	"bytes"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type BackupService interface {
	CreateBackup(ctx context.Context, orgID uuid.UUID) (*models.Backup, error)
	RestoreBackup(ctx context.Context, backupID uuid.UUID) error
	ListBackups(ctx context.Context, orgID uuid.UUID) ([]models.Backup, error)
	DeleteBackup(ctx context.Context, backupID uuid.UUID) error
	ScheduleBackup(ctx context.Context, orgID uuid.UUID, schedule string) error
}

type backupService struct {
	repo        repository.Repository
	encryption  EncryptionService
	vault       VaultService
}

func NewBackupService(repo repository.Repository, encryption EncryptionService, vault VaultService) BackupService {
	return &backupService{
		repo:       repo,
		encryption: encryption,
		vault:      vault,
	}
}

func (s *backupService) CreateBackup(ctx context.Context, orgID uuid.UUID) (*models.Backup, error) {
	// Create backup container
	backup := &models.Backup{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Status:         "in_progress",
		CreatedAt:      time.Now(),
	}

	// Get all vault items for the organization
	items, err := s.vault.ListOrganizationItems(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Create ZIP archive
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add vault items to archive
	for _, item := range items {
		itemData, err := json.Marshal(item)
		if err != nil {
			continue
		}

		writer, err := zipWriter.Create(item.ID.String() + ".json")
		if err != nil {
			continue
		}

		if _, err := writer.Write(itemData); err != nil {
			continue
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	// Encrypt backup data
	encryptedData, err := s.encryption.EncryptSymmetric(buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	backup.Data = encryptedData
	backup.Status = "completed"
	backup.CompletedAt = time.Now()

	if err := s.repo.CreateBackup(ctx, backup); err != nil {
		return nil, err
	}

	// Create audit log
	metadata := createBasicMetadata("backup_created", "Backup created")
	metadata["backup_id"] = backup.ID.String()
	if err := s.createAuditLog(ctx, "backup.created", uuid.Nil, orgID, metadata); err != nil {
		return nil, err
	}

	return backup, nil
}

func (s *backupService) RestoreBackup(ctx context.Context, backupID uuid.UUID) error {
	backup, err := s.repo.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	// Decrypt backup data
	decryptedData, err := s.encryption.DecryptSymmetric(backup.Data, nil)
	if err != nil {
		return err
	}

	// Read ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(decryptedData), int64(len(decryptedData)))
	if err != nil {
		return err
	}

	// Restore vault items
	for _, file := range zipReader.File {
		var item models.VaultItem

		rc, err := file.Open()
		if err != nil {
			continue
		}

		if err := json.NewDecoder(rc).Decode(&item); err != nil {
			rc.Close()
			continue
		}
		rc.Close()

		if err := s.vault.RestoreVaultItem(ctx, &item); err != nil {
			continue
		}
	}

	// Create audit log
	metadata := createBasicMetadata("backup_restored", "Backup restored")
	metadata["backup_id"] = backupID.String()
	if err := s.createAuditLog(ctx, "backup.restored", uuid.Nil, backup.OrganizationID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *backupService) ListBackups(ctx context.Context, orgID uuid.UUID) ([]models.Backup, error) {
	return s.repo.ListBackups(ctx, orgID)
}

func (s *backupService) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	backup, err := s.repo.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("backup_deleted", "Backup deleted")
	metadata["backup_id"] = backupID.String()
	if err := s.createAuditLog(ctx, "backup.deleted", uuid.Nil, backup.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.DeleteBackup(ctx, backupID)
}

func (s *backupService) ScheduleBackup(ctx context.Context, orgID uuid.UUID, schedule string) error {
	// Create audit log
	metadata := createBasicMetadata("backup_scheduled", "Backup schedule updated")
	metadata["schedule"] = schedule
	if err := s.createAuditLog(ctx, "backup.scheduled", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdateBackupSchedule(ctx, orgID, schedule)
}
