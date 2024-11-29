package payment

import (
	"time"
	"github.com/shopspring/decimal"
)

// Amount represents a monetary amount with currency
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// PaymentMetadata contains structured metadata for payments
type PaymentMetadata struct {
	OrganizationID string   `json:"organization_id"`
	Features       []string `json:"features,omitempty"`
	Bundles        []string `json:"bundles,omitempty"`
	Duration       string   `json:"duration"`
}

// PaymentRequest represents a payment creation request
type PaymentRequest struct {
	Amount      Amount          `json:"amount"`
	Description string         `json:"description"`
	RedirectURL string         `json:"redirect_url"`
	WebhookURL  string         `json:"webhook_url"`
	Metadata    PaymentMetadata `json:"metadata"`
}

// PaymentResponse represents a payment response from the provider
type PaymentResponse struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	Amount      Amount          `json:"amount"`
	RedirectURL string         `json:"redirect_url"`
	WebhookURL  string         `json:"webhook_url"`
	Metadata    PaymentMetadata `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	PaidAt      *time.Time     `json:"paid_at,omitempty"`
}

// Payment represents an internal payment record
type Payment struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	Description string         `json:"description"`
	Metadata    PaymentMetadata `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	PaidAt      *time.Time     `json:"paid_at,omitempty"`
}
