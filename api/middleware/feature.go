package middleware

import (
	"net/http"

	"github.com/emailimmunity/passwordimmunity/config"
)

// RequireFeature creates middleware that checks if a feature is active for an organization
func RequireFeature(featureID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get organization ID from context (set by auth middleware)
			orgID := r.Context().Value("organization_id").(string)

			if !config.IsFeatureActive(featureID, orgID) {
				http.Error(w, "Feature not active for organization", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireBundle creates middleware that checks if all features in a bundle are active
func RequireBundle(bundleID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := r.Context().Value("organization_id").(string)

			activeBundles := config.GetActiveBundles(orgID)
			isActive := false
			for _, activeBundle := range activeBundles {
				if activeBundle == bundleID {
					isActive = true
					break
				}
			}

			if !isActive {
				http.Error(w, "Bundle not active for organization", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
