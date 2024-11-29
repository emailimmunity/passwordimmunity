package enterprise

import (
	"context"
	"testing"
	"time"
)

func TestSQLFeatureActivationRepository(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	t.Run("create and get feature", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "test_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			Active:         true,
		}

		err := repo.Create(ctx, activation)
		if err != nil {
			t.Fatalf("failed to create feature activation: %v", err)
		}

		retrieved, err := repo.Get(ctx, activation.FeatureID, activation.OrganizationID)
		if err != nil {
			t.Fatalf("failed to get feature activation: %v", err)
		}

		if retrieved.FeatureID != activation.FeatureID {
			t.Errorf("expected feature ID %s, got %s", activation.FeatureID, retrieved.FeatureID)
		}
	})

	t.Run("update feature", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "update_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			Active:         true,
		}

		err := repo.Create(ctx, activation)
		if err != nil {
			t.Fatalf("failed to create feature activation: %v", err)
		}

		activation.ExpiresAt = activation.ExpiresAt.Add(24 * time.Hour)
		err = repo.Update(ctx, activation)
		if err != nil {
			t.Fatalf("failed to update feature activation: %v", err)
		}

		retrieved, err := repo.Get(ctx, activation.FeatureID, activation.OrganizationID)
		if err != nil {
			t.Fatalf("failed to get feature activation: %v", err)
		}

		if !retrieved.ExpiresAt.Equal(activation.ExpiresAt) {
			t.Errorf("expiration time not updated correctly")
		}
	})

	t.Run("delete feature", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "delete_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			Active:         true,
		}

		err := repo.Create(ctx, activation)
		if err != nil {
			t.Fatalf("failed to create feature activation: %v", err)
		}

		err = repo.Delete(ctx, activation.FeatureID, activation.OrganizationID)
		if err != nil {
			t.Fatalf("failed to delete feature activation: %v", err)
		}

		_, err = repo.Get(ctx, activation.FeatureID, activation.OrganizationID)
		if err != ErrFeatureNotFound {
			t.Errorf("expected ErrFeatureNotFound, got %v", err)
		}
	})

	t.Run("get all active features", func(t *testing.T) {
		// Create multiple features
		for i := 0; i < 3; i++ {
			activation := &FeatureActivation{
				FeatureID:      "active_feature",
				OrganizationID: "test_org",
				ActivatedAt:    time.Now(),
				ExpiresAt:      time.Now().Add(24 * time.Hour),
				Active:         true,
			}
			err := repo.Create(ctx, activation)
			if err != nil {
				t.Fatalf("failed to create feature activation: %v", err)
			}
		}

		features, err := repo.GetAllActive(ctx, "test_org")
		if err != nil {
			t.Fatalf("failed to get active features: %v", err)
		}

		if len(features) != 3 {
			t.Errorf("expected 3 active features, got %d", len(features))
		}
	})
}
