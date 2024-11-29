package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/emailimmunity/passwordimmunity/services/enterprise"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockFeatureService struct {
    mock.Mock
}

func (m *mockFeatureService) ActivateFeature(ctx context.Context, orgID, featureID, paymentID string, duration time.Duration) error {
    args := m.Called(ctx, orgID, featureID, paymentID, duration)
    return args.Error(0)
}

func (m *mockFeatureService) ActivateBundle(ctx context.Context, orgID, bundleID, paymentID string, duration time.Duration) error {
    args := m.Called(ctx, orgID, bundleID, paymentID, duration)
    return args.Error(0)
}

func (m *mockFeatureService) IsFeatureActive(ctx context.Context, orgID, featureID string) (bool, error) {
    args := m.Called(ctx, orgID, featureID)
    return args.Bool(0), args.Error(1)
}

func (m *mockFeatureService) IsBundleActive(ctx context.Context, orgID, bundleID string) (bool, error) {
    args := m.Called(ctx, orgID, bundleID)
    return args.Bool(0), args.Error(1)
}

func (m *mockFeatureService) GetActiveFeatures(ctx context.Context, orgID string) ([]string, error) {
    args := m.Called(ctx, orgID)
    return args.Get(0).([]string), args.Error(1)
}

func (m *mockFeatureService) DeactivateFeature(ctx context.Context, orgID, featureID string) error {
    args := m.Called(ctx, orgID, featureID)
    return args.Error(0)
}

func (m *mockFeatureService) DeactivateBundle(ctx context.Context, orgID, bundleID string) error {
    args := m.Called(ctx, orgID, bundleID)
    return args.Error(0)
}

func TestEnterpriseFeatureHandler(t *testing.T) {
    featureService := new(mockFeatureService)
    handler := NewEnterpriseFeatureHandler(featureService)

    t.Run("ActivateFeature", func(t *testing.T) {
        req := ActivateFeatureRequest{
            FeatureID: "advanced_sso",
            PaymentID: "test-payment",
            Duration:  "monthly",
        }
        body, _ := json.Marshal(req)

        r := httptest.NewRequest("POST", "/api/enterprise/features/activate", bytes.NewBuffer(body))
        r = r.WithContext(context.WithValue(r.Context(), "organization_id", "test-org"))
        w := httptest.NewRecorder()

        featureService.On("ActivateFeature", r.Context(), "test-org", req.FeatureID, req.PaymentID, 30*24*time.Hour).Return(nil)

        handler.ActivateFeature(w, r)

        assert.Equal(t, http.StatusOK, w.Code)
        featureService.AssertExpectations(t)
    })

    t.Run("ActivateBundle", func(t *testing.T) {
        req := ActivateBundleRequest{
            BundleID:  "enterprise",
            PaymentID: "test-payment",
            Duration:  "yearly",
        }
        body, _ := json.Marshal(req)

        r := httptest.NewRequest("POST", "/api/enterprise/bundles/activate", bytes.NewBuffer(body))
        r = r.WithContext(context.WithValue(r.Context(), "organization_id", "test-org"))
        w := httptest.NewRecorder()

        featureService.On("ActivateBundle", r.Context(), "test-org", req.BundleID, req.PaymentID, 365*24*time.Hour).Return(nil)

        handler.ActivateBundle(w, r)

        assert.Equal(t, http.StatusOK, w.Code)
        featureService.AssertExpectations(t)
    })

    t.Run("GetActiveFeatures", func(t *testing.T) {
        r := httptest.NewRequest("GET", "/api/enterprise/features", nil)
        r = r.WithContext(context.WithValue(r.Context(), "organization_id", "test-org"))
        w := httptest.NewRecorder()

        features := []string{"advanced_sso", "directory_sync"}
        featureService.On("GetActiveFeatures", r.Context(), "test-org").Return(features, nil)

        handler.GetActiveFeatures(w, r)

        assert.Equal(t, http.StatusOK, w.Code)
        var response map[string]interface{}
        json.NewDecoder(w.Body).Decode(&response)
        assert.Equal(t, features, response["features"])
        featureService.AssertExpectations(t)
    })

    t.Run("DeactivateFeature", func(t *testing.T) {
        r := httptest.NewRequest("POST", "/api/enterprise/features/deactivate?feature_id=advanced_sso", nil)
        r = r.WithContext(context.WithValue(r.Context(), "organization_id", "test-org"))
        w := httptest.NewRecorder()


        featureService.On("DeactivateFeature", r.Context(), "test-org", "advanced_sso").Return(nil)

        handler.DeactivateFeature(w, r)

        assert.Equal(t, http.StatusOK, w.Code)
        featureService.AssertExpectations(t)
    })
}
