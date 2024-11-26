package router

import (
    "github.com/emailimmunity/passwordimmunity/api/handlers"
    "github.com/emailimmunity/passwordimmunity/api/middleware"
    "github.com/go-chi/chi/v5"
)

func RegisterEnterpriseRoutes(r chi.Router, h *handlers.Handlers, m *middleware.Middleware) {
    // SSO Routes
    r.Route("/sso", func(r chi.Router) {
        r.Use(m.RequireFeature("sso"))
        r.Mount("/", h.SSO.Routes())
    })

    // Directory Sync Routes
    r.Route("/directory", func(r chi.Router) {
        r.Use(m.RequireFeature("directory_sync"))
        r.Mount("/", h.DirectorySync.Routes())
    })

    // Advanced Reporting Routes
    r.Route("/reports", func(r chi.Router) {
        r.Use(m.RequireFeature("advanced_reporting"))
        r.Mount("/", h.Reports.Routes())
    })

    // Custom Roles Routes
    r.Route("/roles", func(r chi.Router) {
        r.Use(m.RequireFeature("custom_roles"))
        r.Mount("/", h.Roles.Routes())
    })

    // Advanced Groups Routes
    r.Route("/groups", func(r chi.Router) {
        r.Use(m.RequireFeature("advanced_groups"))
        r.Mount("/", h.Groups.Routes())
    })

    // Multi-tenant Management Routes
    r.Route("/organizations", func(r chi.Router) {
        r.Use(m.RequireFeature("multi_tenant"))
        r.Mount("/", h.Organizations.Routes())
    })

    // Advanced Vault Management Routes
    r.Route("/vault/advanced", func(r chi.Router) {
        r.Use(m.RequireFeature("advanced_vault"))
        r.Mount("/", h.AdvancedVault.Routes())
    })

    // Cross-organization Management Routes
    r.Route("/cross-org", func(r chi.Router) {
        r.Use(m.RequireFeature("cross_org_management"))
        r.Mount("/", h.CrossOrg.Routes())
    })

    // Enterprise Policies Routes
    r.Route("/policies", func(r chi.Router) {
        r.Use(m.RequireFeature("enterprise_policies"))
        r.Mount("/", h.Policies.Routes())
    })

    // API Access Routes
    r.Route("/api", func(r chi.Router) {
        r.Use(m.RequireFeature("api_access"))
        r.Mount("/", h.API.Routes())
    })
}
