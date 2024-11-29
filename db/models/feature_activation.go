package models

import (
    "time"
)

type FeatureActivation struct {
    ID             int64     `db:"id"`
    OrganizationID string    `db:"organization_id"`
    FeatureID      *string   `db:"feature_id"`
    BundleID       *string   `db:"bundle_id"`
    Status         string    `db:"status"`
    ExpiresAt      time.Time `db:"expires_at"`
    PaymentID      string    `db:"payment_id"`
    Currency       string    `db:"currency"`
    Amount         float64   `db:"amount"`
    CreatedAt      time.Time `db:"created_at"`
    UpdatedAt      time.Time `db:"updated_at"`
}

func (fa *FeatureActivation) IsActive() bool {
    return fa.Status == "active" && time.Now().Before(fa.ExpiresAt)
}

func (fa *FeatureActivation) IsExpired() bool {
    return fa.Status == "expired" || time.Now().After(fa.ExpiresAt)
}

func (fa *FeatureActivation) IsCancelled() bool {
    return fa.Status == "cancelled"
}
