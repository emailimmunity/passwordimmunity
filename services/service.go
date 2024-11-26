package services

import (
	"context"
	"errors"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/emailimmunity/passwordimmunity/db/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrEmailExists      = errors.New("email already exists")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidOperation = errors.New("invalid operation")
)

type Service interface {
	// User operations
	CreateUser(ctx context.Context, email, name, password string) (*models.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	Enable2FA(ctx context.Context, userID uuid.UUID) (string, error)
	Verify2FA(ctx context.Context, userID uuid.UUID, code string) error

	// Organization operations
	CreateOrganization(ctx context.Context, name, orgType string, ownerID uuid.UUID) (*models.Organization, error)
	AddUserToOrganization(ctx context.Context, orgID, userID, roleID uuid.UUID) error
	RemoveUserFromOrganization(ctx context.Context, orgID, userID uuid.UUID) error

	// Role and Permission operations
	CreateRole(ctx context.Context, orgID uuid.UUID, name, description string) (*models.Role, error)
	AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// Vault operations
	CreateVaultItem(ctx context.Context, userID, orgID uuid.UUID, itemType, name string, data []byte) (*models.VaultItem, error)
	GetVaultItems(ctx context.Context, userID, orgID uuid.UUID) ([]models.VaultItem, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

// User operations implementation
func (s *service) CreateUser(ctx context.Context, email, name, password string) (*models.User, error) {
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// Additional service implementations will be added in subsequent files...
