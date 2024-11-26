package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"time"
	"bytes"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type ImportFormat string
type ExportFormat string

const (
	FormatBitwarden ImportFormat = "bitwarden"
	FormatLastPass  ImportFormat = "lastpass"
	FormatKeePass   ImportFormat = "keepass"
	FormatCSV       ImportFormat = "csv"
	FormatJSON      ImportFormat = "json"
)

type ImportExportService interface {
	ImportData(ctx context.Context, userID uuid.UUID, format ImportFormat, data []byte) error
	ExportData(ctx context.Context, userID uuid.UUID, format ExportFormat) ([]byte, error)
	ValidateImportData(ctx context.Context, format ImportFormat, data []byte) error
}

type importExportService struct {
	repo      repository.Repository
	vault     VaultService
	encryption EncryptionService
}

func NewImportExportService(repo repository.Repository, vault VaultService, encryption EncryptionService) ImportExportService {
	return &importExportService{
		repo:       repo,
		vault:      vault,
		encryption: encryption,
	}
}

func (s *importExportService) ImportData(ctx context.Context, userID uuid.UUID, format ImportFormat, data []byte) error {
	if err := s.ValidateImportData(ctx, format, data); err != nil {
		return err
	}

	switch format {
	case FormatBitwarden:
		return s.importBitwarden(ctx, userID, data)
	case FormatLastPass:
		return s.importLastPass(ctx, userID, data)
	case FormatKeePass:
		return s.importKeePass(ctx, userID, data)
	case FormatCSV:
		return s.importCSV(ctx, userID, data)
	default:
		return errors.New("unsupported import format")
	}
}

func (s *importExportService) ExportData(ctx context.Context, userID uuid.UUID, format ExportFormat) ([]byte, error) {
	items, err := s.vault.ListVaultItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatJSON:
		return s.exportJSON(items)
	case FormatCSV:
		return s.exportCSV(items)
	default:
		return nil, errors.New("unsupported export format")
	}
}

func (s *importExportService) ValidateImportData(ctx context.Context, format ImportFormat, data []byte) error {
	switch format {
	case FormatBitwarden:
		return s.validateBitwardenFormat(data)
	case FormatLastPass:
		return s.validateLastPassFormat(data)
	case FormatKeePass:
		return s.validateKeePassFormat(data)
	case FormatCSV:
		return s.validateCSVFormat(data)
	default:
		return errors.New("unsupported import format")
	}
}

func (s *importExportService) importBitwarden(ctx context.Context, userID uuid.UUID, data []byte) error {
	var items []models.VaultItem
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	for _, item := range items {
		item.ID = uuid.New()
		item.UserID = userID
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		if err := s.vault.CreateVaultItem(ctx, userID, &item); err != nil {
			continue
		}
	}

	// Create audit log
	metadata := createBasicMetadata("data_imported", "Data imported from Bitwarden")
	metadata["format"] = "bitwarden"
	return s.createAuditLog(ctx, "import.completed", userID, uuid.Nil, metadata)
}

func (s *importExportService) exportJSON(items []models.VaultItem) ([]byte, error) {
	return json.Marshal(items)
}

func (s *importExportService) exportCSV(items []models.VaultItem) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Write header
	header := []string{"name", "username", "password", "url", "notes"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write items
	for _, item := range items {
		row := []string{
			item.Name,
			item.Username,
			item.Password,
			item.URL,
			item.Notes,
		}
		if err := writer.Write(row); err != nil {
			continue
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}
