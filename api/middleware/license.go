package middleware

import (
    "net/http"

    "github.com/emailimmunity/passwordimmunity/services/featureflag"
    "github.com/go-chi/chi/v5"
)

func RequireFeature(featureService featureflag.Service, feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := r.Context().Value("organization_id").(uuid.UUID)

            enabled, err := featureService.IsFeatureEnabled(r.Context(), orgID, feature)
            if err != nil {
                http.Error(w, "Error checking feature access", http.StatusInternalServerError)
                return
            }

            if !enabled {
                http.Error(w, "Feature not available for your license tier", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func RequireLicenseType(featureService featureflag.Service, requiredType string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := r.Context().Value("organization_id").(uuid.UUID)

            features, err := featureService.GetAvailableFeatures(r.Context(), orgID)
            if err != nil {
                http.Error(w, "Error checking license", http.StatusInternalServerError)
                return
            }

            var hasRequiredAccess bool
            switch requiredType {
            case "enterprise":
                // Check for any enterprise-only feature
                hasRequiredAccess = containsAny(features, []string{
                    "sso", "directory_sync", "enterprise_policies",
                    "advanced_reporting", "custom_roles",
                })
            case "premium":
                // Check for any premium feature
                hasRequiredAccess = containsAny(features, []string{
                    "advanced_2fa", "emergency_access", "priority_support",
                })
            default:
                hasRequiredAccess = true
            }

            if !hasRequiredAccess {
                http.Error(w, "Required license tier not available", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func containsAny(slice []string, elements []string) bool {
    for _, element := range elements {
        for _, item := range slice {
            if item == element {
                return true
            }
        }
    }
    return false
}
