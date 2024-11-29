package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emailimmunity/passwordimmunity/services/logger"
	"github.com/go-chi/chi/v5"
)

// createTestContext creates a new context with route parameters and user ID for testing
func createTestContext(r *http.Request, orgID, userID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("orgID", orgID)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(context.WithValue(ctx, "user_id", userID))
}

// createTestRecorder creates a new response recorder for testing
func createTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// createTestLogger creates a new logger instance for testing
func createTestLogger(t *testing.T) logger.Logger {
	return logger.NewTestLogger()
}

// assertStatusCode checks if the response status code matches the expected value
func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Status code = %d; want %d", got, want)
	}
}

// assertLogEntry checks if a log entry exists with the expected properties
func assertLogEntry(t *testing.T, events []interface{}, expectedAction string, expectedUserID string) {
	t.Helper()
	if len(events) == 0 {
		t.Error("Expected audit log entry, got none")
		return
	}

	lastEvent := events[len(events)-1]
	if e, ok := lastEvent.(map[string]interface{}); ok {
		if action, ok := e["action"].(string); !ok || action != expectedAction {
			t.Errorf("Expected action %q, got %v", expectedAction, e["action"])
		}
		if userID, ok := e["user_id"].(string); !ok || userID != expectedUserID {
			t.Errorf("Expected user_id %q, got %v", expectedUserID, e["user_id"])
		}
	} else {
		t.Errorf("Expected map[string]interface{}, got %T", lastEvent)
	}
}
