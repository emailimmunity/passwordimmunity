package payment

import (
	"context"
	"fmt"
	"time"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type PaymentService struct {
	provider     PaymentProvider
	notifications *NotificationService
	webhookHandler *WebhookHandler
}

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
}

type PaymentRequest struct {
	Amount      Amount  `json:"amount" validate:"required"`
	Description string  `json:"description" validate:"required"`
	RedirectURL string  `json:"redirect_url" validate:"required,url"`
	WebhookURL  string  `json:"webhook_url" validate:"required,url"`
	Metadata    PaymentMetadata `json:"metadata" validate:"required"`
}

type PaymentMetadata struct {
	OrganizationID string   `json:"organization_id" validate:"required"`
	Features       []string `json:"features" validate:"omitempty,dive,required"`
	Bundles        []string `json:"bundles" validate:"omitempty,dive,required"`
	Duration       string   `json:"duration" validate:"required"`
}

type Amount struct {
	Currency string `json:"currency" validate:"required,oneof=EUR USD GBP"`
	Value    string `json:"value" validate:"required"`
}

func (a *Amount) Validate() error {
	if err := validate.Struct(a); err != nil {
		return err
	}

	value, err := decimal.NewFromString(a.Value)
	if err != nil {
		return fmt.Errorf("invalid amount value %q: %w", a.Value, err)
	}

	// Check for negative values
	if value.IsNegative() {
		return fmt.Errorf("amount value cannot be negative")
	}

	// Check minimum amount
	if minAmount, ok := minAmounts[a.Currency]; ok && value.LessThan(minAmount) {
		return fmt.Errorf("amount must be at least %v %s", minAmount, a.Currency)
	}

	return nil
}

type PaymentResponse struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	Amount      Amount         `json:"amount"`
	RedirectURL string         `json:"redirect_url"`
	WebhookURL  string         `json:"webhook_url"`
	Metadata    PaymentMetadata `json:"metadata"`
}

var (
	validate = validator.New()
	minAmounts = map[string]decimal.Decimal{
		"EUR": decimal.NewFromFloat(1.00),
		"USD": decimal.NewFromFloat(1.00),
		"GBP": decimal.NewFromFloat(1.00),
	}
)

func validateDuration(duration string) error {
    validDurations := map[string]bool{
        "monthly":   true,
        "quarterly": true,
        "yearly":    true,
    }
    if !validDurations[duration] {
        return fmt.Errorf("invalid billing period: %s", duration)
    }
    return nil
}

func (r *PaymentRequest) Validate() error {
	if err := validate.Struct(r); err != nil {
		return fmt.Errorf("invalid payment request: %w", err)
	}

	// Additional validation for duration
	if err := validateDuration(r.Metadata.Duration); err != nil {
		return err
	}

	// Ensure at least one feature or bundle is specified
	if len(r.Metadata.Features) == 0 && len(r.Metadata.Bundles) == 0 {
		return fmt.Errorf("at least one feature or bundle must be specified")
	}

	return nil
}

func NewPaymentService(provider PaymentProvider, notifications *NotificationService) *PaymentService {
    service := &PaymentService{
        provider:      provider,
        notifications: notifications,
    }
    service.webhookHandler = NewWebhookHandler(service)
    return service
}

func (s *PaymentService) SetNotificationService(notifications *NotificationService) {
	s.notifications = notifications
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Validate amount first
	if err := req.Amount.Validate(); err != nil {
		if s.notifications != nil {
			s.notifications.NotifyPaymentFailure(ctx, req.Metadata.OrganizationID, "bundle_purchase", err.Error())
		}
		return nil, fmt.Errorf("invalid payment amount: %w", err)
	}

	// Validate the entire request using our validator
	if err := req.Validate(); err != nil {
		if s.notifications != nil {
			s.notifications.NotifyPaymentFailure(ctx, req.Metadata.OrganizationID, "bundle_purchase", err.Error())
		}
		return nil, err
	}

	payment, err := s.provider.CreatePayment(ctx, req)
	if err != nil {
		if s.notifications != nil {
			s.notifications.NotifyPaymentFailure(ctx, req.Metadata.OrganizationID, "bundle_purchase", err.Error())
		}
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) VerifyAndActivate(ctx context.Context, paymentID string) error {
    payment, err := s.provider.VerifyPayment(ctx, paymentID)
    if err != nil {
        return fmt.Errorf("failed to verify payment: %w", err)
    }

    if payment.Status != "paid" {
        if s.notifications != nil {
            s.notifications.NotifyPaymentFailure(ctx, payment.Metadata.OrganizationID, "payment_verification",
                fmt.Sprintf("Payment status: %s", payment.Status))
        }
        return fmt.Errorf("payment not completed: %s", payment.Status)
    }

    // Parse duration from metadata
    duration, err := parseDuration(payment.Metadata.Duration)
    if err != nil {
        return fmt.Errorf("invalid duration: %w", err)
    }

    // Create license activation
    licensingSvc := licensing.GetService()
    license, err := licensingSvc.ActivateLicense(
        ctx,
        payment.Metadata.OrganizationID,
        payment.Metadata.Features,
        payment.Metadata.Bundles,
        duration,
    )
    if err != nil {
        if s.notifications != nil {
            s.notifications.NotifyPaymentFailure(ctx, payment.Metadata.OrganizationID, "license_activation",
                fmt.Sprintf("License activation failed: %v", err))
        }
        return fmt.Errorf("failed to activate license: %w", err)
    }

    // Log successful activation
    fmt.Printf("Successfully activated license %s for organization %s\n", license.ID, payment.Metadata.OrganizationID)

    if s.notifications != nil {
        s.notifications.NotifyPaymentSuccess(ctx, payment.Metadata.OrganizationID, "license_activation")
    }

    return nil
}

func parseDuration(period string) (time.Duration, error) {
    switch period {
    case "monthly":
        return 30 * 24 * time.Hour, nil
    case "quarterly":
        return 90 * 24 * time.Hour, nil
    case "yearly":
        return 365 * 24 * time.Hour, nil
    default:
        return 0, fmt.Errorf("invalid billing period: %s", period)
    }
}
