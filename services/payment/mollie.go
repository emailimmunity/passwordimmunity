package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

var (
	validate = validator.New()
	minAmounts = map[string]decimal.Decimal{
		"EUR": decimal.NewFromFloat(1.00),
		"USD": decimal.NewFromFloat(1.00),
		"GBP": decimal.NewFromFloat(1.00),
	}
)

type MollieService struct {
	client     *http.Client
	apiKey     string
	webhookURL string
	baseURL    string
	logger     logger.Logger
}

type PaymentRequest struct {
	Amount      Amount `json:"amount" validate:"required"`
	Description string `json:"description" validate:"required"`
	RedirectURL string `json:"redirectUrl" validate:"required,url"`
	WebhookURL  string `json:"webhookUrl" validate:"required,url"`
	Metadata    struct {
		OrganizationID string   `json:"organizationId" validate:"required"`
		Features       []string `json:"features" validate:"omitempty,dive,required"`
		Bundles        []string `json:"bundles" validate:"omitempty,dive,required"`
		Duration       string   `json:"duration" validate:"required"`
	} `json:"metadata" validate:"required"`
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

	if value.IsNegative() {
		return fmt.Errorf("amount value cannot be negative")
	}

	if minAmount, ok := minAmounts[a.Currency]; ok && value.LessThan(minAmount) {
		return fmt.Errorf("amount must be at least %v %s", minAmount, a.Currency)
	}

	return nil
}

type PaymentResponse struct {
	ID          string    `json:"id" validate:"required"`
	Status      string    `json:"status" validate:"required,oneof=open paid expired canceled failed"`
	CreatedAt   time.Time `json:"createdAt" validate:"required"`
	Amount      Amount    `json:"amount" validate:"required"`
	Description string    `json:"description" validate:"required"`
	RedirectURL string    `json:"redirectUrl" validate:"required,url"`
	WebhookURL  string    `json:"webhookUrl" validate:"required,url"`
	Metadata    struct {
		OrganizationID string   `json:"organizationId" validate:"required"`
		Features       []string `json:"features" validate:"omitempty,dive,required"`
		Bundles        []string `json:"bundles" validate:"omitempty,dive,required"`
		Duration       string   `json:"duration" validate:"required"`
	} `json:"metadata" validate:"required"`
}

func NewMollieService(apiKey, webhookURL string, logger logger.Logger) *MollieService {
	return &MollieService{
		client:     &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		webhookURL: webhookURL,
		baseURL:    "https://api.mollie.com/v2",
		logger:     logger,
	}
}

