package db

import (
	"time"
	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key"`
	Email         string    `gorm:"uniqueIndex;not null"`
	Name          string
	PasswordHash  string    `gorm:"not null"`
	TwoFactorType string    // Type of 2FA enabled (totp, yubikey, etc.)
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	Organizations []Organization `gorm:"many2many:user_organizations;"`
	Roles         []Role        `gorm:"many2many:user_roles;"`
}

// Organization represents a multi-tenant organization
type Organization struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"not null"`
	Type        string    // enterprise, team, personal
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Users       []User    `gorm:"many2many:user_organizations;"`
	Roles       []Role    `gorm:"many2many:organization_roles;"`
}

// Role represents a system role with permissions
type Role struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key"`
	Name          string    `gorm:"not null"`
	Description   string
	Permissions   []Permission `gorm:"many2many:role_permissions;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Organizations []Organization `gorm:"many2many:organization_roles;"`
}

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Roles       []Role    `gorm:"many2many:role_permissions;"`
}

// VaultItem represents an encrypted vault item
type VaultItem struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index"`
	Type           string    // login, card, identity, secure_note
	Name           string
	Notes          string
	Data           []byte    // Encrypted data
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// AuditLog represents system audit logs
type AuditLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID `gorm:"type:uuid;index"`
	Action         string
	Details        string
	IPAddress      string
	CreatedAt      time.Time
}
