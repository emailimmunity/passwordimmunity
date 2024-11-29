package repository

import (
    "context"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/stretchr/testify/assert"
)

func TestFeatureActivationRepository(t *testing.T) {
    db := setupTestDB(t)
    repo := NewFeatureActivationRepository(db)
    ctx := context.Background()

    featureID := "advanced_sso"
    orgID := "test-org-1"

    t.Run("Create and Get Feature Activation", func(t *testing.T) {
        activation := &models.FeatureActivation{
            OrganizationID: orgID,
            FeatureID:      &featureID,
            Status:         "active",
            ExpiresAt:      time.Now().Add(24 * time.Hour),
            PaymentID:      "test-payment-1",
            Currency:       "EUR",
            Amount:         99.99,
        }

        err := repo.Create(ctx, activation)
        assert.NoError(t, err)
        assert.NotZero(t, activation.ID)

        retrieved, err := repo.GetByID(ctx, activation.ID)
        assert.NoError(t, err)
        assert.Equal(t, activation.OrganizationID, retrieved.OrganizationID)
        assert.Equal(t, activation.FeatureID, retrieved.FeatureID)
        assert.Equal(t, activation.Status, retrieved.Status)
    })

    t.Run("Get Active By Organization", func(t *testing.T) {
        activations, err := repo.GetActiveByOrganization(ctx, orgID)
        assert.NoError(t, err)
        assert.NotEmpty(t, activations)
        for _, activation := range activations {
            assert.Equal(t, "active", activation.Status)
            assert.True(t, activation.ExpiresAt.After(time.Now()))
        }
    })

    t.Run("Update Status", func(t *testing.T) {
        activation := &models.FeatureActivation{
            OrganizationID: orgID,
            FeatureID:      &featureID,
            Status:         "active",
            ExpiresAt:      time.Now().Add(24 * time.Hour),
            PaymentID:      "test-payment-2",
            Currency:       "EUR",
            Amount:         99.99,
        }

        err := repo.Create(ctx, activation)
        assert.NoError(t, err)

        err = repo.UpdateStatus(ctx, activation.ID, "cancelled")
        assert.NoError(t, err)

        retrieved, err := repo.GetByID(ctx, activation.ID)
        assert.NoError(t, err)
        assert.Equal(t, "cancelled", retrieved.Status)
    })

    t.Run("Get By Feature", func(t *testing.T) {
        activation, err := repo.GetByFeature(ctx, orgID, featureID)
        assert.NoError(t, err)
        if assert.NotNil(t, activation) {
            assert.Equal(t, orgID, activation.OrganizationID)
            assert.Equal(t, &featureID, activation.FeatureID)
            assert.Equal(t, "active", activation.Status)
        }
    })

    t.Run("Get By Bundle", func(t *testing.T) {
        bundleID := "enterprise"
        activation := &models.FeatureActivation{
            OrganizationID: orgID,
            BundleID:       &bundleID,
            Status:         "active",
            ExpiresAt:      time.Now().Add(24 * time.Hour),
            PaymentID:      "test-payment-3",
            Currency:       "EUR",
            Amount:         299.99,
        }

        err := repo.Create(ctx, activation)
        assert.NoError(t, err)

        retrieved, err := repo.GetByBundle(ctx, orgID, bundleID)
        assert.NoError(t, err)
        if assert.NotNil(t, retrieved) {
            assert.Equal(t, orgID, retrieved.OrganizationID)
            assert.Equal(t, &bundleID, retrieved.BundleID)
            assert.Equal(t, "active", retrieved.Status)
        }
    })
}
