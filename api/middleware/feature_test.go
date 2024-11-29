package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emailimmunity/passwordimmunity/config"
)

type mockFeatureService struct {
	activeFeatures map[string]bool
}

func (m *mockFeatureService) IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error) {
	key := featureID + ":" + organizationID
	return m.activeFeatures[key], nil
}

func TestRequireFeature(t *testing.T) {
	// Setup mock service
	mockService := &mockFeatureService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":    true,
			"multi_tenant:test_org":    false,
		},
	}
	config.SetEnterpriseService(mockService)

	tests := []struct {
		name           string
		featureID      string
		organizationID string
		wantStatus     int
	}{
		{
			name:           "active feature",
			featureID:      "advanced_sso",
			organizationID: "test_org",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "inactive feature",
			featureID:      "multi_tenant",
			organizationID: "test_org",
			wantStatus:     http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireFeature(tt.featureID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.organizationID))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RequireFeature() status = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireBundle(t *testing.T) {
	// Setup mock service with bundle features
	mockService := &mockFeatureService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":    true,
			"advanced_policy:test_org": true,  // Makes security bundle active
			"multi_tenant:test_org":    false, // Makes management bundle inactive
		},
	}
	config.SetEnterpriseService(mockService)

	tests := []struct {
		name           string
		bundleID       string
		organizationID string
		wantStatus     int
	}{
		{
			name:           "active bundle",
			bundleID:       "security",
			organizationID: "test_org",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "inactive bundle",
			bundleID:       "management",
			organizationID: "test_org",
			wantStatus:     http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireBundle(tt.bundleID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.organizationID))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RequireBundle() status = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
