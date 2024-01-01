package repo

import (
	"common/slugerror"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"policy/domain/entity"
	"policy/domain/repository"
)

var RecordNotFoundError = slugerror.NewSlugError(slugerror.ErrUnprocessable,
	"request unprocessable", "record not found in the database")

type Policy struct {
	gorm.Model
	AccountID  uint   `gorm:"index, not null"`
	PolicyType string `gorm:"not null"`
	AccessKey  string `gorm:"not null"`
	SecretKey  string `gorm:"not null"`
	Bucket     string `gorm:"not null"`
	Area       string `gorm:"not null"`
	Detail     string
	Status     entity.PolicyStatus `gorm:"not null"`
}

type policyRepo struct {
	db *gorm.DB
}

func NewPolicyRepo(db *gorm.DB) (repository.PolicyRepo, error) {
	if db == nil {
		panic("missing db")
	}
	err := db.AutoMigrate(&Policy{})
	if err != nil {
		return nil, err
	}
	return &policyRepo{db: db}, nil
}

func fromDomainPolicy(policy entity.Policy) (*Policy, error) {
	detailBlob, err := json.Marshal(policy.Detail())
	if err != nil {
		return nil, err
	}
	return &Policy{
		AccountID:  policy.AccountID(),
		PolicyType: policy.PolicyType(),
		AccessKey:  policy.AccessKey(),
		SecretKey:  policy.SecretKey(),
		Bucket:     policy.Bucket(),
		Area:       policy.Area(),
		Detail:     string(detailBlob),
		Status:     policy.Status(),
	}, nil
}

func toDomainPolicy(policy Policy) (*entity.Policy, error) {
	return entity.UnmarshallPolicy(policy.ID, policy.AccountID, policy.PolicyType, policy.AccessKey, policy.SecretKey,
		policy.Bucket, policy.Area, policy.Detail, policy.Status)
}

func toDomainPolicies(policies []Policy) ([]entity.Policy, error) {
	result := make([]entity.Policy, len(policies))
	for i, policy := range policies {
		domainPolicy, err := toDomainPolicy(policy)
		if err != nil {
			return nil, err
		}
		result[i] = *domainPolicy
	}
	return result, nil
}

func (p *policyRepo) CreatePolicy(policy entity.Policy) (*entity.Policy, error) {
	mysqlPolicy, err := fromDomainPolicy(policy)
	if err != nil {
		return nil, err
	}
	// 先检查该账户是否已经有了该类型的policy，有则不创建
	existedPolicy := Policy{}
	if err := p.db.First(&existedPolicy, "account_id = ? and policy_type = ?",
		mysqlPolicy.AccountID, mysqlPolicy.PolicyType).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if existedPolicy.ID != 0 { // policy类型对于该账户已存在
		return nil, slugerror.NewSlugError(slugerror.ErrInvalidInput, "invalid policy type", "existed policy type for the account")
	}
	if err := p.db.Create(mysqlPolicy).Error; err != nil {
		return nil, err
	}
	return toDomainPolicy(*mysqlPolicy)
}

func (p *policyRepo) GetPolicy(policyID uint) (*entity.Policy, error) {
	mysqlPolicy := Policy{}
	if err := p.db.First(&mysqlPolicy, "id = ?", policyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, RecordNotFoundError
		}
		return nil, err
	}
	return toDomainPolicy(mysqlPolicy)
}

func (p *policyRepo) GetPolicies(accountID uint) ([]entity.Policy, error) {
	mysqlPolicies := []Policy{}
	if err := p.db.Find(&mysqlPolicies, "account_id = ?", accountID).Error; err != nil {
		return nil, err
	}
	return toDomainPolicies(mysqlPolicies)
}

func (p *policyRepo) UpdatePolicyFields(policy entity.Policy) error {
	mysqlPolicy, err := fromDomainPolicy(policy)
	if err != nil {
		return err
	}
	result := p.db.Model(&Policy{}).Where("id = ?", policy.Id()).Updates(*mysqlPolicy)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return RecordNotFoundError
	}
	return nil
}

func (p *policyRepo) SetDefaultPolicyOff(accountID uint) error {
	// 更新该账号中on的policy为off，没有已启用的也不报错（第一次会这样）
	if err := p.db.Model(&Policy{}).Where("account_id = ? and status = ?",
		accountID, entity.PolicyOn).Update("status", entity.PolicyOff).Error; err != nil {
		return err
	}
	return nil
}

func (p *policyRepo) SetPolicyOn(policyID uint) error {
	if err := p.db.Model(&Policy{}).Where("id = ?", policyID).
		Update("status", entity.PolicyOn).Error; err != nil {
		return err
	}
	return nil
}
