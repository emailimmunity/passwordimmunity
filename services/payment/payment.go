package payment

import (
	"context"
	"fmt"
	"time"
)

type service struct {
	client MollieClient
	config *Config
}

func NewService(config *Config) PaymentService {
	return &service{
		client: NewMollieClient(config),
		config: config,
	}
}

func (s *service) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	if req.Currency == "" {
		req.Currency = s.config.Currency
	}

	if req.WebhookURL == "" {
		req.WebhookURL = fmt.Sprintf("%s%s", s.config.APIEndpoint, s.config.WebhookPath)
	}

	var payment *Payment
	var err error

	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		payment, err = s.client.CreatePayment(ctx, req)
		if err == nil {
			break
		}

		if attempt == s.config.RetryAttempts {
			return nil, fmt.Errorf("failed to create payment after %d attempts: %w", s.config.RetryAttempts, err)
		}

		time.Sleep(s.config.RetryDelay)
	}

	return payment, nil
}

func (s *service) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	var payment *Payment
	var err error

	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		payment, err = s.client.GetPayment(ctx, paymentID)
		if err == nil {
			break
		}

		if attempt == s.config.RetryAttempts {
			return nil, fmt.Errorf("failed to get payment after %d attempts: %w", s.config.RetryAttempts, err)
		}

		time.Sleep(s.config.RetryDelay)
	}

	return payment, nil
}

func (s *service) CancelPayment(ctx context.Context, paymentID string) error {
	var err error

	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		err = s.client.CancelPayment(ctx, paymentID)
		if err == nil {
			break
		}

		if attempt == s.config.RetryAttempts {
			return fmt.Errorf("failed to cancel payment after %d attempts: %w", s.config.RetryAttempts, err)
		}

		time.Sleep(s.config.RetryDelay)
	}

	return nil
}

func (s *service) HandleWebhook(ctx context.Context, req *WebhookRequest) error {
	payment, err := s.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment for webhook: %w", err)
	}

	if payment.Status != req.Status {
		return fmt.Errorf("payment status mismatch: expected %s, got %s", req.Status, payment.Status)
	}

	// TODO: Implement license activation/deactivation based on payment status
	switch payment.Status {
	case PaymentStatusPaid:
		// Activate license
	case PaymentStatusFailed, PaymentStatusCancelled, PaymentStatusExpired:
		// Deactivate license
	}

	return nil
}
