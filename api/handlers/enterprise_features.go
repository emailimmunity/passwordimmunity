package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/emailimmunity/passwordimmunity/services/enterprise"
	"github.com/shopspring/decimal"
)

type EnterpriseFeatureHandler struct {
	featureManager *enterprise.FeatureManager
	logger         Logger
}

func NewEnterpriseFeatureHandler(featureManager *enterprise.FeatureManager, logger Logger) *EnterpriseFeatureHandler {
	return &EnterpriseFeatureHandler{
		featureManager: featureManager,
		logger:        logger,
	}
}

type FeatureStatusResponse struct {
	FeatureID    string   `json:"featureId"`
	IsEnabled    bool     `json:"isEnabled"`
	ExpiresAt    string   `json:"expiresAt,omitempty"`
	BundleID     string   `json:"bundleId,omitempty"`
	TierID       string   `json:"tierId,omitempty"`
	ActiveBundle []string `json:"activeBundle,omitempty"`
}

type ActivateFeatureRequest struct {
	FeatureID  string  `json:"featureId,omitempty"`
	BundleID   string  `json:"bundleId,omitempty"`
	TierID     string  `json:"tierId,omitempty"`
	Currency   string  `json:"currency"`
	Amount     string  `json:"amount"`
	IsYearly   bool    `json:"isYearly"`
	PaymentID  string  `json:"paymentId"`
}

func (h *EnterpriseFeatureHandler) GetFeatureStatus(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Context().Value("organizationID").(string)
	featureID := r.URL.Query().Get("featureId")

	if featureID == "" {
		http.Error(w, "featureId is required", http.StatusBadRequest)
		return
	}

	isEnabled := h.featureManager.IsFeatureEnabled(r.Context(), organizationID, featureID)

	activeFeatures, err := h.featureManager.GetActiveFeatures(r.Context(), organizationID)
	if err != nil {
		h.logger.Error("Failed to get active features",
			"error", err,
			"organization_id", organizationID,
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := FeatureStatusResponse{
		FeatureID: featureID,
		IsEnabled: isEnabled,
		ActiveBundle: activeFeatures,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *EnterpriseFeatureHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
	var req ActivateFeatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	organizationID := r.Context().Value("organizationID").(string)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	activation := enterprise.FeatureActivation{
		OrganizationID: organizationID,
		FeatureID:      req.FeatureID,
		BundleID:       req.BundleID,
		TierID:         req.TierID,
		Currency:       req.Currency,
		Amount:         amount,
		IsYearly:       req.IsYearly,
		PaymentID:      req.PaymentID,
		ExpiresAt:      time.Now().AddDate(req.IsYearly ? 1 : 0, req.IsYearly ? 0 : 1, 0),
	}

	if err := h.featureManager.ActivateFeature(r.Context(), activation); err != nil {
		h.logger.Error("Failed to activate feature",
			"error", err,
			"organization_id", organizationID,
		)
		http.Error(w, "Failed to activate feature", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EnterpriseFeatureHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Context().Value("organizationID").(string)

	activeFeatures, err := h.featureManager.GetActiveFeatures(r.Context(), organizationID)
	if err != nil {
		h.logger.Error("Failed to get active features",
			"error", err,
			"organization_id", organizationID,
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"features": activeFeatures,
	})
}
