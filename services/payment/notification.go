package payment

import (
    "context"
    "fmt"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/rs/zerolog/log"
)

type NotificationService interface {
    NotifyPaymentFailed(ctx context.Context, payment *models.Payment, reason string) error
    NotifyLicenseCanceled(ctx context.Context, license *models.License, reason string) error
}

type notificationService struct {
    emailService EmailService
}

type EmailService interface {
    SendEmail(ctx context.Context, to string, subject string, body string) error
}

// Variable to allow test overriding
var NewNotificationService = func(emailService EmailService) NotificationService {
    return NewNotificationServiceWithEmail(emailService)
}

// NewNotificationServiceWithEmail creates a new notification service with the given email service
func NewNotificationServiceWithEmail(emailService EmailService) NotificationService {
    return &notificationService{
        emailService: emailService,
    }
}

func (s *notificationService) NotifyPaymentFailed(ctx context.Context, payment *models.Payment, reason string) error {
    logger := log.With().
        Str("organization_id", payment.OrganizationID.String()).
        Str("payment_id", payment.ID.String()).
        Str("reason", reason).
        Logger()

    subject := fmt.Sprintf("Payment Failed - %s", payment.ProviderID)
    body := fmt.Sprintf(`Payment failed for organization %s
Payment Details:
- Provider ID: %s
- Amount: %s %s
- License Type: %s
- Period: %s
- Reason: %s`,
        payment.OrganizationID,
        payment.ProviderID,
        payment.Amount,
        payment.Currency,
        payment.LicenseType,
        payment.Period,
        reason)

    if err := s.emailService.SendEmail(ctx, "admin@passwordimmunity.com", subject, body); err != nil {
        logger.Error().Err(err).Msg("failed to send payment failure notification")
        return err
    }

    logger.Info().Msg("payment failure notification sent")
    return nil
}

func (s *notificationService) NotifyLicenseCanceled(ctx context.Context, license *models.License, reason string) error {
    logger := log.With().
        Str("organization_id", license.OrganizationID.String()).
        Str("license_id", license.ID.String()).
        Str("reason", reason).
        Logger()

    subject := fmt.Sprintf("License Canceled - %s", license.ID)
    body := fmt.Sprintf(`License canceled for organization %s
License Details:
- License ID: %s
- Type: %s
- Reason: %s`,
        license.OrganizationID,
        license.ID,
        license.Type,
        reason)

    if err := s.emailService.SendEmail(ctx, "admin@passwordimmunity.com", subject, body); err != nil {
        logger.Error().Err(err).Msg("failed to send license cancellation notification")
        return err
    }

    logger.Info().Msg("license cancellation notification sent")
    return nil
}
