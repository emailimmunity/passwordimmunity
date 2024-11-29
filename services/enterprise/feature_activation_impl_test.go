package enterprise

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockRepository struct {
	mu       sync.RWMutex
	features map[string]map[string]*FeatureActivation
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		features: make(map[string]map[string]*FeatureActivation),
	}
}

func (m *mockRepository) Create(ctx context.Context, activation *FeatureActivation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.features[activation.OrganizationID]; !ok {
		m.features[activation.OrganizationID] = make(map[string]*FeatureActivation)
	}
	m.features[activation.OrganizationID][activation.FeatureID] = activation
	return nil
}

func (m *mockRepository) Get(ctx context.Context, featureID, orgID string) (*FeatureActivation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if org, ok := m.features[orgID]; ok {
		if feature, ok := org[featureID]; ok {
			return feature, nil
		}
	}
	return nil, ErrFeatureNotFound
}

func (m *mockRepository) Update(ctx context.Context, activation *FeatureActivation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.features[activation.OrganizationID]; !ok {
		return ErrFeatureNotFound
	}
	m.features[activation.OrganizationID][activation.FeatureID] = activation
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, featureID, orgID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if org, ok := m.features[orgID]; ok {
		delete(org, featureID)
		return nil
	}
	return ErrFeatureNotFound
}

func (m *mockRepository) GetAllActive(ctx context.Context, orgID string) ([]FeatureActivation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []FeatureActivation
	if org, ok := m.features[orgID]; ok {
		for _, feature := range org {
			if feature.Active && time.Now().Before(feature.ExpiresAt) {
				result = append(result, *feature)
			}
		}
	}
	return result, nil
}

func TestFeatureActivationServiceConcurrency(t *testing.T) {
	repo := newMockRepository()
	service := NewFeatureActivationService(repo)
	ctx := context.Background()
	const numGoroutines = 10
	var wg sync.WaitGroup

	t.Run("concurrent feature activation", func(t *testing.T) {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				featureID := fmt.Sprintf("feature_%d", id)
				_, err := service.ActivateFeature(ctx, featureID, "test_org", 24*time.Hour)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}(i)
		}
		wg.Wait()

		features, err := service.GetActiveFeatures(ctx, "test_org")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(features) != numGoroutines {
			t.Errorf("expected %d features, got %d", numGoroutines, len(features))
		}
	})

	t.Run("concurrent feature deactivation", func(t *testing.T) {
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				featureID := fmt.Sprintf("feature_%d", id)
				err := service.DeactivateFeature(ctx, featureID, "test_org")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}(i)
		}
		wg.Wait()

		features, err := service.GetActiveFeatures(ctx, "test_org")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(features) != 0 {
			t.Errorf("expected 0 features, got %d", len(features))
		}
	})

	t.Run("concurrent feature extension", func(t *testing.T) {
		// First activate features
		for i := 0; i < numGoroutines; i++ {
			featureID := fmt.Sprintf("extend_feature_%d", i)
			_, _ = service.ActivateFeature(ctx, featureID, "test_org", 24*time.Hour)
		}

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				featureID := fmt.Sprintf("extend_feature_%d", id)
				_, err := service.ExtendFeatureActivation(ctx, featureID, "test_org", 24*time.Hour)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}(i)
		}
		wg.Wait()

		features, _ := service.GetActiveFeatures(ctx, "test_org")
		for _, feature := range features {
			if time.Until(feature.ExpiresAt) <= 24*time.Hour {
				t.Error("feature was not properly extended")
			}
		}
	})
}

func TestFeatureActivationService_ErrorCases(t *testing.T) {
	repo := newMockRepository()
	service := NewFeatureActivationService(repo)
	ctx := context.Background()

	t.Run("invalid duration", func(t *testing.T) {
		_, err := service.ActivateFeature(ctx, "test_feature", "test_org", -1*time.Hour)
		if err != ErrInvalidDuration {
			t.Errorf("expected ErrInvalidDuration, got %v", err)
		}
	})

	t.Run("already active feature", func(t *testing.T) {
		// First activation
		_, err := service.ActivateFeature(ctx, "test_feature", "test_org", 24*time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Try to activate again
		_, err = service.ActivateFeature(ctx, "test_feature", "test_org", 24*time.Hour)
		if err != ErrFeatureAlreadyActive {
			t.Errorf("expected ErrFeatureAlreadyActive, got %v", err)
		}
	})

	t.Run("extend non-existent feature", func(t *testing.T) {
		_, err := service.ExtendFeatureActivation(ctx, "non_existent", "test_org", 24*time.Hour)
		if err != ErrFeatureNotFound {
			t.Errorf("expected ErrFeatureNotFound, got %v", err)
		}
	})

	t.Run("deactivate non-existent feature", func(t *testing.T) {
		err := service.DeactivateFeature(ctx, "non_existent", "test_org")
		if err != ErrFeatureNotFound {
			t.Errorf("expected ErrFeatureNotFound, got %v", err)
		}
	})

	t.Run("feature expiration", func(t *testing.T) {
		// Activate with short duration
		_, err := service.ActivateFeature(ctx, "expiring_feature", "test_org", 1*time.Millisecond)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		// Check if feature is active
		active, err := service.IsFeatureActive(ctx, "expiring_feature", "test_org")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if active {
			t.Error("feature should be inactive after expiration")
		}
	})
}
