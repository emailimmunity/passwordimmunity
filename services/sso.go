package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type SSOProvider string

const (
	SSOProviderSAML   SSOProvider = "saml"
	SSOProviderOIDC   SSOProvider = "oidc"
	SSOProviderOAuth2 SSOProvider = "oauth2"
)

type SSOConfig struct {
	Provider      SSOProvider         `json:"provider"`
	ClientID      string             `json:"client_id"`
	ClientSecret  string             `json:"client_secret"`
	MetadataURL   string             `json:"metadata_url"`
	CallbackURL   string             `json:"callback_url"`
	Configuration map[string]string  `json:"configuration"`
	Enabled       bool               `json:"enabled"`
}

type SSOService interface {
	ConfigureSSO(ctx context.Context, orgID uuid.UUID, config SSOConfig) error
	GetSSOConfig(ctx context.Context, orgID uuid.UUID) (*SSOConfig, error)
	InitiateSSO(ctx context.Context, orgID uuid.UUID, provider SSOProvider) (string, error)
	HandleCallback(ctx context.Context, orgID uuid.UUID, code string) (*models.User, error)
}

type ssoService struct {
	repo repository.Repository
	encryption EncryptionService
}

func NewSSOService(repo repository.Repository, encryption EncryptionService) SSOService {
	return &ssoService{
		repo: repo,
		encryption: encryption,
	}
}

