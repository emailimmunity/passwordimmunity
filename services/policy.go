package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type PolicyType string

const (
	PolicyTwoFactorAuth     PolicyType = "two_factor_auth"
	PolicyPasswordComplexity PolicyType = "password_complexity"
	PolicySessionTimeout    PolicyType = "session_timeout"
	PolicyIPAllowlist      PolicyType = "ip_allowlist"
	PolicyMasterPassword   PolicyType = "master_password"
	PolicyVaultTimeout     PolicyType = "vault_timeout"
)

type Policy struct {
	Type        PolicyType          `json:"type"`
	Enabled     bool               `json:"enabled"`
	Settings    json.RawMessage    `json:"settings"`
}

type PolicyService interface {
	CreatePolicy(ctx context.Context, orgID uuid.UUID, policy Policy) error
	UpdatePolicy(ctx context.Context, orgID uuid.UUID, policy Policy) error
	DeletePolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType) error
	GetPolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType) (*Policy, error)
	ListPolicies(ctx context.Context, orgID uuid.UUID) ([]Policy, error)
	EvaluatePolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType, data interface{}) (bool, error)
}

type policyService struct {
	repo repository.Repository
}

func NewPolicyService(repo repository.Repository) PolicyService {
	return &policyService{repo: repo}
}

func (s *policyService) CreatePolicy(ctx context.Context, orgID uuid.UUID, policy Policy) error {
	// Create audit log
	metadata := createBasicMetadata("policy_created", "Policy created")
	metadata["policy_type"] = string(policy.Type)
	if err := s.createAuditLog(ctx, "policy.created", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.CreatePolicy(ctx, orgID, policy)
}

func (s *policyService) UpdatePolicy(ctx context.Context, orgID uuid.UUID, policy Policy) error {
	// Create audit log
	metadata := createBasicMetadata("policy_updated", "Policy updated")
	metadata["policy_type"] = string(policy.Type)
	if err := s.createAuditLog(ctx, "policy.updated", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdatePolicy(ctx, orgID, policy)
}

func (s *policyService) DeletePolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType) error {
	// Create audit log
	metadata := createBasicMetadata("policy_deleted", "Policy deleted")
	metadata["policy_type"] = string(policyType)
	if err := s.createAuditLog(ctx, "policy.deleted", uuid.Nil, orgID, metadata); err != nil {
		return err
	}

	return s.repo.DeletePolicy(ctx, orgID, policyType)
}

func (s *policyService) GetPolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType) (*Policy, error) {
	return s.repo.GetPolicy(ctx, orgID, policyType)
}

func (s *policyService) ListPolicies(ctx context.Context, orgID uuid.UUID) ([]Policy, error) {
	return s.repo.ListPolicies(ctx, orgID)
}

func (s *policyService) EvaluatePolicy(ctx context.Context, orgID uuid.UUID, policyType PolicyType, data interface{}) (bool, error) {
	policy, err := s.GetPolicy(ctx, orgID, policyType)
	if err != nil {
		return false, err
	}

	if policy == nil || !policy.Enabled {
		return true, nil
	}

	switch policyType {
	case PolicyTwoFactorAuth:
		return s.evaluateTwoFactorAuthPolicy(policy, data)
	case PolicyPasswordComplexity:
		return s.evaluatePasswordComplexityPolicy(policy, data)
	case PolicySessionTimeout:
		return s.evaluateSessionTimeoutPolicy(policy, data)
	case PolicyIPAllowlist:
		return s.evaluateIPAllowlistPolicy(policy, data)
	default:
		return true, nil
	}
}

func (s *policyService) evaluateTwoFactorAuthPolicy(policy *Policy, data interface{}) (bool, error) {
	// TODO: Implement 2FA policy evaluation
	return true, nil
}

func (s *policyService) evaluatePasswordComplexityPolicy(policy *Policy, data interface{}) (bool, error) {
	// TODO: Implement password complexity policy evaluation
	return true, nil
}

func (s *policyService) evaluateSessionTimeoutPolicy(policy *Policy, data interface{}) (bool, error) {
	// TODO: Implement session timeout policy evaluation
	return true, nil
}

func (s *policyService) evaluateIPAllowlistPolicy(policy *Policy, data interface{}) (bool, error) {
	// TODO: Implement IP allowlist policy evaluation
	return true, nil
}
