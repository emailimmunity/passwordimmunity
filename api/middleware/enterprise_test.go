package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnterpriseFeatureMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		featureID  string
		orgID      string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "valid enterprise feature access",
			featureID:  "advanced_sso",
			orgID:      "org_with_license",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "unauthorized feature access",
			featureID:  "advanced_sso",
			orgID:      "org_without_license",
			wantStatus: http.StatusForbidden,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request with organization context
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create test handler
			handler := EnterpriseFeatureMiddleware(tt.featureID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Serve request
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			// Check error response format if error expected
			if tt.wantError {
				var errorResp ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&errorResp); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if errorResp.Error == "" {
					t.Error("expected error message in response")
				}
				if errorResp.UpgradeURL == "" {
					t.Error("expected upgrade URL in error response")
				}
			}
		})
	}
}

func TestCheckGracePeriod(t *testing.T) {
	tests := []struct {
		name              string
		featureID         string
		orgID             string
		expectGraceHeader bool
	}{
		{
			name:              "feature in grace period",
			featureID:         "advanced_sso",
			orgID:             "org_in_grace",
			expectGraceHeader: true,
		},
		{
			name:              "feature not in grace period",
			featureID:         "advanced_sso",
			orgID:             "org_active",
			expectGraceHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))
			rr := httptest.NewRecorder()

			handler := CheckGracePeriod(tt.featureID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if tt.expectGraceHeader {
				if rr.Header().Get("X-Grace-Period-Warning") == "" {
					t.Error("expected grace period warning header")
				}
			} else {
				if rr.Header().Get("X-Grace-Period-Warning") != "" {
					t.Error("unexpected grace period warning header")
				}
			}
		})
	}
}

func TestRequireEnterpriseLicense(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "valid enterprise license",
			orgID:      "org_with_license",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "no enterprise license",
			orgID:      "org_without_license",
			wantStatus: http.StatusForbidden,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))
			rr := httptest.NewRecorder()

			handler := RequireEnterpriseLicense(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errorResp ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&errorResp); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if errorResp.Error == "" {
					t.Error("expected error message in response")
				}
				if errorResp.UpgradeURL == "" {
					t.Error("expected upgrade URL in error response")
				}
			}
		})
	}
}

func TestWithBundleAccess(t *testing.T) {
	tests := []struct {
		name           string
		bundleID       string
		orgID          string
		wantStatus     int
		wantError      bool
	}{
		{
			name:           "valid bundle access",
			bundleID:       "security",
			orgID:          "org_with_bundle",
			wantStatus:     http.StatusOK,
			wantError:      false,
		},
		{
			name:           "unauthorized bundle access",
			bundleID:       "security",
			orgID:          "org_without_bundle",
			wantStatus:     http.StatusForbidden,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))
			rr := httptest.NewRecorder()

			handler := WithBundleAccess(tt.bundleID, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errorResp ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&errorResp); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if errorResp.Error == "" {
					t.Error("expected error message in response")
				}
				if errorResp.BundleID != tt.bundleID {
					t.Errorf("expected bundle ID %s in response, got %s", tt.bundleID, errorResp.BundleID)
				}
				if errorResp.UpgradeURL == "" {
					t.Error("expected upgrade URL in error response")
				}
			}
		})
	}
}

func TestCheckBundleFeatures(t *testing.T) {
	tests := []struct {
		name           string
		bundleID       string
		orgID          string
		wantStatus     int
		wantError      bool
	}{
		{
			name:           "bundle features available",
			bundleID:       "security",
			orgID:          "org_with_bundle",
			wantStatus:     http.StatusOK,
			wantError:      false,
		},
		{
			name:           "bundle features not available",
			bundleID:       "security",
			orgID:          "org_without_bundle",
			wantStatus:     http.StatusForbidden,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organization_id", tt.orgID))
			rr := httptest.NewRecorder()

			handler := CheckBundleFeatures(tt.bundleID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errorResp ErrorResponse
				if err := json.NewDecoder(rr.Body).Decode(&errorResp); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if errorResp.Error == "" {
					t.Error("expected error message in response")
				}
				if errorResp.BundleID != tt.bundleID {
					t.Errorf("expected bundle ID %s in response, got %s", tt.bundleID, errorResp.BundleID)
				}
				if errorResp.UpgradeURL == "" {
					t.Error("expected upgrade URL in error response")
				}
				if errorResp.Bundles == nil {
					t.Error("expected available bundles in error response")
				}
			}
		})
	}
}
