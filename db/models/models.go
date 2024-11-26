package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common fields
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User represents a system user
type User struct {
	Base
	Email            string `gorm:"uniqueIndex;not null"`
	Name             string `gorm:"not null"`
	PasswordHash     string `gorm:"not null"`
	TwoFactorEnabled bool   `gorm:"default:false"`
	TwoFactorSecret  string
	Organizations    []Organization `gorm:"many2many:user_organizations;"`
}

// Organization represents a group of users
type Organization struct {
	Base
	Name  string `gorm:"not null"`
	Type  string `gorm:"not null"`
	Users []User `gorm:"many2many:user_organizations;"`
	Roles []Role
}

// Role represents a set of permissions
type Role struct {
	Base
	Name           string `gorm:"not null"`
	Description    string
	OrganizationID uuid.UUID
	Organization   Organization
	Permissions    []Permission `gorm:"many2many:role_permissions;"`
}

// Permission represents an action that can be performed
type Permission struct {
	Base
	Name        string `gorm:"not null"`
	Description string
	Roles       []Role `gorm:"many2many:role_permissions;"`
}

// VaultItem represents an encrypted item in a user's vault
type VaultItem struct {
	Base
	UserID         uuid.UUID
	User           User
	OrganizationID uuid.UUID
	Organization   Organization
	Type           string `gorm:"not null"`
	Name           string `gorm:"not null"`
	EncryptedData  string `gorm:"not null;type:text"`
}

// AuditLog represents a system audit event
type AuditLog struct {
	Base
	UserID         uuid.UUID
	User           User
	OrganizationID uuid.UUID
	Organization   Organization
	Action         string `gorm:"not null"`
	Details        []byte `gorm:"type:jsonb"`
}

// Payment represents a payment transaction
type Payment struct {
    Base
    OrganizationID uuid.UUID
    Organization   Organization
    ProviderID     string    `gorm:"uniqueIndex"`
    Amount         string
    Currency       string
    Status         string
    LicenseType    string
    Period         string
    PaidAt         time.Time
}

// License represents an organization's license information
type License struct {
    Base
    OrganizationID uuid.UUID
    Organization   Organization
    Type           string    // "free", "premium", "enterprise"
    Status         string    // "active", "expired", "cancelled"
    ExpiresAt      time.Time
    Features       []string  `gorm:"type:text[]"`
    PaymentID      uuid.UUID
    Payment        Payment
}

// FeatureFlag represents a configurable feature flag
type FeatureFlag struct {
	Base
	Name        string            `gorm:"uniqueIndex;not null"`
	Enabled     bool              `gorm:"default:false"`
	Options     FeatureFlagOptions `gorm:"type:jsonb"`
	Description string
}

// FeatureFlagOptions contains configuration options for feature flags
type FeatureFlagOptions struct {
	RequiresEnterprise bool        `json:"requires_enterprise"`
	RequiresPremium    bool        `json:"requires_premium"`
	AllowedOrgs        []uuid.UUID `json:"allowed_orgs"`
	RolloutPercentage  float64     `json:"rollout_percentage"`
	StartsAt           time.Time    `json:"starts_at"`
	ExpiresAt          time.Time    `json:"expires_at"`
	Description        string       `json:"description"`
}

// BeforeCreate is called before creating a record
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}
