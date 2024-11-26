package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emailimmunity/passwordimmunity/api"
)

func TestAPIHandlers(t *testing.T) {
	t.Run("Login Handler", func(t *testing.T) {
		payload := `{"email":"test@example.com","password":"password123"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(payload))
		w := httptest.NewRecorder()

		api.SetupRoutes().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Vault Items Handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/vault/items", nil)
		req.Header.Set("Authorization", "Bearer test_token")
		w := httptest.NewRecorder()

		api.SetupRoutes().ServeHTTP(w, req)

		var resp api.Response
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}
	})

	t.Run("Organization Handler", func(t *testing.T) {
		payload := `{"name":"Test Org","type":"enterprise"}`
		req := httptest.NewRequest("POST", "/api/organizations", strings.NewReader(payload))
		req.Header.Set("Authorization", "Bearer test_token")
		w := httptest.NewRecorder()

		api.SetupRoutes().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}
