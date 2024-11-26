package services

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

// Enable2FA generates a new TOTP secret for a user and returns the secret key
func (s *service) Enable2FA(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	// Generate random secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// Update user with new 2FA secret
	user.TwoFactorSecret = secretBase32
	user.TwoFactorEnabled = false // Will be enabled after verification
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return "", err
	}

	// Create audit log
	metadata := createBasicMetadata("2fa_setup_initiated", "Two-factor authentication setup initiated")
	if err := s.createAuditLog(ctx, AuditEventUserPasswordChanged, user.ID, uuid.Nil, metadata); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return secretBase32, nil
}

// Verify2FA validates a TOTP code and enables 2FA if valid
func (s *service) Verify2FA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify TOTP code
	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return ErrInvalidOperation
	}

	// Enable 2FA
	user.TwoFactorEnabled = true
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("2fa_enabled", "Two-factor authentication enabled")
	if err := s.createAuditLog(ctx, AuditEventUserPasswordChanged, user.ID, uuid.Nil, metadata); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

// Validate2FACode validates a TOTP code for an already-enabled 2FA user
func (s *service) Validate2FACode(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if !user.TwoFactorEnabled {
		return ErrInvalidOperation
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return ErrUnauthorized
	}

	return nil
}

// Generate2FABackupCodes generates new backup codes for a user
func (s *service) Generate2FABackupCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if !user.TwoFactorEnabled {
		return nil, ErrInvalidOperation
	}

	// Generate 8 backup codes
	var backupCodes []string
	for i := 0; i < 8; i++ {
		code := make([]byte, 4)
		if _, err := rand.Read(code); err != nil {
			return nil, err
		}
		backupCodes = append(backupCodes, fmt.Sprintf("%x", code))
	}

	// TODO: Store hashed backup codes in the database

	return backupCodes, nil
}
