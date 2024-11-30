package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

func (s *service) CreateVaultItem(ctx context.Context, userID, orgID uuid.UUID, itemType, name string, data []byte) (*models.VaultItem, error) {
	// Verify user exists and has access
	hasAccess, err := s.hasPermission(ctx, userID, orgID, "create_vault_item")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	// Encrypt the data
	encryptedData, err := encryptData(data)
	if err != nil {
		return nil, err
	}

	item := &models.VaultItem{
		UserID:         userID,
		OrganizationID: orgID,
		Type:           itemType,
		Name:           name,
		Data:           encryptedData,
	}

	if err := s.repo.CreateVaultItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *service) GetVaultItems(ctx context.Context, userID, orgID uuid.UUID) ([]models.VaultItem, error) {
	// Verify user has access to organization
	hasAccess, err := s.hasPermission(ctx, userID, orgID, "read_vault_items")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	// TODO: Implement repository method to get vault items by user and org ID
	// For now, return empty slice
	return []models.VaultItem{}, nil
}

// Helper functions for encryption/decryption
func encryptData(data []byte) ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptData(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
