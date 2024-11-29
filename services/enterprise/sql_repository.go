package enterprise

import (
	"context"
	"database/sql"
	"time"
)

type sqlFeatureActivationRepository struct {
	db *sql.DB
}

func NewSQLFeatureActivationRepository(db *sql.DB) FeatureActivationRepository {
	return &sqlFeatureActivationRepository{db: db}
}

func (r *sqlFeatureActivationRepository) Create(ctx context.Context, activation *FeatureActivation) error {
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

func (r *sqlFeatureActivationRepository) Get(ctx context.Context, featureID, organizationID string) (*FeatureActivation, error) {
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

func (r *sqlFeatureActivationRepository) Update(ctx context.Context, activation *FeatureActivation) error {
	query := `
		UPDATE feature_activations
		SET expires_at = $1, active = $2
		WHERE feature_id = $3 AND organization_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		activation.ExpiresAt,
		activation.Active,
		activation.FeatureID,
		activation.OrganizationID,
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

func (r *sqlFeatureActivationRepository) Delete(ctx context.Context, featureID, organizationID string) error {
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

func (r *sqlFeatureActivationRepository) GetAllActive(ctx context.Context, organizationID string) ([]FeatureActivation, error) {
	query := `
		SELECT feature_id, organization_id, activated_at, expires_at, active
		FROM feature_activations
		WHERE organization_id = $1 AND active = true AND expires_at > NOW()
	`

	rows, err := r.db.QueryContext(ctx, query, organizationID)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return activations, nil
}
