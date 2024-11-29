package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateRetentionPolicy(t *testing.T) {
	tests := []struct {
		name           string
		request        RetentionPolicyRequest
		expectedStatus int
	}{
		{
			name: "Valid Policy",
			request: RetentionPolicyRequest{
				DailyRetention:   "48h",
				WeeklyRetention:  "168h",
				MonthlyRetention: "720h",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Daily Retention",
			request: RetentionPolicyRequest{
				DailyRetention:   "12h", // Less than minimum
				WeeklyRetention:  "168h",
				MonthlyRetention: "720h",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Weekly Retention",
			request: RetentionPolicyRequest{
				DailyRetention:   "48h",
				WeeklyRetention:  "120h", // Less than minimum
				MonthlyRetention: "720h",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Monthly Retention",
			request: RetentionPolicyRequest{
				DailyRetention:   "48h",
				WeeklyRetention:  "168h",
				MonthlyRetention: "500h", // Less than minimum
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/retention-policy", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler := ValidateRetentionPolicy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, w.Code)
			}
		})
	}
}
