package repository

import "policy/domain/entity"

type PolicyRepo interface {
	CreatePolicy(policy entity.Policy) (*entity.Policy, error)
	GetPolicy(policyID uint) (*entity.Policy, error)
	GetPolicies(accountID uint) ([]entity.Policy, error)
	UpdatePolicyFields(policy entity.Policy) error
	SetDefaultPolicyOff(accountID uint) error
	SetPolicyOn(policyID uint) error
}
