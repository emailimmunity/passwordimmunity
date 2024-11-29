package enterprise

import (
	"context"
	"time"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*Payment, error)
	VerifyPayment(ctx context.Context, paymentID string) (*Payment, error)
	GetPendingPayment(ctx context.Context, orgID, featureID string) (*Payment, error)
}

type LicenseService interface {
	GenerateLicense(ctx context.Context, req LicenseRequest) (*License, error)
	VerifyLicense(ctx context.Context, license *License) (bool, error)
	RevokeLicense(ctx context.Context, licenseID string) error
}

type FeatureService interface {
	GetFeature(featureID string) (*Feature, error)
	IsFeatureActive(ctx context.Context, orgID, featureID string) (bool, error)
	ActivateFeature(ctx context.Context, orgID, featureID string, license *License) error
	GetFeatureExpiry(ctx context.Context, orgID, featureID string) (time.Time, error)
}

type PaymentRequest struct {
	OrganizationID string
	FeatureID      string
	Amount         float64
	Description    string
}

type Payment struct {
	ID             string
	OrganizationID string
	FeatureID      string
	Amount         float64
	Status         string
	ExpiresAt      time.Time
}

type LicenseRequest struct {
	OrganizationID string
	FeatureID      string
	ExpiresAt      time.Time
}

type License struct {
	ID             string
	OrganizationID string
	FeatureID      string
	Key            string
	ExpiresAt      time.Time
}

type Feature struct {
	ID          string
	Name        string
	Description string
	Price       PriceInfo
	Tier        string
}

type PriceInfo struct {
	Monthly float64
	Yearly  float64
}

type WebhookData struct {
	PaymentID string
	Status    string
}

func (f *Feature) IsEnterprise() bool {
	return f.Tier == "enterprise" || f.Tier == "business"
}

func (f *Feature) GetPrice(planType string) float64 {
	if planType == "yearly" {
		return f.Price.Yearly
	}
	return f.Price.Monthly
}

func (p *Payment) GetExpiryTime() time.Time {
	// Default to 1 year from payment
	return p.ExpiresAt.AddDate(1, 0, 0)
}
