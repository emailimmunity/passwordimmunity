package services

import (
	"context"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

func (s *service) CreateOrganization(ctx context.Context, name, orgType string, ownerID uuid.UUID) (*models.Organization, error) {
	// Create organization
	org := &models.Organization{
		Name: name,
		Type: orgType,
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, err
	}

	// Create default admin role
	adminRole := &models.Role{
		Name:           "Admin",
		Description:    "Organization Administrator",
		OrganizationID: org.ID,
	}

	if err := s.repo.CreateRole(ctx, adminRole); err != nil {
		return nil, err
	}

	// Add owner to organization with admin role
	user, err := s.repo.GetUserByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	org.Users = append(org.Users, *user)
	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *service) AddUserToOrganization(ctx context.Context, orgID, userID, roleID uuid.UUID) error {
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return err
	}
	if org == nil {
		return ErrInvalidOperation
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil || role.OrganizationID != orgID {
		return ErrInvalidOperation
	}

	org.Users = append(org.Users, *user)
	return s.repo.UpdateOrganization(ctx, org)
}

func (s *service) RemoveUserFromOrganization(ctx context.Context, orgID, userID uuid.UUID) error {
	org, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return err
	}
	if org == nil {
		return ErrInvalidOperation
	}

	// Filter out the user from the organization's users
	var updatedUsers []models.User
	for _, u := range org.Users {
		if u.ID != userID {
			updatedUsers = append(updatedUsers, u)
		}
	}
	org.Users = updatedUsers

	return s.repo.UpdateOrganization(ctx, org)
}
