package middleware

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

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

func TestFeatureCheckMiddleware(t *testing.T) {
    featureService := new(mockFeatureService)
    middleware := NewFeatureCheckMiddleware(featureService)

    t.Run("RequireFeature - Feature Active", func(t *testing.T) {
        handler := middleware.RequireFeature("advanced_sso")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
        }))

        req := httptest.NewRequest("GET", "/", nil)
        req = req.WithContext(context.WithValue(req.Context(), "organization_id", "test-org"))
        rr := httptest.NewRecorder()

        featureService.On("IsFeatureActive", req.Context(), "test-org", "advanced_sso").Return(true, nil)

        handler.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        featureService.AssertExpectations(t)
    })

    t.Run("RequireFeature - Feature Inactive", func(t *testing.T) {
        handler := middleware.RequireFeature("advanced_sso")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
        }))

        req := httptest.NewRequest("GET", "/", nil)
        req = req.WithContext(context.WithValue(req.Context(), "organization_id", "test-org"))
        rr := httptest.NewRecorder()

        featureService.On("IsFeatureActive", req.Context(), "test-org", "advanced_sso").Return(false, nil)

        handler.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusForbidden, rr.Code)
        featureService.AssertExpectations(t)
    })

    t.Run("RequireBundle - Bundle Active", func(t *testing.T) {
        handler := middleware.RequireBundle("enterprise")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
        }))

        req := httptest.NewRequest("GET", "/", nil)
        req = req.WithContext(context.WithValue(req.Context(), "organization_id", "test-org"))
        rr := httptest.NewRecorder()

        featureService.On("IsBundleActive", req.Context(), "test-org", "enterprise").Return(true, nil)

        handler.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        featureService.AssertExpectations(t)
    })

    t.Run("RequireBundle - Bundle Inactive", func(t *testing.T) {
        handler := middleware.RequireBundle("enterprise")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
        }))

        req := httptest.NewRequest("GET", "/", nil)
        req = req.WithContext(context.WithValue(req.Context(), "organization_id", "test-org"))
        rr := httptest.NewRecorder()

        featureService.On("IsBundleActive", req.Context(), "test-org", "enterprise").Return(false, nil)


        handler.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusForbidden, rr.Code)
        featureService.AssertExpectations(t)
    })
}
