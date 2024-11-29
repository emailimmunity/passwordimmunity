package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/logger"
	"github.com/go-chi/chi/v5"
)

type mockLicensingService struct {
	*licensing.Service
	loggedEvents []licensing.RetentionPolicyAuditEvent
}

func (m *mockLicensingService) logRetentionPolicyChange(ctx context.Context, orgID string, action string, oldPolicy, newPolicy *licensing.Policy) error {
	m.loggedEvents = append(m.loggedEvents, licensing.RetentionPolicyAuditEvent{
		OrganizationID: orgID,
		Action:         action,
		OldPolicy:      oldPolicy,
		NewPolicy:      newPolicy,
		UserID:         ctx.Value("user_id").(string),
	})
	return nil
}

func TestRetentionPolicyHandlers(t *testing.T) {
	mockLic := &mockLicensingService{
		Service: &licensing.Service{
			storage: licensing.NewReportStorage(t.TempDir()),
		},
	}
	scheduler := licensing.NewReportScheduler(mockLic)
	testLogger := createTestLogger(t)
	handler := NewHandler(scheduler, mockLic, testLogger)

	t.Run("Set Retention Policy", func(t *testing.T) {
		req := RetentionPolicyRequest{
			DailyRetention:   "48h",
			WeeklyRetention:  "168h",
			MonthlyRetention: "720h",
		}
		body, _ := json.Marshal(req)

		r := httptest.NewRequest("POST", "/api/v1/organizations/org1/retention-policy", bytes.NewBuffer(body))
		r = createTestContext(r, "org1", "test-user")
		w := createTestRecorder()

		handler.SetRetentionPolicy(w, r)
		assertStatusCode(t, w.Code, http.StatusOK)

		// Verify policy was set
		policy := scheduler.GetRetentionPolicy("org1")
		if policy.DailyReports != 48*time.Hour {
			t.Error("Policy not set correctly")
		}
	})

	t.Run("Get Retention Policy", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v1/organizations/org1/retention-policy", nil)
		r = createTestContext(r, "org1", "test-user")
		w := createTestRecorder()

		handler.GetRetentionPolicy(w, r)
		assertStatusCode(t, w.Code, http.StatusOK)

		var response RetentionPolicyRequest
		json.NewDecoder(w.Body).Decode(&response)

		if response.DailyRetention != "48h" {
			t.Error("Incorrect policy returned")
		}
	})

	t.Run("Remove Retention Policy", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/v1/organizations/org1/retention-policy", nil)
		r = createTestContext(r, "org1", "test-user")
		w := createTestRecorder()

		handler.RemoveRetentionPolicy(w, r)
		assertStatusCode(t, w.Code, http.StatusOK)

		// Verify policy was removed
		policy := scheduler.GetRetentionPolicy("org1")
		if policy != licensing.DefaultRetentionPolicy {
			t.Error("Policy not reset to default")
		}
	})

	t.Run("Set Retention Policy with Audit", func(t *testing.T) {
		req := RetentionPolicyRequest{
			DailyRetention:   "48h",
			WeeklyRetention:  "168h",
			MonthlyRetention: "720h",
		}
		body, _ := json.Marshal(req)

		r := httptest.NewRequest("POST", "/api/v1/organizations/org2/retention-policy", bytes.NewBuffer(body))
		r = createTestContext(r, "org2", "test-user")
		w := createTestRecorder()

		handler.SetRetentionPolicy(w, r)
		assertStatusCode(t, w.Code, http.StatusOK)
		assertLogEntry(t, mockLic.loggedEvents, "set", "test-user")
	})

	t.Run("Remove Retention Policy with Audit", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/v1/organizations/org2/retention-policy", nil)
		r = createTestContext(r, "org2", "test-user")
		w := createTestRecorder()

		handler.RemoveRetentionPolicy(w, r)
		assertStatusCode(t, w.Code, http.StatusOK)
		assertLogEntry(t, mockLic.loggedEvents, "remove", "test-user")
	})
}

		if len(mockLic.loggedEvents) == 0 {
			t.Error("Expected audit log entry, got none")
		}

		lastEvent := mockLic.loggedEvents[len(mockLic.loggedEvents)-1]
		if lastEvent.Action != "set" || lastEvent.UserID != "test-user" {
			t.Error("Audit log entry not properly recorded")
		}
	})

	t.Run("Get Retention Policy", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v1/organizations/org1/retention-policy", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("orgID", "org1")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		handler.GetRetentionPolicy(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		var response RetentionPolicyRequest
		json.NewDecoder(w.Body).Decode(&response)

		if response.DailyRetention != "48h" {
			t.Error("Incorrect policy returned")
		}
	})

	t.Run("Remove Retention Policy", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/v1/organizations/org1/retention-policy", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("orgID", "org1")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		handler.RemoveRetentionPolicy(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		// Verify policy was removed
		policy := scheduler.GetRetentionPolicy("org1")
		if policy != licensing.DefaultRetentionPolicy {
			t.Error("Policy not reset to default")
		}
	})

	t.Run("Remove Retention Policy with Audit", func(t *testing.T) {
		r := httptest.NewRequest("DELETE", "/api/v1/organizations/org2/retention-policy", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("orgID", "org2")
		r = r.WithContext(context.WithValue(
			context.WithValue(r.Context(), chi.RouteCtxKey, rctx),
			"user_id", "test-user",
		))

		handler.RemoveRetentionPolicy(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		if len(mockLic.loggedEvents) == 0 {
			t.Error("Expected audit log entry, got none")
		}

		lastEvent := mockLic.loggedEvents[len(mockLic.loggedEvents)-1]
		if lastEvent.Action != "remove" || lastEvent.UserID != "test-user" {
			t.Error("Audit log entry not properly recorded")
		}
	})
}
