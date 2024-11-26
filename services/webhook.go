package services

import (
	"context"
	"time"
	"encoding/json"
	"net/http"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type WebhookService interface {
	RegisterWebhook(ctx context.Context, webhook models.Webhook) error
	SendWebhook(ctx context.Context, notification models.Notification) error
	GetWebhook(ctx context.Context, webhookID uuid.UUID) (*models.Webhook, error)
	ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]models.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID uuid.UUID) error
	ValidateWebhook(ctx context.Context, webhook models.Webhook) error
}

type webhookService struct {
	repo        repository.Repository
	audit       AuditService
	httpClient  *http.Client
}

func NewWebhookService(
	repo repository.Repository,
	audit AuditService,
) WebhookService {
	return &webhookService{
		repo:       repo,
		audit:      audit,
		httpClient: &http.Client{Timeout: time.Second * 10},
	}
}

func (s *webhookService) RegisterWebhook(ctx context.Context, webhook models.Webhook) error {
	webhook.ID = uuid.New()
	webhook.Status = "active"
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	if err := s.ValidateWebhook(ctx, webhook); err != nil {
		return err
	}

	if err := s.repo.CreateWebhook(ctx, &webhook); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("webhook_registered", "Webhook registered")
	metadata["url"] = webhook.URL
	metadata["events"] = webhook.Events
	if err := s.createAuditLog(ctx, "webhook.registered", uuid.Nil, webhook.OrganizationID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *webhookService) SendWebhook(ctx context.Context, notification models.Notification) error {
	webhooks, err := s.repo.GetWebhooksForEvent(ctx, notification.Type)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		delivery := &models.WebhookDelivery{
			ID:           uuid.New(),
			WebhookID:   webhook.ID,
			Payload:     notification,
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.repo.CreateWebhookDelivery(ctx, delivery); err != nil {
			continue
		}

		// Send webhook asynchronously
		go s.deliverWebhook(context.Background(), webhook, delivery)
	}

	return nil
}

func (s *webhookService) GetWebhook(ctx context.Context, webhookID uuid.UUID) (*models.Webhook, error) {
	return s.repo.GetWebhook(ctx, webhookID)
}

func (s *webhookService) ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]models.Webhook, error) {
	return s.repo.ListWebhooks(ctx, orgID)
}

func (s *webhookService) DeleteWebhook(ctx context.Context, webhookID uuid.UUID) error {
	webhook, err := s.GetWebhook(ctx, webhookID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteWebhook(ctx, webhookID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("webhook_deleted", "Webhook deleted")
	metadata["url"] = webhook.URL
	if err := s.createAuditLog(ctx, "webhook.deleted", uuid.Nil, webhook.OrganizationID, metadata); err != nil {
		return err
	}

	return nil
}

func (s *webhookService) ValidateWebhook(ctx context.Context, webhook models.Webhook) error {
	// Validate URL format and accessibility
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return errors.New("webhook endpoint returned server error")
	}

	return nil
}

func (s *webhookService) deliverWebhook(ctx context.Context, webhook models.Webhook, delivery *models.WebhookDelivery) {
	payload, err := json.Marshal(delivery.Payload)
	if err != nil {
		s.updateDeliveryStatus(ctx, delivery, "failed", err.Error())
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		s.updateDeliveryStatus(ctx, delivery, "failed", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-ID", webhook.ID.String())
	req.Header.Set("X-Delivery-ID", delivery.ID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.updateDeliveryStatus(ctx, delivery, "failed", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.updateDeliveryStatus(ctx, delivery, "delivered", "")
	} else {
		s.updateDeliveryStatus(ctx, delivery, "failed", fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
}

func (s *webhookService) updateDeliveryStatus(ctx context.Context, delivery *models.WebhookDelivery, status string, error string) {
	delivery.Status = status
	delivery.Error = error
	delivery.UpdatedAt = time.Now()
	if status == "delivered" {
		delivery.DeliveredAt = &time.Time{}
		*delivery.DeliveredAt = time.Now()
	}

	if err := s.repo.UpdateWebhookDelivery(ctx, delivery); err != nil {
		log.Printf("Failed to update webhook delivery status: %v", err)
	}
}
