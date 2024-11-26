package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type PaymentHandler struct {
	paymentService payment.PaymentService
}

func NewPaymentHandler(paymentService payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req payment.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payment, err := h.paymentService.CreatePayment(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.URL.Query().Get("id")
	if paymentID == "" {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}

	payment, err := h.paymentService.GetPayment(r.Context(), paymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payment == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var req payment.WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.paymentService.HandleWebhook(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.URL.Query().Get("id")
	if paymentID == "" {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}

	if err := h.paymentService.CancelPayment(r.Context(), paymentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
