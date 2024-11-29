package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/api/middleware"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type mockFeatureService struct {
	activeFeatures map[string]bool
}

func (m *mockFeatureService) IsFeatureActive(ctx context.Context, featureID, organizationID string) (bool, error) {
	key := featureID + ":" + organizationID
	return m.activeFeatures[key], nil
}

func TestEnterpriseRoutes(t *testing.T) {
	// Setup mock services
	mockService := &mockFeatureService{
		activeFeatures: map[string]bool{
			"advanced_sso:test_org":      true,
			"custom_roles:test_org":      true,
			"advanced_reporting:test_org": false,
			"multi_tenant:test_org":      false,
		},
	}
	config.SetEnterpriseService(mockService)

	// Create router with enterprise routes
	r := mux.NewRouter()
	featureManager := &featureflag.FeatureManager{}
	paymentService := &payment.Service{}
	RegisterEnterpriseRoutes(r, featureManager, paymentService)

	tests := []struct {
		name           string
		path           string
		method         string
		organizationID string
		wantStatus     int
	}{
		{
			name:           "access active SSO feature",
			path:           "/api/enterprise/sso/config",
			method:         http.MethodGet,
			organizationID: "test_org",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "access inactive reporting feature",
			path:           "/api/enterprise/reports/audit",
			method:         http.MethodGet,
			organizationID: "test_org",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "access active custom roles feature",
			path:           "/api/enterprise/roles",
			method:         http.MethodGet,
			organizationID: "test_org",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "access inactive multi-tenant feature",
			path:           "/api/enterprise/tenants",
			method:         http.MethodGet,
			organizationID: "test_org",
			wantStatus:     http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.organizationID))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("%s = status %v, want %v", tt.path, rr.Code, tt.wantStatus)
			}
		})
	}
}
