package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/emailimmunity/passwordimmunity/api/handlers"
	"github.com/emailimmunity/passwordimmunity/api/middleware"
)

func RegisterEnterpriseRoutes(
	r chi.Router,
	enterpriseHandler *handlers.EnterpriseFeatureHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	r.Route("/admin/enterprise", func(r chi.Router) {
		// Apply admin authentication middleware
		r.Use(authMiddleware.RequireAdmin)

		// Feature management endpoints
		r.Get("/features", enterpriseHandler.ListFeatures)
		r.Get("/features/status", enterpriseHandler.GetFeatureStatus)

		// Serve admin panel template
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "templates/admin/enterprise_features.html")
		})
	})

	// Serve static assets
	r.Get("/static/css/admin/enterprise_features.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "static/css/admin/enterprise_features.css")
	})
}
