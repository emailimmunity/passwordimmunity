package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFeatureManager struct {
	mock.Mock
}

func (m *MockFeatureManager) ActivateFeature(ctx context.Context, activation featureflag.FeatureActivation) error {
	args := m.Called(ctx, activation)
	return args.Error(0)
}

func (m *MockFeatureManager) GetFeatureStatus(ctx context.Context) (map[string]bool, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]bool), args.Error(1)
}

func TestEnterpriseFeatureHandler_ActivateFeature(t *testing.T) {
	tests := []struct {
		name           string
		request        ActivateFeatureRequest
		setupMock      func(*MockFeatureManager)
		expectedStatus int
	}{
		{
			name: "successful tier activation",
			request: ActivateFeatureRequest{
				TierID: "enterprise",
			},
			setupMock: func(m *MockFeatureManager) {
				m.On("ActivateFeature", mock.Anything, featureflag.FeatureActivation{
					TierID: "enterprise",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful feature activation",
			request: ActivateFeatureRequest{
				FeatureID: "advanced_sso",
			},
			setupMock: func(m *MockFeatureManager) {
				m.On("ActivateFeature", mock.Anything, featureflag.FeatureActivation{
					FeatureID: "advanced_sso",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful bundle activation",
			request: ActivateFeatureRequest{
				BundleID: "security",
			},
			setupMock: func(m *MockFeatureManager) {
				m.On("ActivateFeature", mock.Anything, featureflag.FeatureActivation{
					BundleID: "security",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request - no identifiers",
			request:        ActivateFeatureRequest{},
			setupMock:      func(m *MockFeatureManager) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFM := new(MockFeatureManager)
			tt.setupMock(mockFM)

			handler := NewEnterpriseFeatureHandler(mockFM, nil)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/enterprise/features/activate", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ActivateFeature(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockFM.AssertExpectations(t)
		})
	}
}

func TestEnterpriseFeatureHandler_GetFeatureStatus(t *testing.T) {
	mockFM := new(MockFeatureManager)
	expectedStatus := map[string]bool{
		"advanced_sso": true,
		"multi_tenant": false,
	}

	mockFM.On("GetFeatureStatus", mock.Anything).Return(expectedStatus, nil)

	handler := NewEnterpriseFeatureHandler(mockFM, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/enterprise/features/status", nil)
	rec := httptest.NewRecorder()

	handler.GetFeatureStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]bool
	json.NewDecoder(rec.Body).Decode(&response)
	assert.Equal(t, expectedStatus, response)

	mockFM.AssertExpectations(t)
}
