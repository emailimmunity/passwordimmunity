package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/emailimmunity/passwordimmunity/services/licensing"
    "github.com/emailimmunity/passwordimmunity/services/featureflag"
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type LicenseHandler struct {
    licenseService     licensing.Service
    featureFlagService featureflag.Service
}

func NewLicenseHandler(ls licensing.Service, fs featureflag.Service) *LicenseHandler {
    return &LicenseHandler{
        licenseService:     ls,
        featureFlagService: fs,
    }
}

func (h *LicenseHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Get("/", h.GetLicense)
    r.Get("/features", h.GetFeatures)
    r.Get("/check/{feature}", h.CheckFeature)
    return r
}

func (h *LicenseHandler) GetLicense(w http.ResponseWriter, r *http.Request) {
    orgID := r.Context().Value("organization_id").(uuid.UUID)

    license, err := h.licenseService.GetActiveLicense(r.Context(), orgID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if license == nil {
        http.Error(w, "No active license found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(license)
}

func (h *LicenseHandler) GetFeatures(w http.ResponseWriter, r *http.Request) {
    orgID := r.Context().Value("organization_id").(uuid.UUID)

    features, err := h.featureFlagService.GetAvailableFeatures(r.Context(), orgID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string][]string{"features": features})
}

func (h *LicenseHandler) CheckFeature(w http.ResponseWriter, r *http.Request) {
    orgID := r.Context().Value("organization_id").(uuid.UUID)
    feature := chi.URLParam(r, "feature")

    enabled, err := h.featureFlagService.IsFeatureEnabled(r.Context(), orgID, feature)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"enabled": enabled})
}
