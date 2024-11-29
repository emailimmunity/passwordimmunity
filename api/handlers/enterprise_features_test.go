package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/emailimmunity/passwordimmunity/services/enterprise"
	"github.com/shopspring/decimal"
)

type MockFeatureManager struct {
	mock.Mock
}

func (m *MockFeatureManager) IsFeatureEnabled(ctx context.Context, orgID, featureID string) bool {
	args := m.Called(ctx, orgID, featureID)
	return args.Bool(0)
}

func (m *MockFeatureManager) GetActiveFeatures(ctx context.Context, orgID string) ([]string, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]string), args.Error(1)
}

func TestGetFeatureStatus(t *testing.T) {
	mockManager := &MockFeatureManager{}
	mockLogger := &MockLogger{}
	handler := NewEnterpriseFeatureHandler(mockManager, mockLogger)

	tests := []struct {
		name           string
		orgID          string
		featureID      string
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "feature enabled",
			orgID:     "org1",
			featureID: "advanced_sso",
			setupMocks: func() {
				mockManager.On("IsFeatureEnabled", mock.Anything, "org1", "advanced_sso").Return(true)
				mockManager.On("GetActiveFeatures", mock.Anything, "org1").Return([]string{"advanced_sso"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"featureId":    "advanced_sso",
				"isEnabled":    true,
				"activeBundle": []string{"advanced_sso"},
			},
		},
		{
			name:      "feature disabled",
			orgID:     "org1",
			featureID: "custom_roles",
			setupMocks: func() {
				mockManager.On("IsFeatureEnabled", mock.Anything, "org1", "custom_roles").Return(false)
				mockManager.On("GetActiveFeatures", mock.Anything, "org1").Return([]string{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"featureId":    "custom_roles",
				"isEnabled":    false,
				"activeBundle": []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest("GET", "/api/admin/enterprise/features/status?featureId="+tt.featureID, nil)
			req = req.WithContext(context.WithValue(req.Context(), "organizationID", tt.orgID))
			rr := httptest.NewRecorder()

			handler.GetFeatureStatus(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestListFeatures(t *testing.T) {
	mockManager := &MockFeatureManager{}
	mockLogger := &MockLogger{}
	handler := NewEnterpriseFeatureHandler(mockManager, mockLogger)

	tests := []struct {
		name           string
		orgID          string
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "successful features list",
			orgID: "org1",
			setupMocks: func() {
				mockManager.On("GetActiveFeatures", mock.Anything, "org1").Return(
					[]string{"advanced_sso", "custom_roles"}, nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"features": []string{"advanced_sso", "custom_roles"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest("GET", "/api/admin/enterprise/features", nil)
			req = req.WithContext(context.WithValue(req.Context(), "organizationID", tt.orgID))
			rr := httptest.NewRecorder()

			handler.ListFeatures(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestActivateFeature(t *testing.T) {
	mockManager := &MockFeatureManager{}
	mockLogger := &MockLogger{}
	handler := NewEnterpriseFeatureHandler(mockManager, mockLogger)

	tests := []struct {
		name           string
		request        ActivateFeatureRequest
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "successful tier activation",
			request: ActivateFeatureRequest{
				TierID:    "enterprise",
				Currency:  "USD",
				Amount:    "199.99",
				IsYearly:  true,
				PaymentID: "pay_123",
			},
			setupMocks: func() {
				mockManager.On("ActivateFeature", mock.Anything, mock.MatchedBy(func(activation enterprise.FeatureActivation) bool {
					return activation.TierID == "enterprise" &&
						activation.Currency == "USD" &&
						activation.Amount.Equal(decimal.NewFromString("199.99").Val) &&
						activation.IsYearly
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid amount",
			request: ActivateFeatureRequest{
				TierID:    "enterprise",
				Currency:  "USD",
				Amount:    "invalid",
				IsYearly:  true,
				PaymentID: "pay_123",
			},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "activation error",
			request: ActivateFeatureRequest{
				TierID:    "enterprise",
				Currency:  "USD",
				Amount:    "199.99",
				IsYearly:  true,
				PaymentID: "pay_123",
			},
			setupMocks: func() {
				mockManager.On("ActivateFeature", mock.Anything, mock.Anything).Return(
					fmt.Errorf("activation failed"))
				mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/admin/enterprise/features/activate", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), "organizationID", "org1"))
			rr := httptest.NewRecorder()

			handler.ActivateFeature(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockManager.AssertExpectations(t)
		})
	}
}
