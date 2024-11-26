package payment

import (
	"context"
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

type CreatePaymentRequest struct {
	Amount      string
	Currency    string
	Description string
	OrderID     string
	CustomerID  string
	RedirectURL string
	WebhookURL  string
	Metadata    map[string]string
}

type Payment struct {
	ID          string
	Status      PaymentStatus
	Amount      string
	Currency    string
	Description string
	OrderID     string
	CustomerID  string
	CreatedAt   time.Time
	PaidAt      *time.Time
	ExpiresAt   *time.Time
	RedirectURL string
	WebhookURL  string
	Metadata    map[string]string
}

type PaymentService interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	CancelPayment(ctx context.Context, paymentID string) error
	HandleWebhook(ctx context.Context, req *WebhookRequest) error
}

type WebhookRequest struct {
	PaymentID string
	Status    PaymentStatus
	Metadata  map[string]string
}

type Config struct {
	APIKey         string
	WebhookSecret  string
	APIEndpoint    string
	WebhookPath    string
	SuccessURL     string
	FailureURL     string
	Currency       string
	RetryAttempts  int
	RetryDelay     time.Duration
	TimeoutSeconds int
}

func NewConfig() *Config {
	return &Config{
		APIEndpoint:    "https://api.mollie.com/v2",
		WebhookPath:    "/webhooks/payment",
		Currency:       "EUR",
		RetryAttempts:  3,
		RetryDelay:     time.Second * 2,
		TimeoutSeconds: 30,
	}
}