func (s *MollieService) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		s.logger.Error("Invalid payment request", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, err
	}

	// Validate billing period
	if !config.ValidateBillingPeriod(req.Metadata.Duration) {
		s.logger.Error("Invalid billing period", "duration", req.Metadata.Duration, "organization_id", req.Metadata.OrganizationID)
		return nil, ErrInvalidBillingPeriod
	}

	// Validate enterprise features
	if err := validateEnterpriseFeatures(req.Metadata.Features); err != nil {
		s.logger.Error("Invalid enterprise features", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, err
	}

	url := fmt.Sprintf("%s/payments", s.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		s.logger.Error("Failed to marshal payment request", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		s.logger.Error("Failed to create request", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	request.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(request)
	if err != nil {
		s.logger.Error("Failed to send request", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		s.logger.Error("Unexpected status code",
			"status_code", resp.StatusCode,
			"organization_id", req.Metadata.OrganizationID,
		)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payment PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		s.logger.Error("Failed to decode response", "error", err, "organization_id", req.Metadata.OrganizationID)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	s.logger.Info("Payment created successfully",
		"payment_id", payment.ID,
		"organization_id", req.Metadata.OrganizationID,
		"amount", fmt.Sprintf("%s %s", payment.Amount.Value, payment.Amount.Currency),
	)

	return &payment, nil
}

func (s *MollieService) VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	url := fmt.Sprintf("%s/payments/%s", s.baseURL, paymentID)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.client.Do(request)
	if err != nil {
		s.logger.Error("Failed to send request", "error", err, "organization_id", payment.Metadata.OrganizationID)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Unexpected status code",
			"status_code", resp.StatusCode,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payment PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		s.logger.Error("Failed to decode response", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate payment response
	if err := payment.Validate(); err != nil {
		s.logger.Error("Invalid payment response", "error", err)
		return nil, fmt.Errorf("invalid payment response: %w", err)
	}

	return &payment, nil
}

func (s *MollieService) HandleWebhook(payment *PaymentResponse) error {
	// Validate payment response
	if payment == nil {
		s.logger.Error("Payment response is nil")
		return fmt.Errorf("payment response is nil")
	}

	if err := payment.Validate(); err != nil {
		s.logger.Error("Invalid payment", "error", err, "payment_id", payment.ID)
		return fmt.Errorf("invalid payment: %w", err)
	}

	// Verify payment status
	if payment.Status != "paid" {
		s.logger.Error("Payment not completed",
			"status", payment.Status,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("payment not completed: %s", payment.Status)
	}

	// Validate enterprise features
	if err := validateEnterpriseFeatures(payment.Metadata.Features); err != nil {
		s.logger.Error("Invalid enterprise features",
			"error", err,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("invalid enterprise features: %w", err)
	}

	// Validate feature bundles
	if err := validateFeatureBundles(payment.Metadata.Bundles); err != nil {
		s.logger.Error("Invalid feature bundles",
			"error", err,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("invalid feature bundles: %w", err)
	}

	// Verify payment amount matches expected price
	expectedPrice, err := config.GetFeaturePrice(payment.Metadata.Features, payment.Metadata.Bundles, payment.Amount.Currency)
	if err != nil {
		s.logger.Error("Failed to get expected price",
			"error", err,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("failed to get expected price: %w", err)
	}

	actualPrice, err := decimal.NewFromString(payment.Amount.Value)
	if err != nil {
		s.logger.Error("Invalid payment amount",
			"error", err,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("invalid payment amount: %w", err)
	}

	if !actualPrice.Equal(expectedPrice) {
		s.logger.Error("Payment amount mismatch",
			"actual_amount", fmt.Sprintf("%s %s", payment.Amount.Value, payment.Amount.Currency),
			"expected_amount", fmt.Sprintf("%s %s", expectedPrice.String(), payment.Amount.Currency),
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("payment amount %s %s does not match expected price %s %s",
			payment.Amount.Value, payment.Amount.Currency,
			expectedPrice.String(), payment.Amount.Currency)
	}

	// Parse duration
	duration, err := time.ParseDuration(payment.Metadata.Duration)
	if err != nil {
		s.logger.Error("Invalid duration",
			"error", err,
			"duration", payment.Metadata.Duration,
			"payment_id", payment.ID,
			"organization_id", payment.Metadata.OrganizationID,
		)
		return fmt.Errorf("invalid duration: %w", err)
	}

	// Get licensing service
	licensingSvc := licensing.GetService()

	// Activate license with features and bundles
	_, err = licensingSvc.ActivateLicense(
		context.Background(),
		payment.Metadata.OrganizationID,
		payment.Metadata.Features,
		payment.Metadata.Bundles,
		duration,
		payment.ID,
		payment.Amount.Currency,
		actualPrice.InexactFloat64(),
	)
	if err != nil {
		return fmt.Errorf("failed to activate license: %w", err)
	}

	return nil
}

// validateEnterpriseFeatures checks if the requested features are valid enterprise features
func validateEnterpriseFeatures(features []string) error {
    if len(features) == 0 {
        return nil
    }

    validFeatures := map[string]bool{
        "advanced_sso":       true,
        "directory_sync":     true,
        "advanced_reporting": true,
        "custom_roles":       true,
        "advanced_policies":  true,
        "priority_support":   true,
        "advanced_audit":     true,
        "emergency_access":   true,
        "custom_groups":      true,
        "advanced_api":       true,
    }

    for _, feature := range features {
        if !validFeatures[feature] {
            return fmt.Errorf("invalid enterprise feature: %s", feature)
        }
    }
    return nil
}

// validateFeatureBundles checks if the requested feature bundles are valid
func validateFeatureBundles(bundles []string) error {
    if len(bundles) == 0 {
        return nil
    }

    validBundles := map[string]bool{
        "security":    true,
        "enterprise": true,
        "compliance": true,
        "advanced":   true,
    }

    for _, bundle := range bundles {
        if !validBundles[bundle] {
            return fmt.Errorf("invalid feature bundle: %s", bundle)
        }
    }
    return nil
}
