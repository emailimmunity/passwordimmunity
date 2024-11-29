package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/logger"
)

const (
	MinDailyRetention   = 24 * time.Hour
	MinWeeklyRetention  = 7 * 24 * time.Hour
	MinMonthlyRetention = 30 * 24 * time.Hour
)

type RetentionPolicyRequest struct {
	DailyRetention   string `json:"daily_retention"`
	WeeklyRetention  string `json:"weekly_retention"`
	MonthlyRetention string `json:"monthly_retention"`
}

type RetentionPolicyMiddleware struct {
	licensing *licensing.Service
	logger    logger.Logger
}

func NewRetentionPolicyMiddleware(licensing *licensing.Service, logger logger.Logger) *RetentionPolicyMiddleware {
	return &RetentionPolicyMiddleware{
		licensing: licensing,
		logger:    logger,
	}
}

// ValidateRetentionPolicy validates retention policy requests and checks enterprise licensing
func (m *RetentionPolicyMiddleware) ValidateRetentionPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		orgID := r.Context().Value("organization_id").(string)
		userID := r.Context().Value("user_id").(string)

		// Check enterprise license
		if !m.licensing.HasEnterpriseFeature(orgID, "custom_retention_policy") {
			m.logger.Warn("Unauthorized retention policy access attempt",
				"user_id", userID,
				"organization_id", orgID,
			)
			http.Error(w, "Enterprise license required for custom retention policies", http.StatusForbidden)
			return
		}

		var req RetentionPolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			m.logger.Error("Invalid retention policy request",
				"error", err,
				"user_id", userID,
				"organization_id", orgID,
			)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Validate durations
		daily, err := time.ParseDuration(req.DailyRetention)
		if err != nil || daily < MinDailyRetention {
			http.Error(w, "Daily retention must be at least 24 hours", http.StatusBadRequest)
			return
		}

		weekly, err := time.ParseDuration(req.WeeklyRetention)
		if err != nil || weekly < MinWeeklyRetention {
			http.Error(w, "Weekly retention must be at least 7 days", http.StatusBadRequest)
			return
		}

		monthly, err := time.ParseDuration(req.MonthlyRetention)
		if err != nil || monthly < MinMonthlyRetention {
			http.Error(w, "Monthly retention must be at least 30 days", http.StatusBadRequest)
			return
		}

		m.logger.Info("Valid retention policy request",
			"user_id", userID,
			"organization_id", orgID,
			"daily", req.DailyRetention,
			"weekly", req.WeeklyRetention,
			"monthly", req.MonthlyRetention,
		)

		// Store validated request in context
		ctx := context.WithValue(r.Context(), "retention_policy", req)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
