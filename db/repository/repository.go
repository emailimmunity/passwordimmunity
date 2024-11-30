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

	// Payment operations
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	GetPaymentByProviderID(ctx context.Context, providerID string) (*models.Payment, error)
	UpdatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentsByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Payment, error)

	// License operations
	CreateLicense(ctx context.Context, license *models.License) error
	GetLicenseByID(ctx context.Context, id uuid.UUID) (*models.License, error)
	GetActiveLicenseByOrganization(ctx context.Context, orgID uuid.UUID) (*models.License, error)
	UpdateLicense(ctx context.Context, license *models.License) error
	GetExpiredLicenses(ctx context.Context) ([]models.License, error)
	DeactivateOrganizationLicenses(ctx context.Context, orgID uuid.UUID) error
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

// Payment operations
func (r *repository) CreatePayment(ctx context.Context, payment *models.Payment) error {
    return r.db.WithContext(ctx).Create(payment).Error
}

func (r *repository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
    var payment models.Payment
    if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &payment, nil
}

func (r *repository) GetPaymentByProviderID(ctx context.Context, providerID string) (*models.Payment, error) {
    var payment models.Payment
    if err := r.db.WithContext(ctx).Where("provider_id = ?", providerID).First(&payment).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &payment, nil
}

func (r *repository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
    return r.db.WithContext(ctx).Save(payment).Error
}

func (r *repository) GetPaymentsByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Payment, error) {
    var payments []models.Payment
    if err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&payments).Error; err != nil {
        return nil, err
    }
    return payments, nil
}

// License operations
func (r *repository) CreateLicense(ctx context.Context, license *models.License) error {
    return r.db.WithContext(ctx).Create(license).Error
}

func (r *repository) GetLicenseByID(ctx context.Context, id uuid.UUID) (*models.License, error) {
    var license models.License
    if err := r.db.WithContext(ctx).First(&license, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &license, nil
}

func (r *repository) GetActiveLicenseByOrganization(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
    var license models.License
    if err := r.db.WithContext(ctx).
        Where("organization_id = ? AND status = ? AND expires_at > ?", orgID, "active", time.Now()).
        First(&license).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &license, nil
}

func (r *repository) UpdateLicense(ctx context.Context, license *models.License) error {
    return r.db.WithContext(ctx).Save(license).Error
}

func (r *repository) GetExpiredLicenses(ctx context.Context) ([]models.License, error) {
    var licenses []models.License
    if err := r.db.WithContext(ctx).
        Where("status = ? AND expires_at <= ?", "active", time.Now()).
        Find(&licenses).Error; err != nil {
        return nil, err
    }
    return licenses, nil
}

func (r *repository) DeactivateOrganizationLicenses(ctx context.Context, orgID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&models.License{}).
        Where("organization_id = ? AND status = ?", orgID, "active").
        Updates(map[string]interface{}{
            "status": "inactive",
            "updated_at": time.Now(),
        }).Error
}
