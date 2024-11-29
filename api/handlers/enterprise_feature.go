package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type EnterpriseFeatureHandler struct {
	featureManager *featureflag.FeatureManager
	paymentService *payment.Service
}

func NewEnterpriseFeatureHandler(fm *featureflag.FeatureManager, ps *payment.Service) *EnterpriseFeatureHandler {
	return &EnterpriseFeatureHandler{
		featureManager: fm,
		paymentService: ps,
	}
}

type ActivateFeatureRequest struct {
	TierID    string `json:"tier_id,omitempty"`
	FeatureID string `json:"feature_id,omitempty"`
	BundleID  string `json:"bundle_id,omitempty"`
}

func (h *EnterpriseFeatureHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
	var req ActivateFeatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Handle tier-based activation
	if req.TierID != "" {
		activation := featureflag.FeatureActivation{
			TierID: req.TierID,
		}
		if err := h.featureManager.ActivateFeature(ctx, activation); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle individual feature activation
	if req.FeatureID != "" {
		activation := featureflag.FeatureActivation{
			FeatureID: req.FeatureID,
		}
		if err := h.featureManager.ActivateFeature(ctx, activation); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle bundle activation
	if req.BundleID != "" {
		activation := featureflag.FeatureActivation{
			BundleID: req.BundleID,
		}
		if err := h.featureManager.ActivateFeature(ctx, activation); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Must specify either tier_id, feature_id, or bundle_id", http.StatusBadRequest)
}

func (h *EnterpriseFeatureHandler) GetFeatureStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := h.featureManager.GetFeatureStatus(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