func (s *ssoService) ConfigureSSO(ctx context.Context, orgID uuid.UUID, config SSOConfig) error {
	// Encrypt sensitive configuration data
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	encryptedConfig, err := s.encryption.EncryptSymmetric(configBytes, []byte("your-encryption-key"))
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("sso_configured", "SSO configuration updated")
	metadata["provider"] = string(config.Provider)
	if err := s.createAuditLog(ctx, "sso.configured", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	// TODO: Store encrypted configuration
	return nil
}

func (s *ssoService) GetSSOConfig(ctx context.Context, orgID uuid.UUID) (*SSOConfig, error) {
	encryptedConfig, err := s.repo.GetSSOConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if encryptedConfig == nil {
		return nil, ErrSSONotConfigured
	}

	// Decrypt configuration
	decryptedBytes, err := s.encryption.DecryptSymmetric(encryptedConfig, []byte("your-encryption-key"))
	if err != nil {
		return nil, err
	}

	var config SSOConfig
	if err := json.Unmarshal(decryptedBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Helper function to validate SSO configuration
func (s *ssoService) validateSSOConfig(config SSOConfig) error {
	switch config.Provider {
	case SSOProviderSAML:
		if config.MetadataURL == "" {
			return errors.New("SAML metadata URL is required")
		}
	case SSOProviderOIDC:
		if config.ClientID == "" || config.ClientSecret == "" {
			return errors.New("OIDC client credentials are required")
		}
	case SSOProviderOAuth2:
		if config.ClientID == "" || config.ClientSecret == "" {
			return errors.New("OAuth2 client credentials are required")
		}
	default:
		return errors.New("unsupported SSO provider")
	}
	return nil
}

// Helper functions for SSO token exchange and user info fetching
func (s *ssoService) exchangeCodeForTokens(ctx context.Context, clientID, clientSecret, code, redirectURI, tokenURL string) (map[string]string, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
		"client_id":    {clientID},
		"client_secret": {clientSecret},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ssoService) fetchUserInfo(ctx context.Context, accessToken, userInfoURL string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (s *ssoService) parseSAMLResponse(samlResponse []byte, certPEM string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"email": "user@example.com",
		"name":  "Example User",
	}, nil
}

func (s *ssoService) HandleCallback(ctx context.Context, orgID uuid.UUID, code string) (*models.User, error) {
	// TODO: Implement SSO callback handling
	// 1. Validate state parameter
	// 2. Exchange code for tokens
	// 3. Get user information
	// 4. Create or update user
	// 5. Create audit log
	return nil, errors.New("not implemented")
}

func (s *ssoService) InitiateSSO(ctx context.Context, orgID uuid.UUID, provider SSOProvider) (string, error) {
	config, err := s.GetSSOConfig(ctx, orgID)
	if err != nil {
		return "", err
	}

	// Generate and store state parameter
	state := generateSecureToken()
	stateExpiry := time.Now().Add(15 * time.Minute)

	if err := s.repo.StoreSSOState(ctx, state, orgID, stateExpiry); err != nil {
		return "", err
	}

	// Create authorization URL based on provider
	var authURL string
	switch config.Provider {
	case SSOProviderOIDC, SSOProviderOAuth2:
		authURL = fmt.Sprintf("%s?client_id=%s&response_type=code&state=%s&redirect_uri=%s",
			config.Configuration["auth_url"],
			config.ClientID,
			state,
			config.CallbackURL)
	case SSOProviderSAML:
		authURL = fmt.Sprintf("%s?RelayState=%s", config.MetadataURL, state)
	default:
		return "", errors.New("unsupported SSO provider")
	}

	// Create audit log
	metadata := createBasicMetadata("sso_initiated", "SSO authentication initiated")
	metadata["provider"] = string(provider)
	if err := s.createAuditLog(ctx, AuditEventSSOInitiated, uuid.Nil, orgID, metadata); err != nil {
		return "", err
	}

	return authURL, nil
}

func (s *ssoService) HandleCallback(ctx context.Context, orgID uuid.UUID, code string) (*models.User, error) {
	// Validate state and get SSO config
	config, err := s.GetSSOConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	switch config.Provider {
	case SSOProviderOIDC, SSOProviderOAuth2:
		// Exchange code for tokens and get user info
		tokens, err := s.exchangeCodeForTokens(ctx, config, code)
		if err != nil {
			return nil, err
		}
		userInfo, err = s.getUserInfo(ctx, config, tokens["access_token"])
		if err != nil {
			return nil, err
		}
	case SSOProviderSAML:
		userInfo, err = s.parseSAMLResponse(ctx, code)
		if err != nil {
			return nil, err
		}
	}

	// Create or update user
	user, err := s.repo.GetUserByEmail(ctx, userInfo["email"].(string))
	if err != nil {
		return nil, err
	}

	if user == nil {
		// Create new user
		user = &models.User{
			Email:     userInfo["email"].(string),
			Name:      userInfo["name"].(string),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	// Create audit log
	metadata := createBasicMetadata("sso_completed", "SSO authentication completed")
	metadata["provider"] = string(config.Provider)
	metadata["email"] = user.Email
	if err := s.createAuditLog(ctx, AuditEventSSOCompleted, user.ID, orgID, metadata); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ssoService) InitiateSSO(ctx context.Context, orgID uuid.UUID, provider SSOProvider) (string, error) {
	config, err := s.GetSSOConfig(ctx, orgID)
	if err != nil {
		return "", err
	}

	// Generate and store state parameter
	state := generateSecureToken()
	stateExpiry := time.Now().Add(15 * time.Minute)

	if err := s.repo.StoreSSOState(ctx, state, orgID, stateExpiry); err != nil {
		return "", err
	}

	// Create authorization URL based on provider
	var authURL string
	switch config.Provider {
	case SSOProviderOIDC, SSOProviderOAuth2:
		authURL = fmt.Sprintf("%s?client_id=%s&response_type=code&state=%s&redirect_uri=%s",
			config.Configuration["auth_url"],
			config.ClientID,
			state,
			config.CallbackURL)
	case SSOProviderSAML:
		authURL = fmt.Sprintf("%s?RelayState=%s", config.MetadataURL, state)
	default:
		return "", errors.New("unsupported SSO provider")
	}

	// Create audit log
	metadata := createBasicMetadata("sso_initiated", "SSO authentication initiated")
	metadata["provider"] = string(provider)
	if err := s.createAuditLog(ctx, AuditEventSSOInitiated, uuid.Nil, orgID, metadata); err != nil {
		return "", err
	}

	return authURL, nil
}

func (s *ssoService) HandleCallback(ctx context.Context, orgID uuid.UUID, code string) (*models.User, error) {
	// TODO: Implement SSO callback handling
	// 1. Validate state parameter
	// 2. Exchange code for tokens
	// 3. Get user information
	// 4. Create or update user
	// 5. Create audit log
	return nil, errors.New("not implemented")
}

func (s *ssoService) InitiateSSO(ctx context.Context, orgID uuid.UUID, provider SSOProvider) (string, error) {
	// TODO: Implement SSO initiation
	// 1. Get SSO configuration
	// 2. Generate state parameter
	// 3. Create authorization URL
	// 4. Return URL for redirect
	return "", errors.New("not implemented")
}

func (s *ssoService) HandleCallback(ctx context.Context, orgID uuid.UUID, code string) (*models.User, error) {
	// TODO: Implement SSO callback handling
	// 1. Validate state parameter
	// 2. Exchange code for tokens
	// 3. Get user information
	// 4. Create or update user
	// 5. Create audit log
	return nil, errors.New("not implemented")
}

// Helper function to validate SSO configuration
func (s *ssoService) validateSSOConfig(config SSOConfig) error {
	switch config.Provider {
	case SSOProviderSAML:
		if config.MetadataURL == "" {
			return errors.New("SAML metadata URL is required")
		}
	case SSOProviderOIDC:
		if config.ClientID == "" || config.ClientSecret == "" {
			return errors.New("OIDC client credentials are required")
		}
	case SSOProviderOAuth2:
		if config.ClientID == "" || config.ClientSecret == "" {
			return errors.New("OAuth2 client credentials are required")
		}
	default:
		return errors.New("unsupported SSO provider")
	}
	return nil
}
