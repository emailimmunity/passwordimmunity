package repository

import (
    "context"
    "database/sql"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
)

type FeatureActivationRepository interface {
    Create(ctx context.Context, activation *models.FeatureActivation) error
    GetByID(ctx context.Context, id int64) (*models.FeatureActivation, error)
    GetActiveByOrganization(ctx context.Context, orgID string) ([]*models.FeatureActivation, error)
    UpdateStatus(ctx context.Context, id int64, status string) error
    GetByFeature(ctx context.Context, orgID string, featureID string) (*models.FeatureActivation, error)
    GetByBundle(ctx context.Context, orgID string, bundleID string) (*models.FeatureActivation, error)
}

type featureActivationRepo struct {
    db *sql.DB
}

func NewFeatureActivationRepository(db *sql.DB) FeatureActivationRepository {
    return &featureActivationRepo{db: db}
}

func (r *featureActivationRepo) Create(ctx context.Context, activation *models.FeatureActivation) error {
    query := `
        INSERT INTO feature_activations (
            organization_id, feature_id, bundle_id, status, expires_at,
            payment_id, currency, amount
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at`

    return r.db.QueryRowContext(
        ctx, query,
        activation.OrganizationID, activation.FeatureID, activation.BundleID,
        activation.Status, activation.ExpiresAt, activation.PaymentID,
        activation.Currency, activation.Amount,
    ).Scan(&activation.ID, &activation.CreatedAt, &activation.UpdatedAt)
}

func (r *featureActivationRepo) GetByID(ctx context.Context, id int64) (*models.FeatureActivation, error) {
    activation := &models.FeatureActivation{}
    query := `SELECT * FROM feature_activations WHERE id = $1`

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &activation.ID, &activation.OrganizationID, &activation.FeatureID,
        &activation.BundleID, &activation.Status, &activation.ExpiresAt,
        &activation.PaymentID, &activation.Currency, &activation.Amount,
        &activation.CreatedAt, &activation.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return activation, err
}

func (r *featureActivationRepo) GetActiveByOrganization(ctx context.Context, orgID string) ([]*models.FeatureActivation, error) {
    query := `
        SELECT * FROM feature_activations
        WHERE organization_id = $1
        AND status = 'active'
        AND expires_at > $2`

    rows, err := r.db.QueryContext(ctx, query, orgID, time.Now())
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var activations []*models.FeatureActivation
    for rows.Next() {
        activation := &models.FeatureActivation{}
        err := rows.Scan(
            &activation.ID, &activation.OrganizationID, &activation.FeatureID,
            &activation.BundleID, &activation.Status, &activation.ExpiresAt,
            &activation.PaymentID, &activation.Currency, &activation.Amount,
            &activation.CreatedAt, &activation.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        activations = append(activations, activation)
    }
    return activations, rows.Err()
}

func (r *featureActivationRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
    query := `UPDATE feature_activations SET status = $1 WHERE id = $2`
    _, err := r.db.ExecContext(ctx, query, status, id)
    return err
}

func (r *featureActivationRepo) GetByFeature(ctx context.Context, orgID string, featureID string) (*models.FeatureActivation, error) {
    activation := &models.FeatureActivation{}
    query := `
        SELECT * FROM feature_activations
        WHERE organization_id = $1
        AND feature_id = $2
        AND status = 'active'
        AND expires_at > $3
        ORDER BY expires_at DESC
        LIMIT 1`

    err := r.db.QueryRowContext(ctx, query, orgID, featureID, time.Now()).Scan(
        &activation.ID, &activation.OrganizationID, &activation.FeatureID,
        &activation.BundleID, &activation.Status, &activation.ExpiresAt,
        &activation.PaymentID, &activation.Currency, &activation.Amount,
        &activation.CreatedAt, &activation.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return activation, err
}

func (r *featureActivationRepo) GetByBundle(ctx context.Context, orgID string, bundleID string) (*models.FeatureActivation, error) {
    activation := &models.FeatureActivation{}
    query := `
        SELECT * FROM feature_activations
        WHERE organization_id = $1
        AND bundle_id = $2
        AND status = 'active'
        AND expires_at > $3
        ORDER BY expires_at DESC
        LIMIT 1`

    err := r.db.QueryRowContext(ctx, query, orgID, bundleID, time.Now()).Scan(
        &activation.ID, &activation.OrganizationID, &activation.FeatureID,
        &activation.BundleID, &activation.Status, &activation.ExpiresAt,
        &activation.PaymentID, &activation.Currency, &activation.Amount,
        &activation.CreatedAt, &activation.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return activation, err
}
