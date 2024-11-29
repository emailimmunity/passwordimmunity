package featureflag

import (
	"context"
	"github.com/emailimmunity/passwordimmunity/services/licensing"
)

// EnterpriseFeatureFlag represents a feature that requires enterprise licensing
type EnterpriseFeatureFlag struct {
	Feature     licensing.EnterpriseFeature
	Name        string
	Description string
}

var (
	// Enterprise features
	AdvancedSSO = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureAdvancedSSO,
		Name:        "Advanced SSO",
		Description: "SAML and OIDC integration capabilities",
	}
	CustomRoles = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureCustomRoles,
		Name:        "Custom Roles",
		Description: "Advanced role management and customization",
	}
	AdvancedReporting = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureAdvancedReporting,
		Name:        "Advanced Reporting",
		Description: "Detailed audit logs and reporting capabilities",
	}
	APIAccess = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureAPIAccess,
		Name:        "API Access",
		Description: "Enterprise API access for automation",
	}
	AdvancedGroups = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureAdvancedGroups,
		Name:        "Advanced Groups",
		Description: "Advanced group management capabilities",
	}
	MultiTenant = EnterpriseFeatureFlag{
		Feature:     licensing.FeatureMultiTenant,
		Name:        "Multi-Tenant",
		Description: "Multi-tenant system capabilities",
	}
	PaymentProcessing = EnterpriseFeatureFlag{
		Feature:     licensing.FeaturePaymentProcessing,
		Name:        "Payment Processing",
		Description: "Mollie payment integration",
	}
)

// FeatureManager handles feature flag checks
type FeatureManager struct {
	licenseVerifier licensing.LicenseVerifier
}

func NewFeatureManager(verifier licensing.LicenseVerifier) *FeatureManager {
	return &FeatureManager{
		licenseVerifier: verifier,
	}
}

// IsEnabled checks if a feature is enabled for the current license
func (m *FeatureManager) IsEnabled(ctx context.Context, feature EnterpriseFeatureFlag) bool {
	license := ctx.Value("license").(*licensing.License)
	return m.licenseVerifier.HasFeature(license, feature.Feature)
}

// GetEnabledFeatures returns all enabled enterprise features for the current license
func (m *FeatureManager) GetEnabledFeatures(ctx context.Context) []EnterpriseFeatureFlag {
	license := ctx.Value("license").(*licensing.License)
	var enabled []EnterpriseFeatureFlag

	features := []EnterpriseFeatureFlag{
		AdvancedSSO,
		CustomRoles,
		AdvancedReporting,
		APIAccess,
		AdvancedGroups,
		MultiTenant,
		PaymentProcessing,
	}

	for _, feature := range features {
		if m.licenseVerifier.HasFeature(license, feature.Feature) {
			enabled = append(enabled, feature)
		}
	}

	return enabled
}
