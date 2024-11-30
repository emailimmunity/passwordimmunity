package payment

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

type mockService struct {
    mock.Mock
}

func (m *mockService) ActivateLicense(ctx context.Context, paymentID string) error {
    args := m.Called(ctx, paymentID)
    return args.Error(0)
}

func (m *mockService) CancelLicense(ctx context.Context, paymentID string) error {
    args := m.Called(ctx, paymentID)
    return args.Error(0)
}

func TestWebhookHandler(t *testing.T) {
    mockSvc := new(mockService)
    mockClient := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: http.StatusOK,
                Body: bytes.NewBufferString(`{
                    "id": "tr_test123",
                    "status": "paid"
                }`),
            }, nil
        },
    }

    handler := NewWebhookHandler(mockSvc, mockClient)

    t.Run("successful payment webhook", func(t *testing.T) {
        payload := WebhookPayload{ID: "tr_test123"}
        body, _ := json.Marshal(payload)
        req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
        w := httptest.NewRecorder()

        mockSvc.On("ActivateLicense", mock.Anything, "tr_test123").Return(nil)

        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        mockSvc.AssertExpectations(t)
    })

    t.Run("invalid method", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
        w := httptest.NewRecorder()

        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
    })

    t.Run("invalid payload", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString("invalid json"))
        w := httptest.NewRecorder()

        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}
