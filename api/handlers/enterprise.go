package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type EnterpriseHandler struct {
	featureManager *featureflag.FeatureManager
}

func NewEnterpriseHandler(fm *featureflag.FeatureManager) *EnterpriseHandler {
	return &EnterpriseHandler{
		featureManager: fm,
	}
}

// GetFeatureStatus returns the status of all enterprise features for the current license
func (h *EnterpriseHandler) GetFeatureStatus(w http.ResponseWriter, r *http.Request) {
	features := h.featureManager.GetEnabledFeatures(r.Context())

	response := struct {
		Features []featureflag.EnterpriseFeatureFlag `json:"features"`
	}{
		Features: features,
	}

	json.NewEncoder(w).Encode(response)
}

// ActivateFeature activates an enterprise feature after payment verification
func (h *EnterpriseHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	featureID := vars["feature"]

	// Verify payment and license status
	// This will be called by the Mollie webhook after successful payment
	// Implementation will integrate with the payment service

	w.WriteHeader(http.StatusOK)
}

// DeactivateFeature handles feature deactivation (e.g., after license expiration)
func (h *EnterpriseHandler) DeactivateFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	featureID := vars["feature"]

	// Handle feature deactivation logic
	// This might be called by a license management system or admin

	w.WriteHeader(http.StatusOK)
}

// GetLicenseInfo returns detailed information about the current license
func (h *EnterpriseHandler) GetLicenseInfo(w http.ResponseWriter, r *http.Request) {
	license := r.Context().Value("license").(*licensing.License)

	response := struct {
		Type       licensing.LicenseType `json:"type"`
		ValidUntil string               `json:"validUntil"`
		Features   []string             `json:"features"`
	}{
		Type:       license.Type,
		ValidUntil: license.ValidUntil.Format("2006-01-02"),
		Features:   license.Features,
	}

	json.NewEncoder(w).Encode(response)
}
