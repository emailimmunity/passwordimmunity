package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/emailimmunity/passwordimmunity/services/featureflag"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type EnterpriseActivationHandler struct {
	featureManager *featureflag.FeatureManager
	paymentService *payment.Service
}

func NewEnterpriseActivationHandler(fm *featureflag.FeatureManager, ps *payment.Service) *EnterpriseActivationHandler {
	return &EnterpriseActivationHandler{
		featureManager: fm,
		paymentService: ps,
	}
}

// InitiateFeatureActivation starts the payment process for an enterprise feature
func (h *EnterpriseActivationHandler) InitiateFeatureActivation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Feature string `json:"feature"`
		Plan    string `json:"plan"` // "monthly" or "yearly"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create payment for feature activation
	payment, err := h.paymentService.CreatePayment(r.Context(), &payment.CreatePaymentRequest{
		Feature: req.Feature,
		Plan:    req.Plan,
	})
	if err != nil {
		http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		return
	}

	response := struct {
		PaymentURL string `json:"paymentUrl"`
		PaymentID  string `json:"paymentId"`
	}{
		PaymentURL: payment.CheckoutURL,
		PaymentID:  payment.ID,
	}

	json.NewEncoder(w).Encode(response)
}

// HandlePaymentWebhook processes Mollie payment webhook callbacks
func (h *EnterpriseActivationHandler) HandlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	paymentID := r.FormValue("id")
	if paymentID == "" {
		http.Error(w, "Missing payment ID", http.StatusBadRequest)
		return
	}

	// Verify payment status with Mollie
	status, err := h.paymentService.VerifyPayment(r.Context(), paymentID)
	if err != nil {
		http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
		return
	}

	if status.IsPaid {
		// Activate the enterprise feature
		err = h.paymentService.ActivateFeature(r.Context(), status.FeatureID, status.OrganizationID)
		if err != nil {
			http.Error(w, "Failed to activate feature", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GetActivationStatus returns the current activation status of enterprise features
func (h *EnterpriseActivationHandler) GetActivationStatus(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organizationID").(string)

	status, err := h.paymentService.GetFeatureStatus(r.Context(), orgID)
	if err != nil {
		http.Error(w, "Failed to get feature status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}
