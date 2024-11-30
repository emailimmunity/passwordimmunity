package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/emailimmunity/passwordimmunity/db/models"
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockLicenseService struct {
    mock.Mock
}

func (m *mockLicenseService) GetActiveLicense(ctx context.Context, orgID uuid.UUID) (*models.License, error) {
    args := m.Called(ctx, orgID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.License), args.Error(1)
}

type mockFeatureFlagService struct {
    mock.Mock
}

func (m *mockFeatureFlagService) IsFeatureEnabled(ctx context.Context, orgID uuid.UUID, feature string) (bool, error) {
    args := m.Called(ctx, orgID, feature)
    return args.Bool(0), args.Error(1)
}

func (m *mockFeatureFlagService) GetAvailableFeatures(ctx context.Context, orgID uuid.UUID) ([]string, error) {
    args := m.Called(ctx, orgID)
    return args.Get(0).([]string), args.Error(1)
}

func TestGetLicense(t *testing.T) {
    mockLS := new(mockLicenseService)
    mockFS := new(mockFeatureFlagService)
    handler := NewLicenseHandler(mockLS, mockFS)

    orgID := uuid.New()
    ctx := context.WithValue(context.Background(), "organization_id", orgID)

    license := &models.License{
        ID:        uuid.New(),
        Type:      "enterprise",
        Status:    "active",
        Features:  []string{"sso", "api_access"},
    }

    mockLS.On("GetActiveLicense", ctx, orgID).Return(license, nil)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/", nil)
    r = r.WithContext(ctx)

    handler.GetLicense(w, r)

    assert.Equal(t, http.StatusOK, w.Code)

    var response models.License
    json.NewDecoder(w.Body).Decode(&response)
    assert.Equal(t, license.Type, response.Type)
    assert.Equal(t, license.Features, response.Features)
}

func TestGetFeatures(t *testing.T) {
    mockLS := new(mockLicenseService)
    mockFS := new(mockFeatureFlagService)
    handler := NewLicenseHandler(mockLS, mockFS)

    orgID := uuid.New()
    ctx := context.WithValue(context.Background(), "organization_id", orgID)

    features := []string{"sso", "api_access", "advanced_reporting"}
    mockFS.On("GetAvailableFeatures", ctx, orgID).Return(features, nil)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/features", nil)
    r = r.WithContext(ctx)

    handler.GetFeatures(w, r)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string][]string
    json.NewDecoder(w.Body).Decode(&response)
    assert.Equal(t, features, response["features"])
}

func TestCheckFeature(t *testing.T) {
    mockLS := new(mockLicenseService)
    mockFS := new(mockFeatureFlagService)
    handler := NewLicenseHandler(mockLS, mockFS)

    orgID := uuid.New()
    ctx := context.WithValue(context.Background(), "organization_id", orgID)

    mockFS.On("IsFeatureEnabled", ctx, orgID, "sso").Return(true, nil)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/check/sso", nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("feature", "sso")
    r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

    handler.CheckFeature(w, r)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]bool
    json.NewDecoder(w.Body).Decode(&response)
    assert.True(t, response["enabled"])
}
