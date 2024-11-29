package payment

import (
	"context"
	"errors"
)

var (
	ErrInvalidAmount      = errors.New("invalid payment amount")
	ErrPaymentFailed      = errors.New("payment processing failed")
	ErrVerificationFailed = errors.New("payment verification failed")
	ErrFeatureActivation  = errors.New("feature activation failed")
)

// PaymentProvider defines the interface for payment processing
type PaymentProvider interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	VerifyPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	HandleWebhook(payment *PaymentResponse) error
}

// PaymentService manages payment processing and feature activation
type PaymentService struct {
	provider PaymentProvider
}

func NewPaymentService(provider PaymentProvider) *PaymentService {
	return &PaymentService{
		provider: provider,
	}
}

// ProcessPayment handles the complete payment flow
func (s *PaymentService) ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if err := validatePaymentRequest(req); err != nil {
		return nil, err
	}

	// Create payment
	payment, err := s.provider.CreatePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// VerifyAndActivate verifies payment and activates features
func (s *PaymentService) VerifyAndActivate(ctx context.Context, paymentID string) error {
	// Verify payment
	payment, err := s.provider.VerifyPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	// Handle webhook (activate features)
	if err := s.provider.HandleWebhook(payment); err != nil {
		return err
	}

	return nil
}

func validatePaymentRequest(req PaymentRequest) error {
	if req.Amount.Value == "" || req.Amount.Currency == "" {
		return ErrInvalidAmount
	}
	return nil
}
