package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name       string
        token      string
        wantStatus int
    }{
        {
            name:       "Valid token",
            token:      "Bearer valid-token",
            wantStatus: http.StatusOK,
        },
        {
            name:       "Missing token",
            token:      "",
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "Invalid token format",
            token:      "invalid-token",
            wantStatus: http.StatusUnauthorized,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            }))

            req := httptest.NewRequest("GET", "/", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", tt.token)
            }

            rr := httptest.NewRecorder()
            handler.ServeHTTP(rr, req)

            assert.Equal(t, tt.wantStatus, rr.Code)
        })
    }
}

func TestOrganizationContext(t *testing.T) {
    handler := OrganizationContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        orgID := r.Context().Value("organization_id")
        assert.NotNil(t, orgID)
        _, ok := orgID.(uuid.UUID)
        assert.True(t, ok)
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)
    assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCors(t *testing.T) {
    handler := Cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    tests := []struct {
        name           string
        method         string
        wantStatus     int
        checkCorsFlags bool
    }{
        {
            name:           "Normal request",
            method:         "GET",
            wantStatus:     http.StatusOK,
            checkCorsFlags: true,
        },
        {
            name:           "Options request",
            method:         "OPTIONS",
            wantStatus:     http.StatusOK,
            checkCorsFlags: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, "/", nil)
            rr := httptest.NewRecorder()

            handler.ServeHTTP(rr, req)

            assert.Equal(t, tt.wantStatus, rr.Code)

            if tt.checkCorsFlags {
                assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
                assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
                assert.Equal(t, "Accept, Authorization, Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))
                assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
            }
        })
    }
}
