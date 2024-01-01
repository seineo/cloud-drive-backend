package service

import (
	"policy/domain/entity"
	"policy/domain/repository"
)

type PolicyService interface {
	CreatePolicy(accountID uint, policyType string, accessKey string, secretKey string,
		bucket string, area string, detail map[string]string) (*entity.Policy, error)
	GetPolicies(accountID uint) ([]entity.Policy, error)
	ModifyPolicy(policyID uint, toUpdate map[string]interface{}) error
	SwitchOnPolicy(accountID uint, policyID uint) error
}

type policyService struct {
	policyRepo repository.PolicyRepo
	policyFc   entity.PolicyFactoryConfig
}

func (p *policyService) CreatePolicy(accountID uint, policyType string, accessKey string, secretKey string,
	bucket string, area string, detail map[string]string) (*entity.Policy, error) {
	policyFactory, err := entity.NewPolicyFactory(p.policyFc)
	if err != nil {
		return nil, err
	}
	policy, err := policyFactory.NewPolicy(accountID, policyType, accessKey, secretKey, bucket,
		area, detail)
	if err != nil {
		return nil, err
	}
	newPolicy, err := p.policyRepo.CreatePolicy(*policy)
	if err != nil {
		return nil, err
	}
	return newPolicy, nil
}

func (p *policyService) GetPolicies(accountID uint) ([]entity.Policy, error) {
	policies, err := p.policyRepo.GetPolicies(accountID)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (p *policyService) ModifyPolicy(policyID uint, toUpdate map[string]interface{}) error {
	policy, err := p.policyRepo.GetPolicy(policyID)
	if err != nil {
		return err
	}
	for key, value := range toUpdate {
		switch key {
		case "accessKey":
			policy.SetAccessKey(value.(string))
		case "secretKey":
			policy.SetSecretKey(value.(string))
		case "bucket":
			policy.SetBucket(value.(string))
		case "area":
			policy.SetArea(value.(string))
		case "detail":
			policy.SetDetail(value.(map[string]string))
		}
	}
	err = p.policyRepo.UpdatePolicyFields(*policy)
	if err != nil {
		return err
	}
	return nil
}

// SwitchOnPolicy TODO 需要事务
func (p *policyService) SwitchOnPolicy(accountID uint, policyID uint) error {
	// 打开一个policy，需要把已打开的policy关闭
	err := p.policyRepo.SetDefaultPolicyOff(accountID)
	if err != nil {
		return err
	}
	err = p.policyRepo.SetPolicyOn(policyID)
	if err != nil {
		return err
	}
	return nil
}

func NewPolicyService(policyRepo repository.PolicyRepo, policyFc entity.PolicyFactoryConfig) PolicyService {
	return &policyService{policyRepo: policyRepo, policyFc: policyFc}
}
