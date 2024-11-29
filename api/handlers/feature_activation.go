package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
	"github.com/emailimmunity/passwordimmunity/services/payment"
)

type ActivationRequest struct {
	FeatureID string `json:"feature_id"`
	BundleID  string `json:"bundle_id,omitempty"`
	Currency  string `json:"currency,omitempty"`
}

type ActivationResponse struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message,omitempty"`
	Prices       map[string]float64 `json:"prices,omitempty"`
	PaymentURL   string             `json:"payment_url,omitempty"`
	GracePeriod  int               `json:"grace_period_days,omitempty"`
	PaymentID    string             `json:"payment_id,omitempty"`
	Status       string             `json:"status,omitempty"`
	ExpiresAt    string             `json:"expires_at,omitempty"`
}

// HandleFeatureActivation processes feature activation requests
func HandleFeatureActivation(w http.ResponseWriter, r *http.Request) {
	var req ActivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgID := r.Context().Value("organization_id").(string)
	licenseService := licensing.GetService()
	paymentService := payment.GetService()
	featureService := config.GetEnterpriseService()

	// Check if feature/bundle is already active
	isActive, err := featureService.IsFeatureActive(r.Context(), req.FeatureID, orgID)
	if err != nil {
		http.Error(w, "Error checking feature status", http.StatusInternalServerError)
		return
	}
	if isActive {
		respondWithJSON(w, ActivationResponse{
			Success: true,
			Message: "Feature is already active",
		})
		return
	}

	// Calculate price and create payment session
	var price float64
	var itemID string
	var description string
	var currency string

	// Set currency, default to EUR if not specified
	if req.Currency != "" && contains(config.GetSupportedCurrencies(), req.Currency) {
		currency = req.Currency
	} else {
		currency = config.GetDefaultCurrency()
	}

	if req.BundleID != "" {
		price = config.GetBundlePrice(req.BundleID, currency)
		itemID = "bundle_" + req.BundleID
		description = "Enterprise bundle activation: " + req.BundleID
	} else {
		price = config.GetFeaturePrice(req.FeatureID, currency)
		itemID = "feature_" + req.FeatureID
		description = "Enterprise feature activation: " + req.FeatureID
	}

	if price == 0 {
		http.Error(w, "Invalid feature or bundle ID", http.StatusBadRequest)
		return
	}

	// Create payment session with Mollie
	session, err := paymentService.CreateSession(r.Context(), payment.SessionRequest{
		OrganizationID: orgID,
		ItemID:         itemID,
		Amount:         price,
		Currency:       currency,
		Description:    description,
		RedirectURL:    config.GetBaseURL() + "/enterprise/activation/complete",
		WebhookURL:     config.GetBaseURL() + "/api/enterprise/activation/webhook",
		Metadata: map[string]string{
			"feature_id": req.FeatureID,
			"bundle_id":  req.BundleID,
			"org_id":     orgID,
			"currency":   currency,
		},
	})
	if err != nil {
		http.Error(w, "Error creating payment session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, ActivationResponse{
		Success:     false,
		Message:     "Payment required to activate feature",
		Prices:      getPricesMap(req.FeatureID, req.BundleID),
		PaymentURL:  session.PaymentURL,
		PaymentID:   session.ID,
		Status:      "pending",
		GracePeriod: config.GetFeatureGracePeriod(req.FeatureID),
	})
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getPricesMap(featureID, bundleID string) map[string]float64 {
	if bundleID != "" {
		if bundle, exists := config.GetBundleByID(bundleID); exists {
			return bundle.Prices
		}
	} else if featureID != "" {
		if feature, exists := config.GetFeatureByID(featureID); exists {
			return feature.Prices
		}
	}
	return nil
}

	respondWithJSON(w, ActivationResponse{
		Success:     false,
		Message:     "Payment required to activate feature",
		Price:       price,
		PaymentURL:  session.PaymentURL,
		PaymentID:   session.ID,
		Status:      "pending",
		GracePeriod: config.GetFeatureGracePeriod(req.FeatureID),
	})
}

func respondWithJSON(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
