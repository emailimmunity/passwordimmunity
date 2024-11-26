package auth

import (
	"context"
	"time"
)

// Provider defines the interface for authentication providers
type Provider interface {
	// Authenticate validates user credentials
	Authenticate(ctx context.Context, username, password string) (*AuthResult, error)

	// ValidateTwoFactor validates 2FA tokens
	ValidateTwoFactor(ctx context.Context, userID string, token string) error

	// GetUserInfo retrieves user information
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
}

// AuthResult contains authentication response data
type AuthResult struct {
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Requires2FA bool    `json:"requires_2fa"`
}

// UserInfo contains user profile information
type UserInfo struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	Roles       []string  `json:"roles"`
}

// NewDefaultProvider creates a new instance of the default authentication provider
func NewDefaultProvider() Provider {
	return &defaultProvider{}
}

type defaultProvider struct {
	// Add provider-specific fields here
}

func (p *defaultProvider) Authenticate(ctx context.Context, username, password string) (*AuthResult, error) {
	// Implementation will be added in separate PR
	return nil, nil
}

func (p *defaultProvider) ValidateTwoFactor(ctx context.Context, userID string, token string) error {
	// Implementation will be added in separate PR
	return nil
}

func (p *defaultProvider) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	// Implementation will be added in separate PR
	return nil, nil
}
