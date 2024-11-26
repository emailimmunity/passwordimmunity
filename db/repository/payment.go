package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	Get(ctx context.Context, id string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
	List(ctx context.Context, filter *models.PaymentFilter) ([]*models.Payment, error)
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	query := `
		INSERT INTO payments (
			id, status, amount, currency, description, order_id,
			customer_id, created_at, paid_at, expires_at,
			redirect_url, webhook_url, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadata, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.Amount, payment.Currency,
		payment.Description, payment.OrderID, payment.CustomerID,
		payment.CreatedAt, payment.PaidAt, payment.ExpiresAt,
		payment.RedirectURL, payment.WebhookURL, metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *paymentRepository) Get(ctx context.Context, id string) (*models.Payment, error) {
	query := `
		SELECT id, status, amount, currency, description, order_id,
			   customer_id, created_at, paid_at, expires_at,
			   redirect_url, webhook_url, metadata
		FROM payments
		WHERE id = ?
	`

	payment := &models.Payment{}
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID, &payment.Status, &payment.Amount, &payment.Currency,
		&payment.Description, &payment.OrderID, &payment.CustomerID,
		&payment.CreatedAt, &payment.PaidAt, &payment.ExpiresAt,
		&payment.RedirectURL, &payment.WebhookURL, &metadata,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	payment.Metadata = metadata
	return payment, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	query := `
		UPDATE payments
		SET status = ?, amount = ?, currency = ?, description = ?,
			order_id = ?, customer_id = ?, paid_at = ?, expires_at = ?,
			redirect_url = ?, webhook_url = ?, metadata = ?
		WHERE id = ?
	`

	metadata, err := json.Marshal(payment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		payment.Status, payment.Amount, payment.Currency,
		payment.Description, payment.OrderID, payment.CustomerID,
		payment.PaidAt, payment.ExpiresAt, payment.RedirectURL,
		payment.WebhookURL, metadata, payment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	return nil
}

func (r *paymentRepository) List(ctx context.Context, filter *models.PaymentFilter) ([]*models.Payment, error) {
	query := `
		SELECT id, status, amount, currency, description, order_id,
			   customer_id, created_at, paid_at, expires_at,
			   redirect_url, webhook_url, metadata
		FROM payments
		WHERE 1=1
	`
	var args []interface{}

	if filter.CustomerID != "" {
		query += " AND customer_id = ?"
		args = append(args, filter.CustomerID)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if !filter.StartDate.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.EndDate)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()


	var payments []*models.Payment
	for rows.Next() {
		payment := &models.Payment{}
		var metadata []byte

		err := rows.Scan(
			&payment.ID, &payment.Status, &payment.Amount, &payment.Currency,
			&payment.Description, &payment.OrderID, &payment.CustomerID,
			&payment.CreatedAt, &payment.PaidAt, &payment.ExpiresAt,
			&payment.RedirectURL, &payment.WebhookURL, &metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		payment.Metadata = metadata
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate payments: %w", err)
	}

	return payments, nil
}
