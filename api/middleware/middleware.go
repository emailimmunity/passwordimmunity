package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/google/uuid"
)

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // TODO: Implement proper JWT validation
        // For now, just checking if token exists
        next.ServeHTTP(w, r)
    })
}

func OrganizationContext(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // TODO: Extract organization ID from validated JWT
        // For now, using a mock organization ID
        orgID := uuid.New()
        ctx := context.WithValue(r.Context(), "organization_id", orgID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func Cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func extractToken(r *http.Request) string {
    bearerToken := r.Header.Get("Authorization")
    if len(strings.Split(bearerToken, " ")) == 2 {
        return strings.Split(bearerToken, " ")[1]
    }
    return ""
}
