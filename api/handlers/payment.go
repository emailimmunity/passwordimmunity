package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/services/payment"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type PaymentHandler struct {
	paymentService *payment.Service
	licenseService *licensing.Service
}

func NewPaymentHandler(ps *payment.Service) *PaymentHandler {
	return &PaymentHandler{
		paymentService: ps,
		licenseService: licensing.GetService(),
	}
}

// HandleWebhook processes Mollie payment webhooks
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	paymentID := r.FormValue("id")
	if paymentID == "" {
		http.Error(w, "Missing payment ID", http.StatusBadRequest)
		return
	}

	payment, err := h.paymentService.VerifyPayment(r.Context(), paymentID)
	if err != nil {
		http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
		return
	}

	if payment.Status == "paid" {
		err = h.licenseService.ActivateFeature(
			payment.Metadata.OrgID,
			payment.Metadata.FeatureID,
		)
		if err != nil {
			http.Error(w, "Failed to activate feature", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// ListPayments returns all payments for an organization
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(string)

	payments, err := h.paymentService.ListPayments(r.Context(), orgID)
	if err != nil {
		http.Error(w, "Failed to list payments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// GetPayment returns details of a specific payment
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentID := vars["id"]
	orgID := r.Context().Value("organization_id").(string)

	payment, err := h.paymentService.GetPayment(r.Context(), paymentID, orgID)
	if err != nil {
		http.Error(w, "Failed to get payment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// RetryPayment attempts to retry a failed payment
func (h *PaymentHandler) RetryPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentID := vars["id"]
	orgID := r.Context().Value("organization_id").(string)

	newPayment, err := h.paymentService.RetryPayment(r.Context(), paymentID, orgID)
	if err != nil {
		http.Error(w, "Failed to retry payment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newPayment)
}
