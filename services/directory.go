package services

import (
	"context"
	"time"
	"sync"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/go-ldap/ldap/v3"
)

type DirectoryType string

const (
	DirectoryTypeLDAP DirectoryType = "ldap"
	DirectoryTypeAD   DirectoryType = "active_directory"
	DirectoryTypeOkta DirectoryType = "okta"
)

type DirectoryService interface {
	ConfigureDirectory(ctx context.Context, orgID uuid.UUID, config models.DirectoryConfig) error
	SyncDirectory(ctx context.Context, orgID uuid.UUID) error
	GetSyncStatus(ctx context.Context, orgID uuid.UUID) (*models.DirectorySync, error)
	ListDirectoryUsers(ctx context.Context, orgID uuid.UUID) ([]models.DirectoryUser, error)
	ValidateDirectoryConfig(ctx context.Context, config models.DirectoryConfig) error
	GetDirectoryGroups(ctx context.Context, orgID uuid.UUID) ([]models.DirectoryGroup, error)
}

type directoryService struct {
	repo        repository.Repository
	audit       AuditService
	licensing   LicensingService
	sync        sync.Mutex
}

func NewDirectoryService(repo repository.Repository, audit AuditService, licensing LicensingService) DirectoryService {
	return &directoryService{
		repo:      repo,
		audit:     audit,
		licensing: licensing,
	}
}

func (s *directoryService) ConfigureDirectory(ctx context.Context, orgID uuid.UUID, config models.DirectoryConfig) error {
	// Check if organization has directory sync access
	hasAccess, err := s.licensing.CheckFeatureAccess(ctx, orgID, "directory_sync")
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("directory sync not available in current license")
	}

	if err := s.ValidateDirectoryConfig(ctx, config); err != nil {
		return err
	}

	config.ID = uuid.New()
	config.OrganizationID = orgID
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := s.repo.CreateDirectoryConfig(ctx, &config); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("directory_configured", "Directory sync configured")
	metadata["directory_type"] = string(config.Type)
	if err := s.createAuditLog(ctx, "directory.configured", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *directoryService) SyncDirectory(ctx context.Context, orgID uuid.UUID) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	config, err := s.repo.GetDirectoryConfig(ctx, orgID)
	if err != nil {
		return err
	}

	sync := &models.DirectorySync{
		ID:             uuid.New(),
		OrganizationID: orgID,
		StartTime:      time.Now(),
		Status:         "in_progress",
	}

	if err := s.repo.CreateDirectorySync(ctx, sync); err != nil {
		return err
	}

	// Perform directory synchronization based on directory type
	var syncErr error
	switch config.Type {
	case DirectoryTypeLDAP:
		syncErr = s.syncLDAP(ctx, config)
	case DirectoryTypeAD:
		syncErr = s.syncActiveDirectory(ctx, config)
	case DirectoryTypeOkta:
		syncErr = s.syncOkta(ctx, config)
	default:
		syncErr = errors.New("unsupported directory type")
	}

	sync.EndTime = time.Now()
	if syncErr != nil {
		sync.Status = "failed"
		sync.Error = syncErr.Error()
	} else {
		sync.Status = "completed"
	}

	if err := s.repo.UpdateDirectorySync(ctx, sync); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("directory_synced", "Directory sync completed")
	metadata["status"] = sync.Status
	if err := s.createAuditLog(ctx, "directory.synced", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return syncErr
}

func (s *directoryService) GetSyncStatus(ctx context.Context, orgID uuid.UUID) (*models.DirectorySync, error) {
	return s.repo.GetLatestDirectorySync(ctx, orgID)
}

func (s *directoryService) ListDirectoryUsers(ctx context.Context, orgID uuid.UUID) ([]models.DirectoryUser, error) {
	return s.repo.ListDirectoryUsers(ctx, orgID)
}

func (s *directoryService) ValidateDirectoryConfig(ctx context.Context, config models.DirectoryConfig) error {
	switch config.Type {
	case DirectoryTypeLDAP:
		return s.validateLDAPConfig(config)
	case DirectoryTypeAD:
		return s.validateADConfig(config)
	case DirectoryTypeOkta:
		return s.validateOktaConfig(config)
	default:
		return errors.New("unsupported directory type")
	}
}

func (s *directoryService) GetDirectoryGroups(ctx context.Context, orgID uuid.UUID) ([]models.DirectoryGroup, error) {
	return s.repo.ListDirectoryGroups(ctx, orgID)
}


// Private helper methods for specific directory types
func (s *directoryService) syncLDAP(ctx context.Context, config models.DirectoryConfig) error {
	// Implement LDAP synchronization
	return nil
}

func (s *directoryService) syncActiveDirectory(ctx context.Context, config models.DirectoryConfig) error {
	// Implement Active Directory synchronization
	return nil
}

func (s *directoryService) syncOkta(ctx context.Context, config models.DirectoryConfig) error {
	// Implement Okta synchronization
	return nil
}

func (s *directoryService) validateLDAPConfig(config models.DirectoryConfig) error {
	// Implement LDAP config validation
	return nil
}

func (s *directoryService) validateADConfig(config models.DirectoryConfig) error {
	// Implement Active Directory config validation
	return nil
}

func (s *directoryService) validateOktaConfig(config models.DirectoryConfig) error {
	// Implement Okta config validation
	return nil
}
