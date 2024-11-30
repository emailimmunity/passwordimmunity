package payment

import (
    "context"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/emailimmunity/passwordimmunity/db/repository"
    "github.com/google/uuid"
    "github.com/rs/zerolog/log"
)

type Service interface {
    CreatePayment(ctx context.Context, orgID uuid.UUID, licenseType, period string) (*models.Payment, error)
    GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error)
    HandlePaymentWebhook(ctx context.Context, providerID string, status string) error
    ActivateLicense(ctx context.Context, payment *models.Payment) error
    CancelLicense(ctx context.Context, paymentID string) error
}

type service struct {
    repo         repository.Repository
    mollieClient MollieClient
    config       *MollieConfig
    notifier     NotificationService
}

func NewService(repo repository.Repository) (Service, error) {
    config, err := LoadMollieConfig()
    if err != nil {
        return nil, err
    }

    client, err := NewMollieClient(config)
    if err != nil {
        return nil, err
    }

    // Initialize notification service with a default email service
    notifier := NewNotificationService(&defaultEmailService{})

    return &service{
        repo:         repo,
        mollieClient: client,
        config:       config,
        notifier:     notifier,
    }, nil
}

func (s *service) CreatePayment(ctx context.Context, orgID uuid.UUID, licenseType, period string) (*models.Payment, error) {
    logger := log.With().
        Str("organization_id", orgID.String()).
        Str("license_type", licenseType).
        Str("period", period).
        Logger()

    amount := calculateAmount(licenseType, period)
    logger.Info().Str("amount", amount).Msg("calculated payment amount")

    molliePayment, err := s.mollieClient.CreatePayment(amount, "EUR", licenseType)
    if err != nil {
        logger.Error().Err(err).Msg("failed to create Mollie payment")
        return nil, err
    }

    payment := &models.Payment{
        OrganizationID: orgID,
        ProviderID:     molliePayment.ID,
        Amount:         amount,
        Currency:       "EUR",
        Status:         "pending",
        LicenseType:    licenseType,
        Period:         period,
    }

    if err := s.repo.CreatePayment(ctx, payment); err != nil {
        logger.Error().Err(err).Msg("failed to create payment record")
        return nil, err
    }

    logger.Info().
        Str("provider_id", molliePayment.ID).
        Msg("payment created successfully")
    return payment, nil
}

func (s *service) GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
    return s.repo.GetPaymentByID(ctx, id)
}

func (s *service) HandlePaymentWebhook(ctx context.Context, providerID string, status string) error {
    logger := log.With().
        Str("provider_id", providerID).
        Str("status", status).
        Logger()

    payment, err := s.repo.GetPaymentByProviderID(ctx, providerID)
    if err != nil {
        logger.Error().Err(err).Msg("failed to retrieve payment")
        return err
    }

    payment.Status = status
    if status == "paid" {
        payment.PaidAt = time.Now()
    }

    if err := s.repo.UpdatePayment(ctx, payment); err != nil {
        logger.Error().Err(err).Msg("failed to update payment status")
        return err
    }

    switch status {
    case "paid":
        logger.Info().Msg("activating license for successful payment")
        return s.ActivateLicense(ctx, payment)
    case "failed", "canceled", "expired":
        logger.Info().Msg("canceling license for failed/canceled/expired payment")
        if err := s.notifier.NotifyPaymentFailed(ctx, payment, fmt.Sprintf("Payment status: %s", status)); err != nil {
            logger.Error().Err(err).Msg("failed to send payment failure notification")
        }
        return s.CancelLicense(ctx, providerID)
    }

    return nil
}

    payment, err := s.repo.GetPaymentByProviderID(ctx, providerID)
    if err != nil {
        logger.Error().Err(err).Msg("failed to get payment by provider ID")
        return err
    }

    logger = logger.With().
        Str("organization_id", payment.OrganizationID.String()).
        Str("payment_id", payment.ID.String()).
        Logger()

    payment.Status = status
    if status == "paid" {
        payment.PaidAt = time.Now()
    }

    if err := s.repo.UpdatePayment(ctx, payment); err != nil {
        logger.Error().Err(err).Msg("failed to update payment status")
        return err
    }

    switch status {
    case "paid":
        logger.Info().Msg("activating license for successful payment")
        return s.ActivateLicense(ctx, payment)
    case "failed", "canceled", "expired":
        logger.Info().Msg("canceling license for failed/canceled/expired payment")
        if err := s.notifier.NotifyPaymentFailed(ctx, payment, fmt.Sprintf("Payment status changed to %s", status)); err != nil {
            logger.Error().Err(err).Msg("failed to send payment failure notification")
        }
        return s.CancelLicense(ctx, providerID)
    }

    return nil
}

func (s *service) ActivateLicense(ctx context.Context, payment *models.Payment) error {
    logger := log.With().
        Str("organization_id", payment.OrganizationID.String()).
        Str("payment_id", payment.ID.String()).
        Str("license_type", payment.LicenseType).
        Logger()

    expiresAt := calculateExpiryDate(payment.Period)
    features := getFeaturesByLicenseType(payment.LicenseType)

    license := &models.License{
        OrganizationID: payment.OrganizationID,
        Type:           payment.LicenseType,
        Status:         "active",
        ExpiresAt:      expiresAt,
        Features:       features,
        PaymentID:      payment.ID,
    }

    if err := s.repo.CreateLicense(ctx, license); err != nil {
        logger.Error().Err(err).Msg("failed to create license")
        return err
    }

    logger.Info().
        Time("expires_at", expiresAt).
        Strs("features", features).
        Msg("license activated successfully")
    return nil
}

func (s *service) CancelLicense(ctx context.Context, paymentID string) error {
    payment, err := s.repo.GetPaymentByProviderID(ctx, paymentID)
    if err != nil {
        return err
    }

    license, err := s.repo.GetLicenseByPaymentID(ctx, payment.ID)
    if err != nil {
        return err
    }

    license.Status = "canceled"
    return s.repo.UpdateLicense(ctx, license)
}

func calculateAmount(licenseType, period string) string {
    switch {
    case licenseType == "enterprise" && period == "yearly":
        return "999.00"
    case licenseType == "enterprise" && period == "monthly":
        return "99.00"
    case licenseType == "premium" && period == "yearly":
        return "499.00"
    case licenseType == "premium" && period == "monthly":
        return "49.00"
    default:
        return "0.00"
    }
}

func calculateExpiryDate(period string) time.Time {
    now := time.Now()
    if period == "yearly" {
        return now.AddDate(1, 0, 0)
    }
    return now.AddDate(0, 1, 0)
}

func getFeaturesByLicenseType(licenseType string) []string {
    switch licenseType {
    case "enterprise":
        return []string{
            "sso",
            "directory_sync",
            "enterprise_policies",
            "advanced_reporting",
            "api_access",
            "custom_roles",
            "advanced_groups",
            "multi_tenant",
            "advanced_vault",
            "cross_org_management",
        }
    case "premium":
        return []string{
            "advanced_2fa",
            "emergency_access",
            "priority_support",
            "basic_api_access",
            "basic_reporting",
        }
    default:
        return []string{
            "basic_vault",
            "basic_sharing",
            "standard_2fa",
        }
    }
}
