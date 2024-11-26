package services

import (
	"context"
	"time"
	"regexp"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type PasswordPolicyService interface {
	ValidatePassword(ctx context.Context, orgID uuid.UUID, password string) (bool, []string)
	GetPolicy(ctx context.Context, orgID uuid.UUID) (*models.PasswordPolicy, error)
	UpdatePolicy(ctx context.Context, orgID uuid.UUID, policy *models.PasswordPolicy) error
	CheckPasswordHistory(ctx context.Context, userID uuid.UUID, password string) (bool, error)
	EnforcePasswordRotation(ctx context.Context, orgID uuid.UUID) error
}

type passwordPolicyService struct {
	repo      repository.Repository
	cache     CacheService
	audit     AuditService
	orgPrefs  OrganizationPreferencesService
}

func NewPasswordPolicyService(
	repo repository.Repository,
	cache CacheService,
	audit AuditService,
	orgPrefs OrganizationPreferencesService,
) PasswordPolicyService {
	return &passwordPolicyService{
		repo:      repo,
		cache:     cache,
		audit:     audit,
		orgPrefs:  orgPrefs,
	}
}

func (s *passwordPolicyService) ValidatePassword(ctx context.Context, orgID uuid.UUID, password string) (bool, []string) {
	policy, err := s.GetPolicy(ctx, orgID)
	if err != nil {
		return false, []string{"Error fetching password policy"}
	}

	var violations []string

	if len(password) < policy.MinLength {
		violations = append(violations, fmt.Sprintf("Password must be at least %d characters", policy.MinLength))
	}

	if policy.RequireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		violations = append(violations, "Password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		violations = append(violations, "Password must contain at least one lowercase letter")
	}

	if policy.RequireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		violations = append(violations, "Password must contain at least one number")
	}

	if policy.RequireSpecial && !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		violations = append(violations, "Password must contain at least one special character")
	}

	return len(violations) == 0, violations
}

func (s *passwordPolicyService) GetPolicy(ctx context.Context, orgID uuid.UUID) (*models.PasswordPolicy, error) {
	cacheKey := fmt.Sprintf("password_policy:%s", orgID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(*models.PasswordPolicy), nil
	}

	policy, err := s.repo.GetPasswordPolicy(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if policy != nil {
		s.cache.Set(ctx, cacheKey, policy, time.Hour*24)
	}

	return policy, nil
}

func (s *passwordPolicyService) UpdatePolicy(ctx context.Context, orgID uuid.UUID, policy *models.PasswordPolicy) error {
	policy.UpdatedAt = time.Now()

	if err := s.repo.UpdatePasswordPolicy(ctx, orgID, policy); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("password_policy:%s", orgID)
	s.cache.Delete(ctx, cacheKey)

	metadata := map[string]interface{}{
		"org_id":           orgID,
		"min_length":       policy.MinLength,
		"require_upper":    policy.RequireUppercase,
		"require_lower":    policy.RequireLowercase,
		"require_numbers":  policy.RequireNumbers,
		"require_special":  policy.RequireSpecial,
		"max_age":         policy.MaxAge,
	}
	if err := s.audit.Log(ctx, "password.policy.updated", metadata); err != nil {
		return err
	}

	return nil
}

func (s *passwordPolicyService) CheckPasswordHistory(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	history, err := s.repo.GetPasswordHistory(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, oldPassword := range history {
		if oldPassword.Hash == password {
			return true, nil
		}
	}

	return false, nil
}

func (s *passwordPolicyService) EnforcePasswordRotation(ctx context.Context, orgID uuid.UUID) error {
	policy, err := s.GetPolicy(ctx, orgID)
	if err != nil {
		return err
	}

	if policy.MaxAge <= 0 {
		return nil
	}

	users, err := s.repo.GetOrganizationUsers(ctx, orgID)
	if err != nil {
		return err
	}

	expirationDate := time.Now().AddDate(0, 0, -policy.MaxAge)

	for _, user := range users {
		if user.LastPasswordChange.Before(expirationDate) {
			if err := s.repo.SetPasswordExpired(ctx, user.ID, true); err != nil {
				return err
			}

			metadata := map[string]interface{}{
				"user_id": user.ID,
				"org_id":  orgID,
			}
			if err := s.audit.Log(ctx, "password.expired", metadata); err != nil {
				return err
			}
		}
	}

	return nil
}
