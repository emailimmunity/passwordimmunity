package payment

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/mollie/mollie-api-go/v2/mollie"
    "github.com/google/uuid"
    "github.com/emailimmunity/passwordimmunity/db/models"
)

type Repository interface {
    CreatePayment(ctx context.Context, payment *models.Payment) error
    UpdatePaymentStatus(ctx context.Context, paymentID, status string) error
}

type LicensingService interface {
    ActivateLicense(ctx context.Context, orgID uuid.UUID, licenseType string, validUntil time.Time) error
}

type MollieClient interface {
    Get(paymentID string) (*mollie.Payment, error)
    Create(payment *mollie.Payment) (*mollie.Payment, error)
}

type MollieService struct {
    client     MollieClient
    repository Repository
    licensing  LicensingService
    config     *MollieConfig
}

type mollieClientWrapper struct {
    *mollie.Client
}

func NewMollieService(repository Repository, licensing LicensingService) (*MollieService, error) {
    config, err := LoadMollieConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load Mollie configuration: %w", err)
    }

    client := mollie.NewClient(config.GetActiveKey(), nil)
    wrapper := &mollieClientWrapper{client}

    return &MollieService{
        client:     wrapper,
        repository: repository,
        licensing:  licensing,
        config:     config,
    }, nil
}

func (w *mollieClientWrapper) Get(paymentID string) (*mollie.Payment, error) {
    return w.Payments.Get(paymentID)
}

func (w *mollieClientWrapper) Create(payment *mollie.Payment) (*mollie.Payment, error) {
    return w.Payments.Create(payment)
}

func NewMollieService(apiKey string, repo Repository, licensing LicensingService) (*MollieService, error) {
    client, err := mollie.NewClient(apiKey, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Mollie client: %w", err)
    }

    return &MollieService{
        client:     &mollieClientWrapper{client},
        repository: repo,
        licensing:  licensing,
    }, nil
}

type PaymentRequest struct {
    OrganizationID uuid.UUID
    LicenseType    string
    Period         string // "monthly" or "yearly"
    Description    string
    RedirectURL    string
    WebhookURL     string
}

func (s *MollieService) CreatePayment(ctx context.Context, req PaymentRequest) (*models.Payment, error) {
    // Validate license type and period
    if !isValidLicenseType(req.LicenseType) {
        return nil, fmt.Errorf("invalid license type: %s", req.LicenseType)
    }
    if !isValidPeriod(req.Period) {
        return nil, fmt.Errorf("invalid period: %s", req.Period)
    }

    // Get pricing based on license type and period
    amount, currency := s.getPricing(req.LicenseType, req.Period)

    molliePayment := &mollie.Payment{
        Amount: &mollie.Amount{
            Currency: currency,
            Value:    amount,
        },
        Description:  req.Description,
        RedirectURL:  req.RedirectURL,
        WebhookURL:   req.WebhookURL,
        Method:       mollie.PaymentMethodCreditCard,
        Metadata:     map[string]string{
            "organization_id": req.OrganizationID.String(),
            "license_type":    req.LicenseType,
            "period":         req.Period,
        },
    }

    payment, err := s.client.Create(molliePayment)
    if err != nil {
        return nil, fmt.Errorf("failed to create Mollie payment: %w", err)
    }

    // Store payment record
    paymentRecord := &models.Payment{
        OrganizationID: req.OrganizationID,
        ProviderID:     payment.ID,
        Amount:         amount,
        Currency:       currency,
        Status:         "pending",
        LicenseType:    req.LicenseType,
        Period:         req.Period,
        CreatedAt:      time.Now(),
    }

    if err := s.repository.CreatePayment(ctx, paymentRecord); err != nil {
        return nil, fmt.Errorf("failed to store payment record: %w", err)
    }

    return paymentRecord, nil
}

func (s *MollieService) HandleWebhook(ctx context.Context, paymentID string) error {
    payment, err := s.client.Get(paymentID)
    if err != nil {
        return fmt.Errorf("failed to get payment details: %w", err)
    }

    // Update payment status
    status := payment.Status
    orgID, err := uuid.Parse(payment.Metadata["organization_id"])
    if err != nil {
        return fmt.Errorf("failed to parse organization ID: %w", err)
    }

    // Validate metadata
    licenseType := payment.Metadata["license_type"]
    period := payment.Metadata["period"]

    if !isValidLicenseType(licenseType) {
        return fmt.Errorf("invalid license type in payment metadata: %s", licenseType)
    }
    if !isValidPeriod(period) {
        return fmt.Errorf("invalid period in payment metadata: %s", period)
    }

    // Update payment record with more details
    if err := s.repository.UpdatePaymentStatus(ctx, paymentID, string(status)); err != nil {
        log.Printf("Warning: Failed to update payment status: %v", err)
        // Continue processing as this is not critical
    }

    switch status {
    case mollie.PaymentStatusPaid:
        // Calculate license duration
        duration := 30 * 24 * time.Hour // monthly
        if period == "yearly" {
            duration = 365 * 24 * time.Hour
        }

        // Activate license
        if err := s.licensing.ActivateLicense(ctx, orgID, licenseType, time.Now().Add(duration)); err != nil {
            log.Printf("Critical: Failed to activate license for organization %s: %v", orgID, err)
            return fmt.Errorf("failed to activate license: %w", err)
        }

        log.Printf("Successfully activated %s license for organization %s with duration %s",
            licenseType, orgID, duration)

    case mollie.PaymentStatusFailed, mollie.PaymentStatusExpired, mollie.PaymentStatusCanceled:
        log.Printf("Payment %s status changed to %s for organization %s", paymentID, status, orgID)
        // Could implement notification system here for failed payments

    default:
        log.Printf("Payment %s status changed to %s for organization %s", paymentID, status, orgID)
    }

    return nil
}

func (s *MollieService) getPricing(licenseType, period string) (string, string) {
    if !isValidLicenseType(licenseType) || !isValidPeriod(period) {
        return "0.00", "EUR"
    }

    switch {
    case licenseType == "enterprise" && period == "yearly":
        return "999.00", "EUR"
    case licenseType == "enterprise" && period == "monthly":
        return "99.00", "EUR"
    case licenseType == "premium" && period == "yearly":
        return "499.00", "EUR"
    case licenseType == "premium" && period == "monthly":
        return "49.00", "EUR"
    default:
        return "0.00", "EUR"
    }
}

func isValidLicenseType(licenseType string) bool {
    validTypes := map[string]bool{
        "enterprise": true,
        "premium":    true,
        "free":      true,
    }
    return validTypes[licenseType]
}

func isValidPeriod(period string) bool {
    validPeriods := map[string]bool{
        "monthly": true,
        "yearly":  true,
    }
    return validPeriods[period]
}
