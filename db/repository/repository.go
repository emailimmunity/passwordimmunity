package repository

import (
	"context"
	"errors"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the interface for database operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// Organization operations
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, org *models.Organization) error
	DeleteOrganization(ctx context.Context, id uuid.UUID) error

	// Role operations
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// VaultItem operations
	CreateVaultItem(ctx context.Context, item *models.VaultItem) error
	GetVaultItemByID(ctx context.Context, id uuid.UUID) (*models.VaultItem, error)
	UpdateVaultItem(ctx context.Context, item *models.VaultItem) error
	DeleteVaultItem(ctx context.Context, id uuid.UUID) error

	// Audit operations
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, userID, orgID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Implementation of Repository interface
func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// Organization operations
func (r *repository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *repository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	if err := r.db.WithContext(ctx).First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *repository) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *repository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Organization{}, id).Error
}

// Role operations
func (r *repository) CreateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *repository) UpdateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *repository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

// VaultItem operations
func (r *repository) CreateVaultItem(ctx context.Context, item *models.VaultItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *repository) GetVaultItemByID(ctx context.Context, id uuid.UUID) (*models.VaultItem, error) {
	var item models.VaultItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *repository) UpdateVaultItem(ctx context.Context, item *models.VaultItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *repository) DeleteVaultItem(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.VaultItem{}, id).Error
}

// Audit operations
func (r *repository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) GetAuditLogs(ctx context.Context, userID, orgID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := r.db.WithContext(ctx)

	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}
	if orgID != uuid.Nil {
		query = query.Where("organization_id = ?", orgID)
	}

	err := query.Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, err
	}
	return logs, nil
}
