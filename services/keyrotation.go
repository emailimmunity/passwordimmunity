package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type KeyRotationService interface {
	RotateUserKeys(ctx context.Context, userID uuid.UUID) error
	RotateOrganizationKeys(ctx context.Context, orgID uuid.UUID) error
	GetCurrentKey(ctx context.Context, keyID uuid.UUID) ([]byte, error)
}

type keyRotationService struct {
	repo        repository.Repository
	encryption  EncryptionService
}

func NewKeyRotationService(repo repository.Repository, encryption EncryptionService) KeyRotationService {
	return &keyRotationService{
		repo:       repo,
		encryption: encryption,
	}
}

type KeyMetadata struct {
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
	KeyType     string    `json:"key_type"`
}

func (s *keyRotationService) RotateUserKeys(ctx context.Context, userID uuid.UUID) error {
	// Generate new key pair
	publicKey, privateKey, err := s.encryption.GenerateKeyPair()
	if err != nil {
		return err
	}

	// Create key metadata
	metadata := KeyMetadata{
		Version:   1,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(1, 0, 0), // 1 year expiration
		IsActive:  true,
		KeyType:   "user",
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Store new keys
	keyID := uuid.New()
	if err := s.repo.StoreKey(ctx, keyID, privateKey, publicKey, metadataBytes); err != nil {
		return err
	}

	// Get old active key
	oldKeyID, err := s.repo.GetActiveKeyID(ctx, userID)
	if err != nil {
		return err
	}

	oldKey, err := s.GetCurrentKey(ctx, oldKeyID)
	if err != nil {
		return err
	}

	// Re-encrypt vault items with new key
	if err := s.reencryptVaultItems(ctx, userID, oldKey, privateKey); err != nil {
		return err
	}

	// Update user's active key reference
	if err := s.repo.UpdateUserActiveKey(ctx, userID, keyID); err != nil {
		return err
	}

	// Mark old key as inactive
	if err := s.repo.DeactivateKey(ctx, oldKeyID); err != nil {
		return err
	}

	// Create audit log
	auditMetadata := createBasicMetadata("key_rotation", "User encryption keys rotated")
	if err := s.createAuditLog(ctx, "key.rotated", userID, uuid.Nil, auditMetadata); err != nil {
		return err
	}

	return nil
}

func (s *keyRotationService) RotateOrganizationKeys(ctx context.Context, orgID uuid.UUID) error {
	// Similar to RotateUserKeys but for organization-wide keys
	// Would need to handle re-encryption of shared vault items
	// and distribution of new keys to organization members
	return nil
}

func (s *keyRotationService) GetCurrentKey(ctx context.Context, keyID uuid.UUID) ([]byte, error) {
	// Retrieve and validate the current active key
	// Implementation would depend on key storage mechanism
	return nil, nil
}

// Helper function to re-encrypt vault items with new key
func (s *keyRotationService) reencryptVaultItems(ctx context.Context, userID uuid.UUID, oldKey, newKey []byte) error {
	// Fetch all vault items for user
	vaultItems, err := s.repo.GetUserVaultItems(ctx, userID)
	if err != nil {
		return err
	}

	// Re-encrypt each vault item
	for _, item := range vaultItems {
		// Decrypt with old key
		decryptedData, err := s.encryption.Decrypt(item.EncryptedData, oldKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt vault item %s: %w", item.ID, err)
		}

		// Re-encrypt with new key
		newEncryptedData, err := s.encryption.Encrypt(decryptedData, newKey)
		if err != nil {
			return fmt.Errorf("failed to re-encrypt vault item %s: %w", item.ID, err)
		}

		// Update vault item with new encrypted data
		item.EncryptedData = newEncryptedData
		if err := s.repo.UpdateVaultItem(ctx, item); err != nil {
			return fmt.Errorf("failed to update vault item %s: %w", item.ID, err)
		}
	}

	return nil
}

func (s *keyRotationService) RotateOrganizationKeys(ctx context.Context, orgID uuid.UUID) error {
	// Similar to RotateUserKeys but for organization-wide keys
	// Would need to handle re-encryption of shared vault items
	// and distribution of new keys to organization members
	return nil
}

func (s *keyRotationService) GetCurrentKey(ctx context.Context, keyID uuid.UUID) ([]byte, error) {
	// Retrieve and validate the current active key
	// Implementation would depend on key storage mechanism
	return nil, nil
}

// Helper function to re-encrypt vault items with new key
func (s *keyRotationService) reencryptVaultItems(ctx context.Context, userID uuid.UUID, oldKey, newKey []byte) error {
	// TODO: Implement re-encryption logic
	// 1. Fetch all vault items for user
	// 2. Decrypt with old key
	// 3. Re-encrypt with new key
	// 4. Update vault items with new encrypted data
	return nil
}
