package repository

import (
    "context"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreatePayment(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewRepository(db)
    ctx := context.Background()

    tests := []struct {
        name          string
        payment       *models.Payment
        expectedError bool
    }{
        {
            name: "valid payment",
            payment: &models.Payment{
                OrganizationID: uuid.New(),
                ProviderID:     "tr_test_123",
                Amount:         "999.00",
                Currency:       "EUR",
                Status:         "pending",
                LicenseType:    "enterprise",
                Period:         "yearly",
                CreatedAt:      time.Now(),
            },
            expectedError: false,
        },
        {
            name: "duplicate provider ID",
            payment: &models.Payment{
                OrganizationID: uuid.New(),
                ProviderID:     "tr_test_123", // Same as above
                Amount:         "49.00",
                Currency:       "EUR",
                Status:         "pending",
                LicenseType:    "premium",
                Period:         "monthly",
                CreatedAt:      time.Now(),
            },
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.CreatePayment(ctx, tt.payment)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)

                // Verify payment was stored
                var stored models.Payment
                err = db.QueryRowContext(ctx,
                    `SELECT organization_id, provider_id, amount, currency, status, license_type, period
                     FROM payments WHERE provider_id = $1`,
                    tt.payment.ProviderID,
                ).Scan(
                    &stored.OrganizationID,
                    &stored.ProviderID,
                    &stored.Amount,
                    &stored.Currency,
                    &stored.Status,
                    &stored.LicenseType,
                    &stored.Period,
                )
                require.NoError(t, err)
                assert.Equal(t, tt.payment.OrganizationID, stored.OrganizationID)
                assert.Equal(t, tt.payment.Amount, stored.Amount)
                assert.Equal(t, tt.payment.Status, stored.Status)
            }
        })
    }
}

func TestUpdatePaymentStatus(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewRepository(db)
    ctx := context.Background()

    // Create initial payment
    payment := &models.Payment{
        OrganizationID: uuid.New(),
        ProviderID:     "tr_test_456",
        Amount:         "99.00",
        Currency:       "EUR",
        Status:         "pending",
        LicenseType:    "premium",
        Period:         "monthly",
        CreatedAt:      time.Now(),
    }
    require.NoError(t, repo.CreatePayment(ctx, payment))

    tests := []struct {
        name          string
        paymentID     string
        status        string
        expectedError bool
    }{
        {
            name:          "update to paid",
            paymentID:     "tr_test_456",
            status:        "paid",
            expectedError: false,
        },
        {
            name:          "non-existent payment",
            paymentID:     "tr_test_999",
            status:        "failed",
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.UpdatePaymentStatus(ctx, tt.paymentID, tt.status)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)

                // Verify status was updated
                var storedStatus string
                err = db.QueryRowContext(ctx,
                    "SELECT status FROM payments WHERE provider_id = $1",
                    tt.paymentID,
                ).Scan(&storedStatus)
                require.NoError(t, err)
                assert.Equal(t, tt.status, storedStatus)
            }
        })
    }
}
