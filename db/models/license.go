package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type License struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key"`
    OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
    Type          string    `gorm:"type:varchar(20);not null"` // free, premium, enterprise
    Status        string    `gorm:"type:varchar(20);not null"` // active, expired, cancelled
    ValidUntil    time.Time `gorm:"not null"`
    Features      []string  `gorm:"type:text[]"` // Array of enabled features
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`

    // Relationships
    Organization   Organization `gorm:"foreignKey:OrganizationID"`
    Payments      []Payment    `gorm:"foreignKey:LicenseID"`
}

func (l *License) BeforeCreate(tx *gorm.DB) error {
    if l.ID == uuid.Nil {
        l.ID = uuid.New()
    }
    return nil
}

// IsValid checks if the license is currently valid
func (l *License) IsValid() bool {
    return l.Status == "active" && time.Now().Before(l.ValidUntil)
}

// HasFeature checks if a specific feature is enabled for this license
func (l *License) HasFeature(feature string) bool {
    for _, f := range l.Features {
        if f == feature {
            return true
        }
    }
    return false
}
