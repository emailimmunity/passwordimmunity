package enterprise

import (
	"context"
	"testing"
	"time"
)

type mockPaymentService struct {
	payments map[string]*Payment
}

type mockLicenseService struct {
	licenses map[string]*License
}

type mockFeatureService struct {
	features map[string]*Feature
	active   map[string]bool
}

func TestActivationService_ActivateFeature(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	ps := &mockPaymentService{
		payments: make(map[string]*Payment),
	}
	ls := &mockLicenseService{
		licenses: make(map[string]*License),
	}
	fs := &mockFeatureService{
		features: map[string]*Feature{
			"test_feature": {
				ID:          "test_feature",
				Name:        "Test Feature",
				Description: "Test enterprise feature",
				Price: PriceInfo{
					Monthly: 49.99,
					Yearly:  499.99,
				},
				Tier: "enterprise",
			},
		},
		active: make(map[string]bool),
	}

	service := NewActivationService(ps, ls, fs)

	tests := []struct {
		name    string
		req     ActivationRequest
		wantErr bool
	}{
		{
			name: "successful activation request",
			req: ActivationRequest{
				OrganizationID: "org1",
				FeatureID:      "test_feature",
				PlanType:       "monthly",
			},
			wantErr: false,
		},
		{
			name: "non-existent feature",
			req: ActivationRequest{
				OrganizationID: "org1",
				FeatureID:      "invalid_feature",
				PlanType:       "monthly",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := service.ActivateFeature(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ActivateFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && status == nil {
				t.Error("ActivateFeature() expected status, got nil")
			}
		})
	}
}

func TestActivationService_HandlePaymentWebhook(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	testPayment := &Payment{
		ID:             "payment1",
		OrganizationID: "org1",
		FeatureID:      "test_feature",
		Status:         "paid",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	ps := &mockPaymentService{
		payments: map[string]*Payment{
			"payment1": testPayment,
		},
	}
	ls := &mockLicenseService{
		licenses: make(map[string]*License),
	}
	fs := &mockFeatureService{
		features: map[string]*Feature{
			"test_feature": {
				ID:   "test_feature",
				Tier: "enterprise",
			},
		},
		active: make(map[string]bool),
	}

	service := NewActivationService(ps, ls, fs)

	tests := []struct {
		name    string
		webhook WebhookData
		wantErr bool
	}{
		{
			name: "successful payment confirmation",
			webhook: WebhookData{
				PaymentID: "payment1",
				Status:    "paid",
			},
			wantErr: false,
		},
		{
			name: "invalid payment ID",
			webhook: WebhookData{
				PaymentID: "invalid_payment",
				Status:    "paid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.HandlePaymentWebhook(ctx, tt.webhook)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandlePaymentWebhook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Mock implementations
func (m *mockPaymentService) CreatePayment(ctx context.Context, req PaymentRequest) (*Payment, error) {
	payment := &Payment{
		ID:             "test_payment",
		OrganizationID: req.OrganizationID,
		FeatureID:      req.FeatureID,
		Amount:         req.Amount,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}
	m.payments[payment.ID] = payment
	return payment, nil
}

func (m *mockPaymentService) VerifyPayment(ctx context.Context, paymentID string) (*Payment, error) {
	if payment, ok := m.payments[paymentID]; ok {
		return payment, nil
	}
	return nil, nil
}

func (m *mockPaymentService) GetPendingPayment(ctx context.Context, orgID, featureID string) (*Payment, error) {
	return nil, nil
}

func (m *mockLicenseService) GenerateLicense(ctx context.Context, req LicenseRequest) (*License, error) {
	license := &License{
		ID:             "test_license",
		OrganizationID: req.OrganizationID,
		FeatureID:      req.FeatureID,
		Key:            "test_key",
		ExpiresAt:      req.ExpiresAt,
	}
	m.licenses[license.ID] = license
	return license, nil
}

func (m *mockLicenseService) VerifyLicense(ctx context.Context, license *License) (bool, error) {
	return true, nil
}

func (m *mockLicenseService) RevokeLicense(ctx context.Context, licenseID string) error {
	delete(m.licenses, licenseID)
	return nil
}

func (m *mockFeatureService) GetFeature(featureID string) (*Feature, error) {
	if feature, ok := m.features[featureID]; ok {
		return feature, nil
	}
	return nil, nil
}

func (m *mockFeatureService) IsFeatureActive(ctx context.Context, orgID, featureID string) (bool, error) {
	key := orgID + ":" + featureID
	return m.active[key], nil
}

func (m *mockFeatureService) ActivateFeature(ctx context.Context, orgID, featureID string, license *License) error {
	key := orgID + ":" + featureID
	m.active[key] = true
	return nil
}

func (m *mockFeatureService) GetFeatureExpiry(ctx context.Context, orgID, featureID string) (time.Time, error) {
	return time.Now().Add(24 * time.Hour), nil
}
