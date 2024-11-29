package enterprise

import (
	"context"
	"database/sql"
	"time"
)

// FeatureActivation represents a feature activation record
type FeatureActivation struct {
	FeatureID      string
	OrganizationID string
	ActivatedAt    time.Time
	ExpiresAt      *time.Time
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// FeatureActivationRepository defines the interface for storing feature activations
type FeatureActivationRepository interface {
	// Create stores a new feature activation
	Create(ctx context.Context, activation *FeatureActivation) error

	// Delete removes a feature activation
	Delete(ctx context.Context, featureID, organizationID string) error

	// Get retrieves a feature activation
	Get(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error)

	// GetAllActive retrieves all active feature activations for an organization
	GetAllActive(ctx context.Context, organizationID string) ([]FeatureActivation, error)

	// Update updates an existing feature activation
	Update(ctx context.Context, activation *FeatureActivation) error
}

// SQLFeatureActivationRepository implements FeatureActivationRepository using SQL
type SQLFeatureActivationRepository struct {
	db *sql.DB
}

// NewSQLFeatureActivationRepository creates a new SQLFeatureActivationRepository
func NewSQLFeatureActivationRepository(db *sql.DB) FeatureActivationRepository {
	return &SQLFeatureActivationRepository{db: db}
}

func (r *SQLFeatureActivationRepository) Create(ctx context.Context, activation *FeatureActivation) error {
	query := `
		INSERT INTO feature_activations (feature_id, organization_id, activated_at, expires_at, active)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		activation.FeatureID,
		activation.OrganizationID,
		activation.ActivatedAt,
		activation.ExpiresAt,
		activation.Active,
	)
	return err
}

func (r *SQLFeatureActivationRepository) Delete(ctx context.Context, featureID, organizationID string) error {
	query := `
		DELETE FROM feature_activations
		WHERE feature_id = $1 AND organization_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, featureID, organizationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrFeatureNotFound
	}

	return nil
}

func (r *SQLFeatureActivationRepository) Get(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error) {
	query := `
		SELECT feature_id, organization_id, activated_at, expires_at, active
		FROM feature_activations
		WHERE feature_id = $1 AND organization_id = $2
	`
	activation := &FeatureActivation{}
	err := r.db.QueryRowContext(ctx, query, featureID, organizationID).Scan(
		&activation.FeatureID,
		&activation.OrganizationID,
		&activation.ActivatedAt,
		&activation.ExpiresAt,
		&activation.Active,
	)
	if err == sql.ErrNoRows {
		return nil, ErrFeatureNotFound
	}
	if err != nil {
		return nil, err
	}
	return activation, nil
}

func (r *SQLFeatureActivationRepository) GetAllActive(ctx context.Context, organizationID string) ([]FeatureActivation, error) {
	query := `
		SELECT feature_id, organization_id, activated_at, expires_at, active
		FROM feature_activations
		WHERE organization_id = $1 AND active = true AND expires_at > $2
	`
	rows, err := r.db.QueryContext(ctx, query, organizationID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activations []FeatureActivation
	for rows.Next() {
		var activation FeatureActivation
		err := rows.Scan(
			&activation.FeatureID,
			&activation.OrganizationID,
			&activation.ActivatedAt,
			&activation.ExpiresAt,
			&activation.Active,
		)
		if err != nil {
			return nil, err
		}
		activations = append(activations, activation)
	}
	return activations, rows.Err()
}

func (r *SQLFeatureActivationRepository) Update(ctx context.Context, activation *FeatureActivation) error {
	query := `
		UPDATE feature_activations
		SET expires_at = $3, active = $4
		WHERE feature_id = $1 AND organization_id = $2
	`
	result, err := r.db.ExecContext(ctx, query,
		activation.FeatureID,
		activation.OrganizationID,
		activation.ExpiresAt,
		activation.Active,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrFeatureNotFound
	}

	return nil
}
