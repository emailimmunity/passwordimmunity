package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type mockPaymentService struct {
	payments map[string]*payment.Payment
}

func (m *mockPaymentService) GetPayment(ctx context.Context, id string) (*payment.Payment, error) {
	if p, ok := m.payments[id]; ok {
		return p, nil
	}
	return nil, nil
}

func TestHandlePaymentWebhook(t *testing.T) {
	tests := []struct {
		name       string
		webhookReq WebhookRequest
		payment    *payment.Payment
		wantStatus int
		wantError  bool
	}{
		{
			name: "successful payment with currency",
			webhookReq: WebhookRequest{
				ID:     "tr_123",
				Status: "paid",
			},
			payment: &payment.Payment{
				ID:     "tr_123",
				Status: "paid",
				Amount: 99.99,
				Currency: "EUR",
				Metadata: map[string]string{
					"org_id":     "org1",
					"feature_id": "advanced_sso",
					"currency":   "EUR",
				},
				ExpiresAt: time.Now().AddDate(1, 0, 0),
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "payment with incorrect amount",
			webhookReq: WebhookRequest{
				ID:     "tr_125",
				Status: "paid",
			},
			payment: &payment.Payment{
				ID:     "tr_125",
				Status: "paid",
				Amount: 50.00,
				Currency: "EUR",
				Metadata: map[string]string{
					"org_id":     "org1",
					"feature_id": "advanced_sso",
					"currency":   "EUR",
				},
				ExpiresAt: time.Now().AddDate(1, 0, 0),
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "pending payment",
			webhookReq: WebhookRequest{
				ID:     "tr_124",
				Status: "pending",
			},
			payment: &payment.Payment{
				ID:     "tr_124",
				Status: "pending",
			},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayment := &mockPaymentService{
				payments: map[string]*payment.Payment{
					tt.payment.ID: tt.payment,
				},
			}
			payment.SetService(mockPayment)

			body, _ := json.Marshal(tt.webhookReq)
			req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			HandlePaymentWebhook(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("HandlePaymentWebhook() status = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
