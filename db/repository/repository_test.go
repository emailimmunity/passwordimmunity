package repository

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgresql://localhost/passwordimmunity_test?sslmode=disable")
    require.NoError(t, err)
    return db
}

func TestPaymentRepository(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := NewPaymentRepository(db)
    ctx := context.Background()

    t.Run("CreateAndGetPayment", func(t *testing.T) {
        payment := &Payment{
            ID:            uuid.New(),
            OrganizationID: uuid.New(),
            Amount:        99.99,
            Currency:      "EUR",
            Status:        "pending",
            PaymentMethod: "ideal",
            MollieID:      "tr_test_123",
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }

        err := repo.CreatePayment(ctx, payment)
        require.NoError(t, err)

        retrieved, err := repo.GetPaymentByMollieID(ctx, payment.MollieID)
        require.NoError(t, err)
        assert.Equal(t, payment.ID, retrieved.ID)
        assert.Equal(t, payment.Amount, retrieved.Amount)
    })

    t.Run("UpdatePaymentStatus", func(t *testing.T) {
        mollieID := "tr_test_456"
        payment := &Payment{
            ID:            uuid.New(),
            OrganizationID: uuid.New(),
            Amount:        49.99,
            Currency:      "EUR",
            Status:        "pending",
            PaymentMethod: "creditcard",
            MollieID:      mollieID,
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }

        err := repo.CreatePayment(ctx, payment)
        require.NoError(t, err)

        err = repo.UpdatePaymentStatus(ctx, mollieID, "paid")
        require.NoError(t, err)

        updated, err := repo.GetPaymentByMollieID(ctx, mollieID)
        require.NoError(t, err)
        assert.Equal(t, "paid", updated.Status)
    })
}

func TestLicenseRepository(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := NewLicenseRepository(db)
    ctx := context.Background()

    t.Run("CreateAndGetLicense", func(t *testing.T) {
        orgID := uuid.New()
        license := &License{
            ID:            uuid.New(),
            OrganizationID: orgID,
            Type:          "premium",
            Status:        "active",
            ValidUntil:    time.Now().Add(365 * 24 * time.Hour),
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }

        err := repo.CreateLicense(ctx, license)
        require.NoError(t, err)

        retrieved, err := repo.GetActiveLicenseByOrgID(ctx, orgID)
        require.NoError(t, err)
        assert.Equal(t, license.ID, retrieved.ID)
        assert.Equal(t, license.Type, retrieved.Type)
    })

    t.Run("UpdateLicenseStatus", func(t *testing.T) {
        license := &License{
            ID:            uuid.New(),
            OrganizationID: uuid.New(),
            Type:          "enterprise",
            Status:        "active",
            ValidUntil:    time.Now().Add(365 * 24 * time.Hour),
            CreatedAt:     time.Now(),
            UpdatedAt:     time.Now(),
        }

        err := repo.CreateLicense(ctx, license)
        require.NoError(t, err)

        err = repo.UpdateLicenseStatus(ctx, license.ID, "cancelled")
        require.NoError(t, err)

        retrieved, err := repo.GetActiveLicenseByOrgID(ctx, license.OrganizationID)
        require.NoError(t, err)
        assert.Nil(t, retrieved) // Should be nil since status is not active
    })
}
