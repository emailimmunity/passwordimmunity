package repository

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
)

type Payment struct {
    ID             uuid.UUID
    OrganizationID uuid.UUID
    Amount         float64
    Currency       string
    Status         string
    PaymentMethod  string
    MollieID       string
    LicenseType    string
    Period         string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type PaymentRepository struct {
    db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
    return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *Payment) error {
    query := `
        INSERT INTO payments (
            id, organization_id, amount, currency, status,
            payment_method, mollie_id, license_type, period,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

    _, err := r.db.ExecContext(ctx, query,
        payment.ID,
        payment.OrganizationID,
        payment.Amount,
        payment.Currency,
        payment.Status,
        payment.PaymentMethod,
        payment.MollieID,
        payment.LicenseType,
        payment.Period,
        payment.CreatedAt,
        payment.UpdatedAt,
    )

    return err
}

func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, mollieID, status string) error {
    query := `
        UPDATE payments
        SET status = $1, updated_at = $2
        WHERE mollie_id = $3
    `

    _, err := r.db.ExecContext(ctx, query, status, time.Now(), mollieID)
    return err
}

func (r *PaymentRepository) GetPaymentByMollieID(ctx context.Context, mollieID string) (*Payment, error) {
    query := `
        SELECT id, organization_id, amount, currency, status,
               payment_method, mollie_id, license_type, period,
               created_at, updated_at
        FROM payments
        WHERE mollie_id = $1
    `

    payment := &Payment{}
    err := r.db.QueryRowContext(ctx, query, mollieID).Scan(
        &payment.ID,
        &payment.OrganizationID,
        &payment.Amount,
        &payment.Currency,
        &payment.Status,
        &payment.PaymentMethod,
        &payment.MollieID,
        &payment.LicenseType,
        &payment.Period,
        &payment.CreatedAt,
        &payment.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, nil
    }

    return payment, err
}
