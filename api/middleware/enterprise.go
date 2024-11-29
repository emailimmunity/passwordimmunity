package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type ErrorResponse struct {
	Error       string   `json:"error"`
	FeatureID   string   `json:"feature_id,omitempty"`
	BundleID    string   `json:"bundle_id,omitempty"`
	UpgradeURL  string   `json:"upgrade_url,omitempty"`
	Bundles     []string `json:"available_bundles,omitempty"`
}

// EnterpriseFeatureMiddleware checks if the requested feature is available under the current license
func EnterpriseFeatureMiddleware(featureID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get organization ID from context
			orgID := r.Context().Value("organization_id").(string)

			// Check if feature requires enterprise license
			if !featureflag.IsEnterpriseFeature(featureID) {
				next.ServeHTTP(w, r)
				return
			}

			// Verify license for the feature
			licenseService := licensing.GetService()
			if !licenseService.HasFeatureAccess(orgID, featureID) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:      "Enterprise license required for this feature",
					FeatureID:  featureID,
					UpgradeURL: "/dashboard/subscription/features/" + featureID,
				})
				return
			}

			// Feature is licensed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// RequireEnterpriseLicense is a middleware that ensures the organization has any valid enterprise license
func RequireEnterpriseLicense(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Context().Value("organization_id").(string)

		licenseService := licensing.GetService()
		if !licenseService.HasValidLicense(orgID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:      "Valid enterprise license required",
				UpgradeURL: "/dashboard/subscription/enterprise",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WithFeatureAccess wraps an API handler to check for specific feature access
func WithFeatureAccess(featureID string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Context().Value("organization_id").(string)

		licenseService := licensing.GetService()
		if !licenseService.HasFeatureAccess(orgID, featureID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:      "Feature not available under current license",
				FeatureID:  featureID,
				UpgradeURL: "/dashboard/subscription/features/" + featureID,
			})
			return
		}

		handler(w, r)
	}
}

// CheckGracePeriod verifies if a feature is in its grace period
func CheckGracePeriod(featureID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := r.Context().Value("organization_id").(string)

			licenseService := licensing.GetService()
			if licenseService.IsInGracePeriod(orgID, featureID) {
				// Add grace period warning header
				w.Header().Set("X-Grace-Period-Warning", "Feature access will expire soon")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithBundleAccess wraps an API handler to check for bundle access
func WithBundleAccess(bundleID string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Context().Value("organization_id").(string)

		licenseService := licensing.GetService()
		if !licenseService.HasBundleAccess(orgID, bundleID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:     "Bundle not available under current license",
				BundleID:  bundleID,
				UpgradeURL: "/dashboard/subscription/bundles/" + bundleID,
			})
			return
		}

		handler(w, r)
	}
}

// CheckBundleFeatures verifies access to all features in a bundle
func CheckBundleFeatures(bundleID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := r.Context().Value("organization_id").(string)
			licenseService := licensing.GetService()

			if !licenseService.HasBundleAccess(orgID, bundleID) {
				availableBundles := licenseService.GetAvailableBundles(orgID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:     "Bundle features not available",
					BundleID:  bundleID,
					Bundles:   availableBundles,
					UpgradeURL: "/dashboard/subscription/bundles/" + bundleID,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
