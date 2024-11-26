package services

import (
	"context"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

func (s *service) CreateRole(ctx context.Context, orgID uuid.UUID, name, description string) (*models.Role, error) {
	// Verify organization exists
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrInvalidOperation
	}

	role := &models.Role{
		Name:           name,
		Description:    description,
		OrganizationID: orgID,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *service) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrInvalidOperation
	}

	var permissions []models.Permission
	for _, permID := range permissionIDs {
		perm := models.Permission{
			Base: models.Base{ID: permID},
		}
		permissions = append(permissions, perm)
	}

	role.Permissions = permissions
	return s.repo.UpdateRole(ctx, role)
}

// Helper function to check if a user has a specific permission in an organization
func (s *service) hasPermission(ctx context.Context, userID, orgID uuid.UUID, permissionName string) (bool, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, ErrUserNotFound
	}

	// Check each organization the user belongs to
	for _, org := range user.Organizations {
		if org.ID == orgID {
			// Check each role in the organization
			for _, role := range org.Roles {
				// Check each permission in the role
				for _, perm := range role.Permissions {
					if perm.Name == permissionName {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
