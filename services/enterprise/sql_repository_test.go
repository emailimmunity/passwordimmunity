package enterprise

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/passwordimmunity_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Clean up existing test data
	_, err = db.Exec("DELETE FROM feature_activations")
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	return db
}

func TestSQLFeatureActivationRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLFeatureActivationRepository(db)
	ctx := context.Background()

	t.Run("create and get feature", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "test_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now().UTC().Truncate(time.Microsecond),
			ExpiresAt:      time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond),
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
		if !retrieved.ActivatedAt.Equal(activation.ActivatedAt) {
			t.Errorf("expected activated at %v, got %v", activation.ActivatedAt, retrieved.ActivatedAt)
		}
	})

	t.Run("update feature", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "update_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now().UTC().Truncate(time.Microsecond),
			ExpiresAt:      time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond),
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
			t.Errorf("expected expires at %v, got %v", activation.ExpiresAt, retrieved.ExpiresAt)
		}
	})

	t.Run("get all active features", func(t *testing.T) {
		// Create multiple features
		for i := 0; i < 3; i++ {
			activation := &FeatureActivation{
				FeatureID:      "active_feature",
				OrganizationID: "test_org",
				ActivatedAt:    time.Now().UTC().Truncate(time.Microsecond),
				ExpiresAt:      time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond),
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

	t.Run("expired features not returned", func(t *testing.T) {
		activation := &FeatureActivation{
			FeatureID:      "expired_feature",
			OrganizationID: "test_org",
			ActivatedAt:    time.Now().UTC().Truncate(time.Microsecond),
			ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond),
			Active:         true,
		}

		err := repo.Create(ctx, activation)
		if err != nil {
			t.Fatalf("failed to create feature activation: %v", err)
		}


		features, err := repo.GetAllActive(ctx, "test_org")
		if err != nil {
			t.Fatalf("failed to get active features: %v", err)
		}

		for _, f := range features {
			if f.FeatureID == "expired_feature" {
				t.Error("expired feature should not be returned in active features")
			}
		}
	})
}
