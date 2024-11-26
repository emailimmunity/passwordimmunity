package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeInfo     NotificationType = "info"
	NotificationTypeWarning  NotificationType = "warning"
	NotificationTypeError    NotificationType = "error"
	NotificationTypeSuccess  NotificationType = "success"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, userID uuid.UUID, notificationType NotificationType, message string, metadata map[string]interface{}) error
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	DeleteNotification(ctx context.Context, notificationID uuid.UUID) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]models.Notification, error)
	SendSystemNotification(ctx context.Context, orgID uuid.UUID, notificationType NotificationType, message string) error
}

type notificationService struct {
	repo repository.Repository
}

func NewNotificationService(repo repository.Repository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID uuid.UUID, notificationType NotificationType, message string, metadata map[string]interface{}) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      string(notificationType),
		Message:   message,
		Metadata:  metadataBytes,
		Read:      false,
		CreatedAt: time.Now(),
	}

	return s.repo.CreateNotification(ctx, notification)
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	notification, err := s.repo.GetNotification(ctx, notificationID)
	if err != nil {
		return err
	}

	notification.Read = true
	notification.UpdatedAt = time.Now()

	return s.repo.UpdateNotification(ctx, notification)
}

func (s *notificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	return s.repo.DeleteNotification(ctx, notificationID)
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]models.Notification, error) {
	return s.repo.GetUserNotifications(ctx, userID, unreadOnly)
}

func (s *notificationService) SendSystemNotification(ctx context.Context, orgID uuid.UUID, notificationType NotificationType, message string) error {
	// Get all users in the organization
	users, err := s.repo.GetOrganizationMembers(ctx, orgID)
	if err != nil {
		return err
	}

	metadata := map[string]interface{}{
		"system_notification": true,
		"organization_id":    orgID.String(),
	}

	// Create notification for each user
	for _, user := range users {
		if err := s.CreateNotification(ctx, user.ID, notificationType, message, metadata); err != nil {
			// Log error but continue with other users
			continue
		}
	}

	// Create audit log
	auditMetadata := createBasicMetadata("system_notification_sent", "System notification sent")
	auditMetadata["notification_type"] = string(notificationType)
	if err := s.createAuditLog(ctx, "notification.system.sent", uuid.Nil, orgID, auditMetadata); err != nil {
		return err
	}

	return nil
}
