package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	scheduler *licensing.ReportScheduler
	licensing *licensing.Service
	logger    logger.Logger
}

// NewHandler creates a new Handler instance with required dependencies
func NewHandler(scheduler *licensing.ReportScheduler, licensing *licensing.Service, logger logger.Logger) *Handler {
	return &Handler{
		scheduler: scheduler,
		licensing: licensing,
		logger:    logger,
	}
}

type RetentionPolicyRequest struct {
	DailyRetention   string `json:"daily_retention"`
	WeeklyRetention  string `json:"weekly_retention"`
	MonthlyRetention string `json:"monthly_retention"`
}

// SetRetentionPolicy handles setting a custom retention policy for an organization
func (h *Handler) SetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")

	var req RetentionPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse duration strings
	daily, err := time.ParseDuration(req.DailyRetention)
	if err != nil {
		http.Error(w, "Invalid daily retention format", http.StatusBadRequest)
		return
	}
	weekly, err := time.ParseDuration(req.WeeklyRetention)
	if err != nil {
		http.Error(w, "Invalid weekly retention format", http.StatusBadRequest)
		return
	}
	monthly, err := time.ParseDuration(req.MonthlyRetention)
	if err != nil {
		http.Error(w, "Invalid monthly retention format", http.StatusBadRequest)
		return
	}

	oldPolicy := h.scheduler.GetRetentionPolicy(orgID)

	policy := licensing.ReportRetentionPolicy{
		DailyReports:   daily,
		WeeklyReports:  weekly,
		MonthlyReports: monthly,
	}

	if err := h.scheduler.SetRetentionPolicy(orgID, policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log policy change
	if err := h.licensing.logRetentionPolicyChange(r.Context(), orgID, "set", &oldPolicy, &policy); err != nil {
		// Log error but don't fail the request
		h.logger.Error("Failed to log retention policy change", "error", err)
	}

	w.WriteHeader(http.StatusOK)
}

// GetRetentionPolicy handles retrieving the current retention policy for an organization
func (h *Handler) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")

	policy := h.scheduler.GetRetentionPolicy(orgID)

	response := RetentionPolicyRequest{
		DailyRetention:   policy.DailyReports.String(),
		WeeklyRetention:  policy.WeeklyReports.String(),
		MonthlyRetention: policy.MonthlyReports.String(),
	}

	json.NewEncoder(w).Encode(response)
}

// RemoveRetentionPolicy handles removing a custom retention policy
func (h *Handler) RemoveRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")

	oldPolicy := h.scheduler.GetRetentionPolicy(orgID)
	h.scheduler.RemoveRetentionPolicy(orgID)

	// Log policy removal
	if err := h.licensing.logRetentionPolicyChange(r.Context(), orgID, "remove", &oldPolicy, nil); err != nil {
		// Log error but don't fail the request
		h.logger.Error("Failed to log retention policy removal", "error", err)
	}

	w.WriteHeader(http.StatusOK)
}
