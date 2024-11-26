package handlers

import (
    "bytes"
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

type mockPaymentService struct {
    mock.Mock
}

func (m *mockPaymentService) CreatePayment(ctx context.Context, orgID uuid.UUID, licenseType, period string) (*models.Payment, error) {
    args := m.Called(ctx, orgID, licenseType, period)
    return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *mockPaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *mockPaymentService) HandlePaymentWebhook(ctx context.Context, providerID, status string) error {
    args := m.Called(ctx, providerID, status)
    return args.Error(0)
}

func TestCreatePayment(t *testing.T) {
    mockPS := new(mockPaymentService)
    mockLS := new(mockLicenseService)
    handler := NewPaymentHandler(mockPS, mockLS)

    orgID := uuid.New()
    ctx := context.WithValue(context.Background(), "organization_id", orgID)

    payment := &models.Payment{
        ID:          uuid.New(),
        ProviderID:  "tr_test123",
        Amount:      "999.00",
        LicenseType: "enterprise",
        Period:      "yearly",
    }

    mockPS.On("CreatePayment", ctx, orgID, "enterprise", "yearly").Return(payment, nil)

    req := CreatePaymentRequest{
        LicenseType: "enterprise",
        Period:      "yearly",
    }
    body, _ := json.Marshal(req)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
    r = r.WithContext(ctx)

    handler.CreatePayment(w, r)

    assert.Equal(t, http.StatusOK, w.Code)

    var response models.Payment
    json.NewDecoder(w.Body).Decode(&response)
    assert.Equal(t, payment.ID, response.ID)
}

func TestHandleWebhook(t *testing.T) {
    mockPS := new(mockPaymentService)
    mockLS := new(mockLicenseService)
    handler := NewPaymentHandler(mockPS, mockLS)

    mockPS.On("HandlePaymentWebhook", mock.Anything, "tr_test123", "paid").Return(nil)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("POST", "/webhook", nil)
    r.Form = make(map[string][]string)
    r.Form.Add("id", "tr_test123")
    r.Form.Add("status", "paid")

    handler.HandleWebhook(w, r)

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPaymentStatus(t *testing.T) {
    mockPS := new(mockPaymentService)
    mockLS := new(mockLicenseService)
    handler := NewPaymentHandler(mockPS, mockLS)

    paymentID := uuid.New()
    payment := &models.Payment{
        ID:     paymentID,
        Status: "paid",
    }

    mockPS.On("GetPayment", mock.Anything, paymentID).Return(payment, nil)

    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/status/"+paymentID.String(), nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("id", paymentID.String())
    r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

    handler.GetPaymentStatus(w, r)

    assert.Equal(t, http.StatusOK, w.Code)

    var response models.Payment
    json.NewDecoder(w.Body).Decode(&response)
    assert.Equal(t, payment.Status, response.Status)
}
