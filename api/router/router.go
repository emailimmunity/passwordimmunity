package router

import (
	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/api/handlers"
	"github.com/emailimmunity/passwordimmunity/api/middleware"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

func NewRouter(
	featureManager *featureflag.FeatureManager,
	paymentService *payment.Service,
	enterpriseFeatureHandler *handlers.EnterpriseFeatureHandler,
) *mux.Router {
	r := mux.NewRouter()

	// Base middleware for all routes
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	public := r.PathPrefix("/api").Subrouter()
	public.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	public.HandleFunc("/version", handlers.Version).Methods("GET")

	// Auth routes
	auth := r.PathPrefix("/api/auth").Subrouter()
	auth.HandleFunc("/login", handlers.Login).Methods("POST")
	auth.HandleFunc("/register", handlers.Register).Methods("POST")
	auth.HandleFunc("/forgot-password", handlers.ForgotPassword).Methods("POST")
	auth.HandleFunc("/reset-password", handlers.ResetPassword).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.RequireAuthentication)

	// User routes
	users := protected.PathPrefix("/users").Subrouter()
	users.HandleFunc("", handlers.ListUsers).Methods("GET")
	users.HandleFunc("/{id}", handlers.GetUser).Methods("GET")
	users.HandleFunc("/{id}", handlers.UpdateUser).Methods("PUT")
	users.HandleFunc("/{id}", handlers.DeleteUser).Methods("DELETE")

	// Vault routes
	vault := protected.PathPrefix("/vault").Subrouter()
	vault.HandleFunc("/items", handlers.ListVaultItems).Methods("GET")
	vault.HandleFunc("/items", handlers.CreateVaultItem).Methods("POST")
	vault.HandleFunc("/items/{id}", handlers.GetVaultItem).Methods("GET")
	vault.HandleFunc("/items/{id}", handlers.UpdateVaultItem).Methods("PUT")
	vault.HandleFunc("/items/{id}", handlers.DeleteVaultItem).Methods("DELETE")

	// Payment and licensing routes
	payment := r.PathPrefix("/api/payment").Subrouter()
	payment.HandleFunc("/webhook", handlers.PaymentWebhook).Methods("POST")
	payment.HandleFunc("/activate", enterpriseFeatureHandler.ActivateFeature).Methods("POST")

	// Payment webhook endpoints (public)
	r.HandleFunc("/api/webhooks/mollie", handlers.PaymentWebhook).Methods("POST")

	// Register all enterprise-specific routes (including feature management, license management,
	// retention policies, and payments) through the centralized registration function
	RegisterEnterpriseRoutes(r, featureManager, paymentService, enterpriseFeatureHandler)

	return r
}

	return r
}
