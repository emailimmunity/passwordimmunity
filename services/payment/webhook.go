package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type WebhookHandler struct {
	service    *PaymentService
	processing sync.Map
}

type WebhookRequest struct {
	ID string `json:"id"`
}

type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewWebhookHandler(service *PaymentService) *WebhookHandler {
	return &WebhookHandler{
		service: service,
	}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendResponse(w, false, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendResponse(w, false, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Prevent duplicate processing
	if _, loaded := h.processing.LoadOrStore(req.ID, true); loaded {
		h.sendResponse(w, true, "Payment already being processed", http.StatusOK)
		return
	}
	defer h.processing.Delete(req.ID)

	// Process webhook asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), webhookTimeout)
		defer cancel()

		if err := h.service.VerifyAndActivate(ctx, req.ID); err != nil {
			log.Printf("Webhook processing failed for payment %s: %v\n", req.ID, err)
			return
		}
		log.Printf("Successfully processed payment %s and activated features\n", req.ID)
	}()

	h.sendResponse(w, true, "Webhook received and being processed", http.StatusOK)
}

func (h *WebhookHandler) sendResponse(w http.ResponseWriter, success bool, message string, status int) {
	resp := WebhookResponse{
		Success: success,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

const webhookTimeout = 30 * time.Second
