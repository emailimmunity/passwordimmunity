package payment

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/emailimmunity/passwordimmunity/services/logger"
)

type WebhookHandler struct {
    service *PaymentService
    logger  logger.Logger
}

func NewWebhookHandler(service *PaymentService) *WebhookHandler {
    return &WebhookHandler{
        service: service,
        logger:  logger.GetLogger(),
    }
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    // Extract payment ID from request
    if err := r.ParseForm(); err != nil {
        h.handleError(w, fmt.Errorf("failed to parse form: %w", err), http.StatusBadRequest)
        return
    }

    paymentID := r.PostForm.Get("id")
    if paymentID == "" {
        h.handleError(w, fmt.Errorf("missing payment ID"), http.StatusBadRequest)
        return
    }

    // Process the payment
    if err := h.service.VerifyAndActivate(ctx, paymentID); err != nil {
        h.handleError(w, fmt.Errorf("payment processing failed: %w", err), http.StatusInternalServerError)
        return
    }

    // Return success
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
    })
}

func (h *WebhookHandler) handleError(w http.ResponseWriter, err error, status int) {
    h.logger.Error("webhook error", "error", err)
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{
        "error": err.Error(),
    })
}
