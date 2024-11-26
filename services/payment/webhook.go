package payment

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
)

type WebhookHandler struct {
    service *Service
    client  MollieClient
}

func NewWebhookHandler(service *Service, client MollieClient) *WebhookHandler {
    return &WebhookHandler{
        service: service,
        client:  client,
    }
}

type WebhookPayload struct {
    ID string `json:"id"`
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var payload WebhookPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    payment, err := h.client.GetPayment(payload.ID)
    if err != nil {
        http.Error(w, "Failed to get payment details", http.StatusInternalServerError)
        return
    }

    if err := h.processPaymentUpdate(r.Context(), payment); err != nil {
        http.Error(w, "Failed to process payment update", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) processPaymentUpdate(ctx context.Context, payment *Payment) error {
    switch payment.Status {
    case "paid":
        return h.service.ActivateLicense(ctx, payment.ID)
    case "failed", "canceled", "expired":
        return h.service.CancelLicense(ctx, payment.ID)
    default:
        return errors.New(fmt.Sprintf("unhandled payment status: %s", payment.Status))
    }
}
