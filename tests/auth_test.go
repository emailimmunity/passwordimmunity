package tests

import (
	"context"
	"testing"
	"time"

	"github.com/emailimmunity/passwordimmunity/auth"
	"github.com/google/uuid"
)

func TestAuthentication(t *testing.T) {
	provider := auth.NewDefaultProvider()

	t.Run("Successful Authentication", func(t *testing.T) {
		ctx := context.Background()
		result, err := provider.Authenticate(ctx, "test@example.com", "password123")
		if err != nil {
			t.Errorf("Expected successful authentication, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected auth result, got nil")
		}
	})

	t.Run("Two Factor Validation", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New().String()
		err := provider.ValidateTwoFactor(ctx, userID, "123456")
		if err != nil {
			t.Errorf("Expected successful 2FA validation, got error: %v", err)
		}
	})

	t.Run("User Info Retrieval", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New().String()
		info, err := provider.GetUserInfo(ctx, userID)
		if err != nil {
			t.Errorf("Expected successful user info retrieval, got error: %v", err)
		}
		if info == nil {
			t.Error("Expected user info, got nil")
		}
	})
}
