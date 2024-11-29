package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/emailimmunity/passwordimmunity/api/handlers"
	"github.com/emailimmunity/passwordimmunity/api/middleware"
)

func RegisterRetentionPolicyRoutes(r chi.Router, h *handlers.Handler) {
	r.Route("/api/v1/organizations/{orgID}/retention-policy", func(r chi.Router) {
		// Require enterprise license and admin permissions
		r.Use(middleware.RequireEnterpriseLicense)
		r.Use(middleware.RequireAdmin)

		r.Post("/", h.SetRetentionPolicy)    // Set custom policy
		r.Get("/", h.GetRetentionPolicy)     // Get current policy
		r.Delete("/", h.RemoveRetentionPolicy) // Remove custom policy
	})
}
