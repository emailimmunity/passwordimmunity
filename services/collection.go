package services

import (
	"context"
	"errors"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type CollectionService interface {
	CreateCollection(ctx context.Context, orgID uuid.UUID, name string) (*models.Collection, error)
	UpdateCollection(ctx context.Context, collection *models.Collection) error
	DeleteCollection(ctx context.Context, collectionID uuid.UUID) error
	GetCollection(ctx context.Context, collectionID uuid.UUID) (*models.Collection, error)
	ListCollections(ctx context.Context, orgID uuid.UUID) ([]models.Collection, error)
	AddUserToCollection(ctx context.Context, collectionID, userID uuid.UUID, readOnly bool) error
	RemoveUserFromCollection(ctx context.Context, collectionID, userID uuid.UUID) error
}

type collectionService struct {
	repo repository.Repository
	roleService RoleService
}

func NewCollectionService(repo repository.Repository, roleService RoleService) CollectionService {
	return &collectionService{
		repo: repo,
		roleService: roleService,
	}
}

func (s *collectionService) CreateCollection(ctx context.Context, orgID uuid.UUID, name string) (*models.Collection, error) {
	collection := &models.Collection{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateCollection(ctx, collection); err != nil {
		return nil, err
	}

	// Create audit log
	metadata := createBasicMetadata("collection_created", "Collection created")
	metadata["collection_name"] = name
	if err := s.createAuditLog(ctx, "collection.created", uuid.Nil, orgID, metadata); err != nil {
		return nil, err
	}

	return collection, nil
}

func (s *collectionService) UpdateCollection(ctx context.Context, collection *models.Collection) error {
	collection.UpdatedAt = time.Now()

	// Create audit log
	metadata := createBasicMetadata("collection_updated", "Collection updated")
	metadata["collection_name"] = collection.Name
	if err := s.createAuditLog(ctx, "collection.updated", uuid.Nil, collection.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.UpdateCollection(ctx, collection)
}

func (s *collectionService) DeleteCollection(ctx context.Context, collectionID uuid.UUID) error {
	collection, err := s.GetCollection(ctx, collectionID)
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("collection_deleted", "Collection deleted")
	metadata["collection_name"] = collection.Name
	if err := s.createAuditLog(ctx, "collection.deleted", uuid.Nil, collection.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.DeleteCollection(ctx, collectionID)
}

func (s *collectionService) GetCollection(ctx context.Context, collectionID uuid.UUID) (*models.Collection, error) {
	return s.repo.GetCollection(ctx, collectionID)
}

func (s *collectionService) ListCollections(ctx context.Context, orgID uuid.UUID) ([]models.Collection, error) {
	return s.repo.ListCollections(ctx, orgID)
}

func (s *collectionService) AddUserToCollection(ctx context.Context, collectionID, userID uuid.UUID, readOnly bool) error {
	collection, err := s.GetCollection(ctx, collectionID)
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("user_added_to_collection", "User added to collection")
	metadata["collection_name"] = collection.Name
	metadata["read_only"] = readOnly
	if err := s.createAuditLog(ctx, "collection.user.added", userID, collection.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.AddCollectionUser(ctx, collectionID, userID, readOnly)
}

func (s *collectionService) RemoveUserFromCollection(ctx context.Context, collectionID, userID uuid.UUID) error {
	collection, err := s.GetCollection(ctx, collectionID)
	if err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("user_removed_from_collection", "User removed from collection")
	metadata["collection_name"] = collection.Name
	if err := s.createAuditLog(ctx, "collection.user.removed", userID, collection.OrganizationID, metadata); err != nil {
		return err
	}

	return s.repo.RemoveCollectionUser(ctx, collectionID, userID)
}
