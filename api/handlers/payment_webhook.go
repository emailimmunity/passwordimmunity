package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type WebhookRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// HandlePaymentWebhook processes payment status webhooks from Mollie
func HandlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	paymentService := payment.GetService()
	activationService := enterprise.GetActivationService()

	// Verify payment status
	paymentResp, err := paymentService.VerifyPayment(r.Context(), req.ID)
	if err != nil {
		http.Error(w, "Error verifying payment", http.StatusInternalServerError)
		return
	}

	if paymentResp.Status != "paid" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract metadata
	metadata := paymentResp.Metadata
	orgID := metadata.OrganizationID
	features := metadata.Features
	bundles := metadata.Bundles
	duration := metadata.Duration

	// Convert duration string to time.Duration
	billingPeriod, err := config.ParseBillingPeriod(duration)
	if err != nil {
		http.Error(w, "Invalid billing period", http.StatusBadRequest)
		return
	}

	// Activate features or bundles
	if len(bundles) > 0 {
		for _, bundleID := range bundles {
			if err := activationService.ActivateBundle(r.Context(), orgID, bundleID, billingPeriod); err != nil {
				http.Error(w, "Error activating bundle", http.StatusInternalServerError)
				return
			}
		}
	}

	if len(features) > 0 {
		for _, featureID := range features {
			if err := activationService.ActivateFeature(r.Context(), orgID, featureID, billingPeriod); err != nil {
				http.Error(w, "Error activating feature", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
