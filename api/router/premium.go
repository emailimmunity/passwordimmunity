package router

import (
    "github.com/emailimmunity/passwordimmunity/api/handlers"
    "github.com/emailimmunity/passwordimmunity/api/middleware"
    "github.com/go-chi/chi/v5"
)

func RegisterPremiumRoutes(r chi.Router, h *handlers.Handlers, m *middleware.Middleware) {
    // Advanced 2FA Routes
    r.Route("/2fa", func(r chi.Router) {
        r.Use(m.RequireFeature("advanced_2fa"))
        r.Mount("/", h.TwoFactor.Routes())
    })

    // Emergency Access Routes
    r.Route("/emergency-access", func(r chi.Router) {
        r.Use(m.RequireFeature("emergency_access"))
        r.Mount("/", h.EmergencyAccess.Routes())
    })

    // Priority Support Routes
    r.Route("/support", func(r chi.Router) {
        r.Use(m.RequireFeature("priority_support"))
        r.Mount("/", h.Support.Routes())
    })

    // Basic API Access Routes
    r.Route("/api/basic", func(r chi.Router) {
        r.Use(m.RequireFeature("basic_api_access"))
        r.Mount("/", h.BasicAPI.Routes())
    })

    // Basic Reporting Routes
    r.Route("/reports/basic", func(r chi.Router) {
        r.Use(m.RequireFeature("basic_reporting"))
        r.Mount("/", h.BasicReports.Routes())
    })
}
