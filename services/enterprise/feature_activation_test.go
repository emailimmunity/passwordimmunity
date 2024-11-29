package enterprise

import (
	"testing"
	"time"
)

type mockFeatureActivationService struct {
	features map[string]map[string]*FeatureActivation
}

func newMockFeatureActivationService() *mockFeatureActivationService {
	return &mockFeatureActivationService{
		features: make(map[string]map[string]*FeatureActivation),
	}
}

func (m *mockFeatureActivationService) ActivateFeature(featureID, organizationID string, duration time.Duration) (*FeatureActivation, error) {
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	if m.features[organizationID] == nil {
		m.features[organizationID] = make(map[string]*FeatureActivation)
	}

	if _, exists := m.features[organizationID][featureID]; exists {
		return nil, ErrFeatureAlreadyActive
	}

	activation := &FeatureActivation{
		FeatureID:      featureID,
		OrganizationID: organizationID,
		ActivatedAt:    time.Now(),
		ExpiresAt:      time.Now().Add(duration),
		Active:         true,
	}

	m.features[organizationID][featureID] = activation
	return activation, nil
}

func (m *mockFeatureActivationService) DeactivateFeature(featureID, organizationID string) error {
	if m.features[organizationID] == nil {
		return ErrFeatureNotFound
	}

	if _, exists := m.features[organizationID][featureID]; !exists {
		return ErrFeatureNotFound
	}

	delete(m.features[organizationID], featureID)
	return nil
}

func (m *mockFeatureActivationService) IsFeatureActive(featureID, organizationID string) (bool, error) {
	if m.features[organizationID] == nil {
		return false, nil
	}

	activation, exists := m.features[organizationID][featureID]
	if !exists {
		return false, nil
	}

	return activation.Active && time.Now().Before(activation.ExpiresAt), nil
}

func TestFeatureActivation(t *testing.T) {
	service := newMockFeatureActivationService()

	t.Run("activate feature", func(t *testing.T) {
		activation, err := service.ActivateFeature("test_feature", "test_org", 30*24*time.Hour)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !activation.Active {
			t.Error("feature should be active")
		}
	})

	t.Run("activate already active feature", func(t *testing.T) {
		_, err := service.ActivateFeature("test_feature", "test_org", 30*24*time.Hour)
		if err != ErrFeatureAlreadyActive {
			t.Errorf("expected ErrFeatureAlreadyActive, got %v", err)
		}
	})

	t.Run("activate with invalid duration", func(t *testing.T) {
		_, err := service.ActivateFeature("test_feature", "test_org", -time.Hour)
		if err != ErrInvalidDuration {
			t.Errorf("expected ErrInvalidDuration, got %v", err)
		}
	})

	t.Run("deactivate feature", func(t *testing.T) {
		err := service.DeactivateFeature("test_feature", "test_org")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		active, err := service.IsFeatureActive("test_feature", "test_org")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if active {
			t.Error("feature should be inactive")
		}
	})

	t.Run("get active features", func(t *testing.T) {
		// Activate two features
		_, _ = service.ActivateFeature("feature1", "test_org", 30*24*time.Hour)
		_, _ = service.ActivateFeature("feature2", "test_org", 30*24*time.Hour)

		features, err := service.GetActiveFeatures("test_org")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(features) != 2 {
			t.Errorf("expected 2 active features, got %d", len(features))
		}
	})

	t.Run("extend feature activation", func(t *testing.T) {
		featureID := "extend_test"
		orgID := "test_org"

		// Activate feature
		initial, _ := service.ActivateFeature(featureID, orgID, 24*time.Hour)
		initialExpiry := initial.ExpiresAt

		// Extend activation
		extended, err := service.ExtendFeatureActivation(featureID, orgID, 24*time.Hour)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !extended.ExpiresAt.After(initialExpiry) {
			t.Error("feature expiry should be extended")
		}
	})

	t.Run("get feature activation", func(t *testing.T) {
		featureID := "get_test"
		orgID := "test_org"

		// Activate feature
		_, _ = service.ActivateFeature(featureID, orgID, 24*time.Hour)

		// Get activation
		activation, err := service.GetFeatureActivation(featureID, orgID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if activation.FeatureID != featureID || activation.OrganizationID != orgID {
			t.Error("incorrect activation details returned")
		}
	})
}

func (m *mockFeatureActivationService) GetActiveFeatures(organizationID string) ([]FeatureActivation, error) {
	var active []FeatureActivation
	if m.features[organizationID] == nil {
		return active, nil
	}

	for _, activation := range m.features[organizationID] {
		if activation.Active && time.Now().Before(activation.ExpiresAt) {
			active = append(active, *activation)
		}
	}
	return active, nil
}

func (m *mockFeatureActivationService) ExtendFeatureActivation(featureID, organizationID string, duration time.Duration) (*FeatureActivation, error) {
	if duration <= 0 {
		return nil, ErrInvalidDuration
	}

	if m.features[organizationID] == nil {
		return nil, ErrFeatureNotFound
	}

	activation, exists := m.features[organizationID][featureID]
	if !exists {
		return nil, ErrFeatureNotFound
	}

	activation.ExpiresAt = activation.ExpiresAt.Add(duration)
	return activation, nil
}

func (m *mockFeatureActivationService) GetFeatureActivation(featureID, organizationID string) (*FeatureActivation, error) {
	if m.features[organizationID] == nil {
		return nil, ErrFeatureNotFound
	}

	activation, exists := m.features[organizationID][featureID]
	if !exists {
		return nil, ErrFeatureNotFound
	}

	return activation, nil
}

func TestFeatureActivation(t *testing.T) {
	service := newMockFeatureActivationService()

	t.Run("activate feature", func(t *testing.T) {
		activation, err := service.ActivateFeature("test_feature", "test_org", 30*24*time.Hour)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !activation.Active {
			t.Error("feature should be active")
		}
	})

	t.Run("activate already active feature", func(t *testing.T) {
		_, err := service.ActivateFeature("test_feature", "test_org", 30*24*time.Hour)
		if err != ErrFeatureAlreadyActive {
			t.Errorf("expected ErrFeatureAlreadyActive, got %v", err)
		}
	})

	t.Run("activate with invalid duration", func(t *testing.T) {
		_, err := service.ActivateFeature("test_feature", "test_org", -time.Hour)
		if err != ErrInvalidDuration {
			t.Errorf("expected ErrInvalidDuration, got %v", err)
		}
	})
}
