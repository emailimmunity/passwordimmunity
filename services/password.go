package services

import (
	"context"
	"errors"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type PasswordPolicy struct {
	MinLength         int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireNumbers    bool
	RequireSpecial    bool
	MaximumAge        int // in days
	PreventReuse      bool
	HistorySize       int
	RequireUnique     bool
	AllowedCharacters string
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:         12,
	RequireUppercase:  true,
	RequireLowercase:  true,
	RequireNumbers:    true,
	RequireSpecial:    true,
	MaximumAge:        90,
	PreventReuse:      true,
	HistorySize:       5,
	RequireUnique:     true,
	AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?",
}

func (s *service) UpdateUserPassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Validate new password against policy
	if err := validatePassword(newPassword, DefaultPasswordPolicy); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("password_changed", "User password changed")
	if err := s.createAuditLog(ctx, AuditEventUserPasswordChanged, user.ID, uuid.Nil, metadata); err != nil {
		// Log error but don't fail the operation
		return nil
	}

	return nil
}

func validatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return errors.New("password too short")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if policy.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if policy.RequireNumbers && !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if policy.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
