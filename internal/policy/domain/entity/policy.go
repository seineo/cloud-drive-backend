package entity

import (
	"common/slugerror"
	"encoding/json"
	"fmt"
)

type PolicyStatus string

const (
	PolicyOn  = "on"
	PolicyOff = "off"
)

type Policy struct {
	id         uint
	accountID  uint
	policyType string
	accessKey  string
	secretKey  string
	bucket     string
	area       string
	detail     map[string]string
	status     PolicyStatus
}

func (p *Policy) Id() uint {
	return p.id
}

func (p *Policy) AccountID() uint {
	return p.accountID
}

func (p *Policy) PolicyType() string {
	return p.policyType
}

func (p *Policy) AccessKey() string {
	return p.accessKey
}

func (p *Policy) SecretKey() string {
	return p.secretKey
}

func (p *Policy) Bucket() string {
	return p.bucket
}

func (p *Policy) Area() string {
	return p.area
}

func (p *Policy) Detail() map[string]string {
	return p.detail
}

func (p *Policy) Status() PolicyStatus {
	return p.status
}

type PolicyFactory struct {
	fc PolicyFactoryConfig
}

type PolicyFactoryConfig struct {
	SupportedPolicyTypes []string
}

func (fc *PolicyFactoryConfig) validateConfig() error {
	if len(fc.SupportedPolicyTypes) == 0 {
		return fmt.Errorf("there's no supported policy types")
	}
	return nil
}

func NewPolicyFactory(fc PolicyFactoryConfig) (*PolicyFactory, error) {
	if err := fc.validateConfig(); err != nil {
		return nil, err
	}
	return &PolicyFactory{fc: fc}, nil
}

func (f *PolicyFactory) isPolicyTypeValid(policyType string) bool {
	for i := 0; i < len(f.fc.SupportedPolicyTypes); i++ {
		if policyType == f.fc.SupportedPolicyTypes[i] {
			return true
		}
	}
	return false
}

func (f *PolicyFactory) NewPolicy(accountID uint, policyType string, accessKey string, secretKey string,
	bucket string, area string, detail map[string]string) (*Policy, error) {
	if !f.isPolicyTypeValid(policyType) {
		return nil, slugerror.NewSlugError(slugerror.ErrInvalidInput, "invalid policy type",
			fmt.Sprintf("%s is not in supported policy types", policyType))
	}
	return &Policy{
		accountID:  accountID,
		policyType: policyType,
		accessKey:  accessKey,
		secretKey:  secretKey,
		bucket:     bucket,
		area:       area,
		detail:     detail,
		status:     PolicyOff,
	}, nil
}

func (p *Policy) SetAccessKey(accessKey string) {
	p.accessKey = accessKey
}

func (p *Policy) SetSecretKey(secretKey string) {
	p.secretKey = secretKey
}

func (p *Policy) SetBucket(bucket string) {
	p.bucket = bucket
}

func (p *Policy) SetArea(area string) {
	p.area = area
}

func (p *Policy) SetDetail(detail map[string]string) {
	p.detail = detail
}

// UnmarshallPolicy 从仓储实体映射回来领域实体，因为本函数不做参数验证和参数转换
func UnmarshallPolicy(id uint, accountID uint, policyType string, accessKey string,
	secretKey string, bucket string, area string, detail string, status PolicyStatus) (*Policy, error) {
	var detailMap map[string]string
	err := json.Unmarshal([]byte(detail), &detailMap)
	if err != nil {
		return nil, err
	}
	return &Policy{
		id:         id,
		accountID:  accountID,
		policyType: policyType,
		accessKey:  accessKey,
		secretKey:  secretKey,
		bucket:     bucket,
		area:       area,
		detail:     detailMap,
		status:     status,
	}, err
}
