package payment

import (
    "context"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockPaymentService struct {
    mock.Mock
}

func (m *mockPaymentService) VerifyAndActivate(ctx context.Context, paymentID string) error {
    args := m.Called(ctx, paymentID)
    return args.Error(0)
}

type mockLogger struct {
    mock.Mock
}

func (m *mockLogger) Error(msg string, keyvals ...interface{}) {
    m.Called(msg, keyvals)
}

func TestWebhookHandler_HandleWebhook(t *testing.T) {
    t.Run("successful payment processing", func(t *testing.T) {
        mockService := &mockPaymentService{}
        handler := &WebhookHandler{
            service: mockService,
            logger:  &mockLogger{},
        }

        mockService.On("VerifyAndActivate", mock.Anything, "test_payment").Return(nil)

        form := url.Values{}
        form.Add("id", "test_payment")
        req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(form.Encode()))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        w := httptest.NewRecorder()
        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        mockService.AssertExpectations(t)
    })

    t.Run("missing payment ID", func(t *testing.T) {
        mockService := &mockPaymentService{}
        mockLog := &mockLogger{}
        handler := &WebhookHandler{
            service: mockService,
            logger:  mockLog,
        }

        mockLog.On("Error", "webhook error", mock.Anything).Return()

        req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
        w := httptest.NewRecorder()
        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        mockLog.AssertExpectations(t)
    })

    t.Run("payment verification failure", func(t *testing.T) {
        mockService := &mockPaymentService{}
        mockLog := &mockLogger{}
        handler := &WebhookHandler{
            service: mockService,
            logger:  mockLog,
        }

        mockService.On("VerifyAndActivate", mock.Anything, "test_payment").
            Return(fmt.Errorf("payment verification failed"))
        mockLog.On("Error", "webhook error", mock.Anything).Return()

        form := url.Values{}
        form.Add("id", "test_payment")
        req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(form.Encode()))
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        w := httptest.NewRecorder()
        handler.HandleWebhook(w, req)

        assert.Equal(t, http.StatusInternalServerError, w.Code)
        mockService.AssertExpectations(t)
        mockLog.AssertExpectations(t)
    })
}
