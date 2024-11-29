package middleware

import (
    "context"
    "net/http"

    "github.com/emailimmunity/passwordimmunity/services/enterprise"
)

type FeatureCheckMiddleware struct {
    featureService enterprise.FeatureService
}

func NewFeatureCheckMiddleware(featureService enterprise.FeatureService) *FeatureCheckMiddleware {
    return &FeatureCheckMiddleware{
        featureService: featureService,
    }
}

func (m *FeatureCheckMiddleware) RequireFeature(featureID string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := getOrganizationID(r) // Implement this based on your auth system
            if orgID == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            active, err := m.featureService.IsFeatureActive(r.Context(), orgID, featureID)
            if err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }

            if !active {
                http.Error(w, "Feature not available", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func (m *FeatureCheckMiddleware) RequireBundle(bundleID string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := getOrganizationID(r)
            if orgID == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            active, err := m.featureService.IsBundleActive(r.Context(), orgID, bundleID)
            if err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }

            if !active {
                http.Error(w, "Bundle features not available", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Helper function to get organization ID from request context
// This should be implemented based on your authentication system
func getOrganizationID(r *http.Request) string {
    // Example implementation - replace with your actual auth logic
    if orgID := r.Context().Value("organization_id"); orgID != nil {
        if id, ok := orgID.(string); ok {
            return id
        }
    }
    return ""
}
