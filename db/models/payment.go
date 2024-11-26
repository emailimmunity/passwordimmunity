package models

import (
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusExpired   PaymentStatus = "expired"
)

type Payment struct {
	ID          string       `db:"id"`
	Status      PaymentStatus `db:"status"`
	Amount      string       `db:"amount"`
	Currency    string       `db:"currency"`
	Description string       `db:"description"`
	OrderID     string       `db:"order_id"`
	CustomerID  string       `db:"customer_id"`
	CreatedAt   time.Time    `db:"created_at"`
	PaidAt      *time.Time   `db:"paid_at"`
	ExpiresAt   *time.Time   `db:"expires_at"`
	RedirectURL string       `db:"redirect_url"`
	WebhookURL  string       `db:"webhook_url"`
	Metadata    []byte       `db:"metadata"` // JSON encoded
}

type PaymentFilter struct {
	CustomerID string
	Status     PaymentStatus
	StartDate  time.Time
	EndDate    time.Time
}
