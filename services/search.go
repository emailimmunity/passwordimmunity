package services

import (
	"context"
	"strings"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type SearchOptions struct {
	Query       string
	Types       []string
	Collections []uuid.UUID
	StartTime   *time.Time
	EndTime     *time.Time
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

type SearchService interface {
	Search(ctx context.Context, userID uuid.UUID, options SearchOptions) ([]models.SearchResult, error)
	SearchVaultItems(ctx context.Context, userID uuid.UUID, query string) ([]models.VaultItem, error)
	SearchCollections(ctx context.Context, userID uuid.UUID, query string) ([]models.Collection, error)
	SearchAuditLogs(ctx context.Context, orgID uuid.UUID, query string) ([]models.AuditLog, error)
	BuildSearchIndex(ctx context.Context, orgID uuid.UUID) error
}

type searchService struct {
	repo        repository.Repository
	vault       VaultService
	collection  CollectionService
	audit       AuditService
}

func NewSearchService(
	repo repository.Repository,
	vault VaultService,
	collection CollectionService,
	audit AuditService,
) SearchService {
	return &searchService{
		repo:       repo,
		vault:      vault,
		collection: collection,
		audit:      audit,
	}
}

func (s *searchService) Search(ctx context.Context, userID uuid.UUID, options SearchOptions) ([]models.SearchResult, error) {
	var results []models.SearchResult

	// Search vault items
	if len(options.Types) == 0 || contains(options.Types, "vault_item") {
		items, err := s.SearchVaultItems(ctx, userID, options.Query)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			results = append(results, models.SearchResult{
				Type:      "vault_item",
				ID:        item.ID,
				Title:     item.Name,
				Subtitle:  item.Username,
				UpdatedAt: item.UpdatedAt,
			})
		}
	}

	// Search collections
	if len(options.Types) == 0 || contains(options.Types, "collection") {
		collections, err := s.SearchCollections(ctx, userID, options.Query)
		if err != nil {
			return nil, err
		}
		for _, col := range collections {
			results = append(results, models.SearchResult{
				Type:      "collection",
				ID:        col.ID,
				Title:     col.Name,
				UpdatedAt: col.UpdatedAt,
			})
		}
	}

	// Apply filters
	results = s.filterResults(results, options)

	// Sort results
	s.sortResults(results, options)

	// Apply pagination
	if options.Limit > 0 {
		end := options.Offset + options.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[options.Offset:end]
	}

	return results, nil
}

func (s *searchService) SearchVaultItems(ctx context.Context, userID uuid.UUID, query string) ([]models.VaultItem, error) {
	return s.repo.SearchVaultItems(ctx, userID, query)
}

func (s *searchService) SearchCollections(ctx context.Context, userID uuid.UUID, query string) ([]models.Collection, error) {
	return s.repo.SearchCollections(ctx, userID, query)
}

func (s *searchService) SearchAuditLogs(ctx context.Context, orgID uuid.UUID, query string) ([]models.AuditLog, error) {
	return s.repo.SearchAuditLogs(ctx, orgID, query)
}

func (s *searchService) BuildSearchIndex(ctx context.Context, orgID uuid.UUID) error {
	// Create audit log
	metadata := createBasicMetadata("search_index_built", "Search index rebuilt")
	if err := s.createAuditLog(ctx, "search.index.built", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.BuildSearchIndex(ctx, orgID)
}

func (s *searchService) filterResults(results []models.SearchResult, options SearchOptions) []models.SearchResult {
	filtered := make([]models.SearchResult, 0)

	for _, result := range results {
		// Filter by time range
		if options.StartTime != nil && result.UpdatedAt.Before(*options.StartTime) {
			continue
		}
		if options.EndTime != nil && result.UpdatedAt.After(*options.EndTime) {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

func (s *searchService) sortResults(results []models.SearchResult, options SearchOptions) {
	if options.SortBy == "" {
		options.SortBy = "updated_at"
	}
	if options.SortOrder == "" {
		options.SortOrder = "desc"
	}

	// Implement sorting logic based on options.SortBy and options.SortOrder
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
