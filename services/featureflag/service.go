package featureflag

import (
	"context"
	"sync"
	"time"

	"github.com/emailimmunity/passwordimmunity/config"
)

type FeatureManager struct {
	features map[string]map[string]time.Time // orgID -> feature -> expiration
	mu       sync.RWMutex
}

func NewFeatureManager() *FeatureManager {
	return &FeatureManager{
		features: make(map[string]map[string]time.Time),
	}
}

func (fm *FeatureManager) ActivateFeature(orgID, feature string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.features[orgID]; !exists {
		fm.features[orgID] = make(map[string]time.Time)
	}

	// Set to far future for permanent activation
	fm.features[orgID][feature] = time.Now().Add(100 * 365 * 24 * time.Hour)
	return nil
}

func (fm *FeatureManager) ActivateFeatureWithGracePeriod(orgID, feature string, config config.FeatureConfig) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.features[orgID]; !exists {
		fm.features[orgID] = make(map[string]time.Time)
	}

	fm.features[orgID][feature] = time.Now().Add(time.Duration(config.GracePeriod) * 24 * time.Hour)
	return nil
}

func (fm *FeatureManager) DeactivateFeature(orgID, feature string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if features, exists := fm.features[orgID]; exists {
		delete(features, feature)
	}
	return nil
}

func (fm *FeatureManager) IsFeatureActive(ctx context.Context, orgID, feature string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if features, exists := fm.features[orgID]; exists {
		if expiration, hasFeature := features[feature]; hasFeature {
			return time.Now().Before(expiration)
		}
	}
	return false
}

func (fm *FeatureManager) ListActiveFeatures(ctx context.Context, orgID string) []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var activeFeatures []string
	now := time.Now()

	if features, exists := fm.features[orgID]; exists {
		for feature, expiration := range features {
			if now.Before(expiration) {
				activeFeatures = append(activeFeatures, feature)
			}
		}
	}

	return activeFeatures
}

func (fm *FeatureManager) GetFeatureExpiration(orgID, feature string) (time.Time, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if features, exists := fm.features[orgID]; exists {
		if expiration, hasFeature := features[feature]; hasFeature {
			return expiration, true
		}
	}
	return time.Time{}, false
}
