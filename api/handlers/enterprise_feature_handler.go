package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/emailimmunity/passwordimmunity/services/enterprise"
)

type EnterpriseFeatureHandler struct {
    featureService enterprise.FeatureService
}

func NewEnterpriseFeatureHandler(featureService enterprise.FeatureService) *EnterpriseFeatureHandler {
    return &EnterpriseFeatureHandler{
        featureService: featureService,
    }
}

type ActivateFeatureRequest struct {
    FeatureID  string `json:"feature_id"`
    PaymentID  string `json:"payment_id"`
    Duration   string `json:"duration"` // "monthly", "quarterly", "yearly"
}

type ActivateBundleRequest struct {
    BundleID   string `json:"bundle_id"`
    PaymentID  string `json:"payment_id"`
    Duration   string `json:"duration"`
}

func (h *EnterpriseFeatureHandler) ActivateFeature(w http.ResponseWriter, r *http.Request) {
    var req ActivateFeatureRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    duration, err := parseDuration(req.Duration)
    if err != nil {
        http.Error(w, "Invalid duration", http.StatusBadRequest)
        return
    }

    orgID := getOrganizationID(r)
    if orgID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    err = h.featureService.ActivateFeature(r.Context(), orgID, req.FeatureID, req.PaymentID, duration)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "activated"})
}

func (h *EnterpriseFeatureHandler) ActivateBundle(w http.ResponseWriter, r *http.Request) {
    var req ActivateBundleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    duration, err := parseDuration(req.Duration)
    if err != nil {
        http.Error(w, "Invalid duration", http.StatusBadRequest)
        return
    }

    orgID := getOrganizationID(r)
    if orgID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    err = h.featureService.ActivateBundle(r.Context(), orgID, req.BundleID, req.PaymentID, duration)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "activated"})
}

func (h *EnterpriseFeatureHandler) GetActiveFeatures(w http.ResponseWriter, r *http.Request) {
    orgID := getOrganizationID(r)
    if orgID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    features, err := h.featureService.GetActiveFeatures(r.Context(), orgID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "features": features,
    })
}

func (h *EnterpriseFeatureHandler) DeactivateFeature(w http.ResponseWriter, r *http.Request) {
    featureID := r.URL.Query().Get("feature_id")
    if featureID == "" {
        http.Error(w, "Feature ID is required", http.StatusBadRequest)
        return
    }

    orgID := getOrganizationID(r)
    if orgID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    err := h.featureService.DeactivateFeature(r.Context(), orgID, featureID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "deactivated"})
}

func parseDuration(duration string) (time.Duration, error) {
    switch duration {
    case "monthly":
        return 30 * 24 * time.Hour, nil
    case "quarterly":
        return 90 * 24 * time.Hour, nil
    case "yearly":
        return 365 * 24 * time.Hour, nil
    default:
        return 0, fmt.Errorf("invalid duration: %s", duration)
    }
}

// Helper function to get organization ID from request context
func getOrganizationID(r *http.Request) string {
    if orgID := r.Context().Value("organization_id"); orgID != nil {
        if id, ok := orgID.(string); ok {
            return id
        }
    }
    return ""
}
