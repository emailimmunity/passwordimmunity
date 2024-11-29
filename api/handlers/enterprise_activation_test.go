package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

func TestEnterpriseFeatureActivation(t *testing.T) {
	// Setup test services
	featureManager := featureflag.NewFeatureManager()
	paymentService := payment.NewService(&payment.Config{
		MollieAPIKey: "test_key",
		WebhookURL:   "http://test.com/webhook",
	})
	licenseService := licensing.NewService()

	handler := NewEnterpriseHandler(featureManager)

	t.Run("ActivateFeature", func(t *testing.T) {
		req := FeatureActivationRequest{
			FeatureID: "advanced_sso",
			OrgID:     "test_org",
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/enterprise/features/activate", bytes.NewBuffer(body))
		r = r.WithContext(WithTestContext(r.Context(), "test_org"))

		handler.ActivateFeature(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp FeatureActivationResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.PaymentURL == "" {
			t.Error("Expected payment URL, got empty string")
		}
	})

	t.Run("VerifyFeatureActivation", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/webhooks/mollie?id=test_payment", nil)

		handler.VerifyFeatureActivation(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("InvalidFeature", func(t *testing.T) {
		req := FeatureActivationRequest{
			FeatureID: "invalid_feature",
			OrgID:     "test_org",
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/enterprise/features/activate", bytes.NewBuffer(body))
		r = r.WithContext(WithTestContext(r.Context(), "test_org"))

		handler.ActivateFeature(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// WithTestContext adds test organization ID to context
func WithTestContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, "organization_id", orgID)
}
