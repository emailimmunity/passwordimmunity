package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/mollie/mollie-api-go/v2/mollie"
)

type PaymentProvider string

const (
	PaymentProviderMollie PaymentProvider = "mollie"
)

type SubscriptionPlan string

const (
	PlanFree     SubscriptionPlan = "free"
	PlanPremium  SubscriptionPlan = "premium"
	PlanBusiness SubscriptionPlan = "business"
)

type PaymentService interface {
	CreateSubscription(ctx context.Context, orgID uuid.UUID, plan SubscriptionPlan) error
	UpdateSubscription(ctx context.Context, orgID uuid.UUID, plan SubscriptionPlan) error
	CancelSubscription(ctx context.Context, orgID uuid.UUID) error
	ProcessPayment(ctx context.Context, orgID uuid.UUID, amount float64) error
	GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*models.Subscription, error)
}

type paymentService struct {
	repo      repository.Repository
	mollieClient *mollie.Client
}

func NewPaymentService(repo repository.Repository, mollieAPIKey string) (PaymentService, error) {
	client, err := mollie.NewClient(mollieAPIKey, nil)
	if err != nil {
		return nil, err
	}

	return &paymentService{
		repo:         repo,
		mollieClient: client,
	}, nil
}

func (s *paymentService) CreateSubscription(ctx context.Context, orgID uuid.UUID, plan SubscriptionPlan) error {
	subscription := &models.Subscription{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Plan:           string(plan),
		Status:         "active",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(1, 0, 0), // 1 year subscription
	}

	// Create audit log
	metadata := createBasicMetadata("subscription_created", "Subscription created")
	metadata["plan"] = string(plan)
	if err := s.createAuditLog(ctx, "subscription.created", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.CreateSubscription(ctx, subscription)
}

func (s *paymentService) UpdateSubscription(ctx context.Context, orgID uuid.UUID, plan SubscriptionPlan) error {
	subscription, err := s.GetSubscriptionStatus(ctx, orgID)
	if err != nil {
		return err
	}

	subscription.Plan = string(plan)
	subscription.UpdatedAt = time.Now()

	// Create audit log
	metadata := createBasicMetadata("subscription_updated", "Subscription updated")
	metadata["plan"] = string(plan)
	if err := s.createAuditLog(ctx, "subscription.updated", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdateSubscription(ctx, subscription)
}

func (s *paymentService) CancelSubscription(ctx context.Context, orgID uuid.UUID) error {
	subscription, err := s.GetSubscriptionStatus(ctx, orgID)
	if err != nil {
		return err
	}

	subscription.Status = "cancelled"
	subscription.UpdatedAt = time.Now()

	// Create audit log
	metadata := createBasicMetadata("subscription_cancelled", "Subscription cancelled")
	if err := s.createAuditLog(ctx, "subscription.cancelled", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdateSubscription(ctx, subscription)
}

func (s *paymentService) ProcessPayment(ctx context.Context, orgID uuid.UUID, amount float64) error {
	payment, err := s.mollieClient.Payments.Create(&mollie.Payment{
		Amount: &mollie.Amount{
			Currency: "EUR",
			Value:    fmt.Sprintf("%.2f", amount),
		},
		Description: fmt.Sprintf("Subscription payment for organization %s", orgID),
		RedirectURL: "https://passwordimmunity.com/payment/callback",
		WebhookURL:  "https://passwordimmunity.com/payment/webhook",
	})
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("payment_processed", "Payment processed")
	metadata["amount"] = fmt.Sprintf("%.2f", amount)
	metadata["payment_id"] = payment.ID
	if err := s.createAuditLog(ctx, "payment.processed", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *paymentService) GetSubscriptionStatus(ctx context.Context, orgID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscription(ctx, orgID)
}
