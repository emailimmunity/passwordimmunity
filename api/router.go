package api

import (
    "github.com/emailimmunity/passwordimmunity/api/handlers"
    "github.com/emailimmunity/passwordimmunity/api/middleware"
    "github.com/emailimmunity/passwordimmunity/api/router"
    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *handlers.Handlers, m *middleware.Middleware) *chi.Mux {
    r := chi.NewRouter()

    // Global middleware
    r.Use(chimiddleware.Logger)
    r.Use(chimiddleware.Recoverer)
    r.Use(chimiddleware.RequestID)
    r.Use(middleware.Cors)

    // Public routes
    r.Group(func(r chi.Router) {
        r.Post("/api/v1/payments/webhook", h.Payment.HandleWebhook)
    })

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(m.AuthMiddleware)
        r.Use(m.OrganizationContext)

        // Core routes (free tier)
        r.Route("/api/v1", func(r chi.Router) {
            // Basic vault operations
            r.Mount("/vault", h.Vault.Routes())
            // Basic user management
            r.Mount("/users", h.Users.Routes())
            // Basic organization management
            r.Mount("/organizations", h.Organizations.Routes())
            // Payment and license management
            r.Mount("/payments", h.Payment.Routes())
            r.Mount("/license", h.License.Routes())
        })

        // Premium features
        r.Route("/api/v1/premium", func(r chi.Router) {
            r.Use(m.RequireLicenseType("premium"))
            router.RegisterPremiumRoutes(r, h, m)
        })

        // Enterprise features
        r.Route("/api/v1/enterprise", func(r chi.Router) {
            r.Use(m.RequireLicenseType("enterprise"))
            router.RegisterEnterpriseRoutes(r, h, m)
        })
    })

    return r
}
