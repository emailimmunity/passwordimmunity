package repository

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
)

type License struct {
    ID            uuid.UUID
    OrganizationID uuid.UUID
    Type          string // "free", "premium", "enterprise"
    Status        string // "active", "expired", "cancelled"
    ValidUntil    time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type LicenseRepository struct {
    db *sql.DB
}

func NewLicenseRepository(db *sql.DB) *LicenseRepository {
    return &LicenseRepository{db: db}
}

func (r *LicenseRepository) CreateLicense(ctx context.Context, license *License) error {
    query := `
        INSERT INTO licenses (
            id, organization_id, type, status,
            valid_until, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

    _, err := r.db.ExecContext(ctx, query,
        license.ID,
        license.OrganizationID,
        license.Type,
        license.Status,
        license.ValidUntil,
        license.CreatedAt,
        license.UpdatedAt,
    )

    return err
}

func (r *LicenseRepository) GetActiveLicenseByOrgID(ctx context.Context, orgID uuid.UUID) (*License, error) {
    query := `
        SELECT id, organization_id, type, status,
               valid_until, created_at, updated_at
        FROM licenses
        WHERE organization_id = $1
          AND status = 'active'
          AND valid_until > NOW()
        ORDER BY valid_until DESC
        LIMIT 1
    `

    license := &License{}
    err := r.db.QueryRowContext(ctx, query, orgID).Scan(
        &license.ID,
        &license.OrganizationID,
        &license.Type,
        &license.Status,
        &license.ValidUntil,
        &license.CreatedAt,
        &license.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }

    return license, err
}

func (r *LicenseRepository) UpdateLicenseStatus(ctx context.Context, id uuid.UUID, status string) error {
    query := `
        UPDATE licenses
        SET status = $1, updated_at = $2
        WHERE id = $3
    `

    _, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
    return err
}
