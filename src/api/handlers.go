package api

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error     `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Handler types for different API endpoints
type (
	AuthHandler struct {
		// Dependencies will be added
	}

	VaultHandler struct {
		// Dependencies will be added
	}

	OrganizationHandler struct {
		// Dependencies will be added
	}

	RoleHandler struct {
		// Dependencies will be added
	}
)

// Common handler methods
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, code, message string) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	sendJSON(w, status, resp)
}

// Basic routes setup
func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/2fa", handle2FA)

	// Vault routes
	mux.HandleFunc("/api/vault/items", handleVaultItems)
	mux.HandleFunc("/api/vault/items/", handleVaultItem)

	// Organization routes
	mux.HandleFunc("/api/organizations", handleOrganizations)
	mux.HandleFunc("/api/organizations/", handleOrganization)

	// Role routes
	mux.HandleFunc("/api/roles", handleRoles)
	mux.HandleFunc("/api/roles/", handleRole)

	return mux
}

// Placeholder handlers - implementations will be added in separate PRs
func handleLogin(w http.ResponseWriter, r *http.Request)        {}
func handleRegister(w http.ResponseWriter, r *http.Request)     {}
func handle2FA(w http.ResponseWriter, r *http.Request)          {}
func handleVaultItems(w http.ResponseWriter, r *http.Request)   {}
func handleVaultItem(w http.ResponseWriter, r *http.Request)    {}
func handleOrganizations(w http.ResponseWriter, r *http.Request){}
func handleOrganization(w http.ResponseWriter, r *http.Request) {}
func handleRoles(w http.ResponseWriter, r *http.Request)        {}
func handleRole(w http.ResponseWriter, r *http.Request)         {}
