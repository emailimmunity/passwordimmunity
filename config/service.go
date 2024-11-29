package config

import (
	"context"
	"sync"
)

// FeatureActivationService defines the interface for feature activation
type FeatureActivationService interface {
	IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error)
}

var (
	enterpriseService FeatureActivationService
	serviceMu         sync.RWMutex
)

// SetEnterpriseService sets the enterprise service instance
func SetEnterpriseService(service FeatureActivationService) {
	serviceMu.Lock()
	defer serviceMu.Unlock()
	enterpriseService = service
}

// GetEnterpriseService returns the current enterprise service instance
func GetEnterpriseService() FeatureActivationService {
	serviceMu.RLock()
	defer serviceMu.RUnlock()
	return enterpriseService
}
