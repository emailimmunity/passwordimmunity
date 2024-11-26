package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MollieClient interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	CancelPayment(ctx context.Context, paymentID string) error
}

type mollieClient struct {
	config     *Config
	httpClient *http.Client
}

func NewMollieClient(config *Config) MollieClient {
	return &mollieClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}
}

type molliePaymentRequest struct {
	Amount      mollieAmount         `json:"amount"`
	Description string              `json:"description"`
	RedirectURL string              `json:"redirectUrl"`
	WebhookURL  string              `json:"webhookUrl"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
}

type mollieAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type molliePaymentResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Amount      mollieAmount           `json:"amount"`
	Description string                 `json:"description"`
	RedirectURL string                 `json:"redirectUrl"`
	WebhookURL  string                 `json:"webhookUrl"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	PaidAt      string                 `json:"paidAt,omitempty"`
	ExpiresAt   string                 `json:"expiresAt,omitempty"`
}

func (c *mollieClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	mollieReq := &molliePaymentRequest{
		Amount: mollieAmount{
			Currency: req.Currency,
			Value:    req.Amount,
		},
		Description: req.Description,
		RedirectURL: req.RedirectURL,
		WebhookURL:  req.WebhookURL,
		Metadata:    req.Metadata,
	}

	body, err := json.Marshal(mollieReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.APIEndpoint+"/payments", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var mollieResp molliePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&mollieResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertMollieResponse(&mollieResp)
}

func (c *mollieClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.APIEndpoint+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var mollieResp molliePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&mollieResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertMollieResponse(&mollieResp)
}

func (c *mollieClient) CancelPayment(ctx context.Context, paymentID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", c.config.APIEndpoint+"/payments/"+paymentID, nil)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *mollieClient) convertMollieResponse(resp *molliePaymentResponse) (*Payment, error) {
	payment := &Payment{
		ID:          resp.ID,
		Status:      PaymentStatus(resp.Status),
		Amount:      resp.Amount.Value,
		Currency:    resp.Amount.Currency,
		Description: resp.Description,
		RedirectURL: resp.RedirectURL,
		WebhookURL:  resp.WebhookURL,
		Metadata:    resp.Metadata,
	}

	createdAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created at: %w", err)
	}
	payment.CreatedAt = createdAt

	if resp.PaidAt != "" {
		paidAt, err := time.Parse(time.RFC3339, resp.PaidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse paid at: %w", err)
		}
		payment.PaidAt = &paidAt
	}

	if resp.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, resp.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires at: %w", err)
		}
		payment.ExpiresAt = &expiresAt
	}

	return payment, nil
}
