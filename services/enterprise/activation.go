package enterprise

import (
	"context"
	"errors"
	"time"
)

type ActivationService struct {
	paymentService PaymentService
	licenseService LicenseService
	featureService FeatureService
}

type ActivationRequest struct {
	OrganizationID string
	FeatureID      string
	PlanType       string // "monthly" or "yearly"
}

type ActivationStatus struct {
	Status    string    // "pending", "active", "failed"
	ExpiresAt time.Time
	Error     string
}

func NewActivationService(ps PaymentService, ls LicenseService, fs FeatureService) *ActivationService {
	return &ActivationService{
		paymentService: ps,
		licenseService: ls,
		featureService: fs,
	}
}

func (s *ActivationService) ActivateFeature(ctx context.Context, req ActivationRequest) (*ActivationStatus, error) {
	// Verify feature exists and is enterprise
	feature, err := s.featureService.GetFeature(req.FeatureID)
	if err != nil {
		return nil, err
	}
	if !feature.IsEnterprise() {
		return nil, errors.New("feature is not an enterprise feature")
	}

	// Create payment for feature
	payment, err := s.paymentService.CreatePayment(ctx, PaymentRequest{
		OrganizationID: req.OrganizationID,
		FeatureID:      req.FeatureID,
		Amount:         feature.GetPrice(req.PlanType),
		Description:    feature.Description,
	})
	if err != nil {
		return nil, err
	}

	// Return activation status
	return &ActivationStatus{
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour), // Payment expires in 24 hours
	}, nil
}

func (s *ActivationService) HandlePaymentWebhook(ctx context.Context, webhookData WebhookData) error {
	// Verify payment status
	payment, err := s.paymentService.VerifyPayment(ctx, webhookData.PaymentID)
	if err != nil {
		return err
	}

	if payment.Status != "paid" {
		return nil // Payment not completed yet
	}

	// Generate license for the feature
	license, err := s.licenseService.GenerateLicense(ctx, LicenseRequest{
		OrganizationID: payment.OrganizationID,
		FeatureID:      payment.FeatureID,
		ExpiresAt:      payment.GetExpiryTime(),
	})
	if err != nil {
		return err
	}

	// Activate the feature
	return s.featureService.ActivateFeature(ctx, payment.OrganizationID, payment.FeatureID, license)
}

func (s *ActivationService) GetActivationStatus(ctx context.Context, orgID, featureID string) (*ActivationStatus, error) {
	// Check if feature is already active
	active, err := s.featureService.IsFeatureActive(ctx, orgID, featureID)
	if err != nil {
		return nil, err
	}

	if active {
		expiry, _ := s.featureService.GetFeatureExpiry(ctx, orgID, featureID)
		return &ActivationStatus{
			Status:    "active",
			ExpiresAt: expiry,
		}, nil
	}

	// Check pending payments
	payment, err := s.paymentService.GetPendingPayment(ctx, orgID, featureID)
	if err != nil {
		return nil, err
	}

	if payment != nil {
		return &ActivationStatus{
			Status:    "pending",
			ExpiresAt: payment.ExpiresAt,
		}, nil
	}

	return &ActivationStatus{
		Status: "inactive",
	}, nil
}
