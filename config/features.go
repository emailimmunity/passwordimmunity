package config

// FeatureConfig defines the configuration for a feature
type FeatureConfig struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    MinTier     string   `json:"min_tier"` // free, premium, enterprise
    Dependencies []string `json:"dependencies,omitempty"`
}

// Features defines all available features and their requirements
var Features = map[string]FeatureConfig{
    // Free tier features
    "basic_vault": {
        Name:        "Basic Vault",
        Description: "Basic password vault functionality",
        MinTier:     "free",
    },
    "basic_2fa": {
        Name:        "Basic 2FA",
        Description: "Basic two-factor authentication",
        MinTier:     "free",
    },

    // Premium tier features
    "advanced_2fa": {
        Name:        "Advanced 2FA",
        Description: "Advanced two-factor authentication including hardware keys",
        MinTier:     "premium",
    },
    "emergency_access": {
        Name:        "Emergency Access",
        Description: "Emergency access to vaults",
        MinTier:     "premium",
    },
    "priority_support": {
        Name:        "Priority Support",
        Description: "Priority technical support",
        MinTier:     "premium",
    },
    "basic_api_access": {
        Name:        "Basic API Access",
        Description: "Basic API access for automation",
        MinTier:     "premium",
    },
    "basic_reporting": {
        Name:        "Basic Reporting",
        Description: "Basic reporting and analytics",
        MinTier:     "premium",
    },

    // Enterprise tier features
    "sso": {
        Name:        "Single Sign-On",
        Description: "Enterprise SSO integration",
        MinTier:     "enterprise",
    },
    "directory_sync": {
        Name:        "Directory Sync",
        Description: "Enterprise directory synchronization",
        MinTier:     "enterprise",
        Dependencies: []string{"sso"},
    },
    "advanced_reporting": {
        Name:        "Advanced Reporting",
        Description: "Advanced reporting and analytics",
        MinTier:     "enterprise",
    },
    "custom_roles": {
        Name:        "Custom Roles",
        Description: "Custom role management",
        MinTier:     "enterprise",
    },
    "advanced_groups": {
        Name:        "Advanced Groups",
        Description: "Advanced group management",
        MinTier:     "enterprise",
    },
    "multi_tenant": {
        Name:        "Multi-tenant Management",
        Description: "Multi-tenant organization management",
        MinTier:     "enterprise",
    },
    "advanced_vault": {
        Name:        "Advanced Vault Management",
        Description: "Advanced vault management features",
        MinTier:     "enterprise",
    },
    "cross_org_management": {
        Name:        "Cross-organization Management",
        Description: "Cross-organization management capabilities",
        MinTier:     "enterprise",
        Dependencies: []string{"multi_tenant"},
    },
    "enterprise_policies": {
        Name:        "Enterprise Policies",
        Description: "Enterprise-wide policy management",
        MinTier:     "enterprise",
    },
    "api_access": {
        Name:        "Full API Access",
        Description: "Complete API access for enterprise automation",
        MinTier:     "enterprise",
        Dependencies: []string{"basic_api_access"},
    },
}

// GetFeaturesByTier returns all features available for a given tier
func GetFeaturesByTier(tier string) []FeatureConfig {
    features := make([]FeatureConfig, 0)

    for _, feature := range Features {
        if isFeatureAvailableForTier(feature.MinTier, tier) {
            features = append(features, feature)
        }
    }

    return features
}

// isFeatureAvailableForTier checks if a feature is available for a given tier
func isFeatureAvailableForTier(featureTier, userTier string) bool {
    tiers := map[string]int{
        "free":       0,
        "premium":    1,
        "enterprise": 2,
    }

    featureLevel, ok := tiers[featureTier]
    if !ok {
        return false
    }

    userLevel, ok := tiers[userTier]
    if !ok {
        return false
    }

    return userLevel >= featureLevel
}
