package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/emailimmunity/passwordimmunity/config"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

type FeatureStatusResponse struct {
	FeatureID    string `json:"feature_id"`
	Active       bool   `json:"active"`
	InGracePeriod bool  `json:"in_grace_period"`
	GracePeriod  int    `json:"grace_period_days,omitempty"`
	Price        float64 `json:"price,omitempty"`
}

type BundleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Features    []string  `json:"features"`
	Price       float64   `json:"price"`
}

// HandleFeatureStatus returns the activation status of a specific feature
func HandleFeatureStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	featureID := vars["feature_id"]
	orgID := r.Context().Value("organization_id").(string)

	licenseService := licensing.GetService()

	response := FeatureStatusResponse{
		FeatureID:     featureID,
		Active:        licenseService.HasFeatureAccess(orgID, featureID),
		InGracePeriod: licenseService.IsInGracePeriod(orgID, featureID),
		GracePeriod:   config.GetFeatureGracePeriod(featureID),
		Price:         config.GetFeaturePrice(featureID),
	}

	respondWithJSON(w, response)
}

// HandleListBundles returns all available feature bundles
func HandleListBundles(w http.ResponseWriter, r *http.Request) {
	bundles := config.GetAllBundles()
	response := make([]BundleResponse, 0, len(bundles))

	for _, bundle := range bundles {
		response = append(response, BundleResponse{
			ID:          bundle.ID,
			Name:        bundle.Name,
			Description: bundle.Description,
			Features:    bundle.Features,
			Price:       bundle.Price,
		})
	}

	respondWithJSON(w, response)
}

// HandleListAvailableFeatures returns all available enterprise features
func HandleListAvailableFeatures(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("organization_id").(string)
	licenseService := licensing.GetService()

	features := config.GetAllFeatures()
	response := make([]FeatureStatusResponse, 0, len(features))

	for _, feature := range features {
		response = append(response, FeatureStatusResponse{
			FeatureID:     feature.ID,
			Active:        licenseService.HasFeatureAccess(orgID, feature.ID),
			InGracePeriod: licenseService.IsInGracePeriod(orgID, feature.ID),
			GracePeriod:   config.GetFeatureGracePeriod(feature.ID),
			Price:         config.GetFeaturePrice(feature.ID),
		})
	}

	respondWithJSON(w, response)
}
