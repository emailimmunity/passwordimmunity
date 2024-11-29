package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

func TestHandleFeatureStatus(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		featureID  string
		setupLicense bool
		wantActive bool
		wantGracePeriod bool
	}{
		{
			name:      "active feature",
			orgID:     "org1",
			featureID: "advanced_sso",
			setupLicense: true,
			wantActive: true,
			wantGracePeriod: false,
		},
		{
			name:      "inactive feature",
			orgID:     "org2",
			featureID: "advanced_sso",
			setupLicense: false,
			wantActive: false,
			wantGracePeriod: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test license if needed
			if tt.setupLicense {
				svc := licensing.GetService()
				_, _ = svc.ActivateLicense(context.Background(), tt.orgID,
					[]string{tt.featureID},
					[]string{},
					30*24*60*60*1000000000)
			}

			// Create request
			req := httptest.NewRequest("GET", "/api/enterprise/features/status/"+tt.featureID, nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))

			// Setup router with vars
			vars := map[string]string{
				"feature_id": tt.featureID,
			}
			req = mux.SetURLVars(req, vars)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			HandleFeatureStatus(rr, req)

			// Check status code
			if rr.Code != http.StatusOK {
				t.Errorf("HandleFeatureStatus() status = %v, want %v", rr.Code, http.StatusOK)
			}

			// Parse response
			var response FeatureStatusResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check response fields
			if response.Active != tt.wantActive {
				t.Errorf("HandleFeatureStatus() active = %v, want %v", response.Active, tt.wantActive)
			}
			if response.InGracePeriod != tt.wantGracePeriod {
				t.Errorf("HandleFeatureStatus() inGracePeriod = %v, want %v", response.InGracePeriod, tt.wantGracePeriod)
			}
		})
	}
}

func TestHandleListBundles(t *testing.T) {
	// Create request
	req := httptest.NewRequest("GET", "/api/enterprise/features/bundles", nil)
	rr := httptest.NewRecorder()

	// Handle request
	HandleListBundles(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("HandleListBundles() status = %v, want %v", rr.Code, http.StatusOK)
	}

	// Parse response
	var response []BundleResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response contains bundles
	if len(response) == 0 {
		t.Error("Expected non-empty bundle list")
	}
}

func TestHandleListAvailableFeatures(t *testing.T) {
	// Setup test organization with some features
	svc := licensing.GetService()
	orgID := "test_org"
	_, _ = svc.ActivateLicense(context.Background(), orgID,
		[]string{"advanced_sso"},
		[]string{"security"},
		30*24*60*60*1000000000)

	// Create request
	req := httptest.NewRequest("GET", "/api/enterprise/features/available", nil)
	req = req.WithContext(context.WithValue(req.Context(), "organization_id", orgID))
	rr := httptest.NewRecorder()

	// Handle request
	HandleListAvailableFeatures(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("HandleListAvailableFeatures() status = %v, want %v", rr.Code, http.StatusOK)
	}

	// Parse response
	var response []FeatureStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response contains features
	if len(response) == 0 {
		t.Error("Expected non-empty feature list")
	}

	// Check that known active feature is marked as active
	found := false
	for _, feature := range response {
		if feature.FeatureID == "advanced_sso" {
			found = true
			if !feature.Active {
				t.Error("Expected advanced_sso feature to be active")
			}
		}
	}
	if !found {
		t.Error("Expected to find advanced_sso feature in response")
	}
}
