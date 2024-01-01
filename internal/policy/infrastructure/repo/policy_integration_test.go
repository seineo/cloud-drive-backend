package repo

import (
	"common/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"policy/domain/entity"
	"policy/domain/repository"
	"testing"
)

type PolicySuite struct {
	suite.Suite
	db         *gorm.DB
	policyRepo repository.PolicyRepo
}

func TestPolicySuite(t *testing.T) {
	suite.Run(t, new(PolicySuite))
}

func (suite *PolicySuite) SetupSuite() {
	// 连接数据库
	db, err := dao.InitMySQLConn("root:TestPassword123.@tcp(localhost:3306)/cloud_drive?parseTime=true")
	assert.NoError(suite.T(), err)
	suite.db = db
}

func (suite *PolicySuite) TearDownSuite() {
	// 断开数据连接
	err := dao.CloseMySQLConn(suite.db)
	assert.NoError(suite.T(), err)
}

func (suite *PolicySuite) BeforeTest(suiteName, testName string) {
	// 创建表格
	repo, err := NewPolicyRepo(suite.db)
	assert.NoError(suite.T(), err)
	suite.policyRepo = repo
	// 插入初始数据
	policyFc := entity.PolicyFactoryConfig{SupportedPolicyTypes: []string{"qiniu"}}
	policyFactory, err := entity.NewPolicyFactory(policyFc)
	assert.NoError(suite.T(), err)
	policy1, err := policyFactory.NewPolicy(1, "qiniu", "access1", "secret1",
		"bucket1", "area1", nil)
	assert.NoError(suite.T(), err)
	policy2, err := policyFactory.NewPolicy(2, "qiniu", "access2", "secret2",
		"bucket2", "area2", nil)
	assert.NoError(suite.T(), err)
	initialPolicies := []*entity.Policy{policy1, policy2}
	for _, policy := range initialPolicies {
		err := suite.db.Exec("insert into policies (account_id, policy_type, access_key, secret_key, bucket, area, status) "+
			"values (?, ?, ?, ?, ?, ?, ?)", policy.AccountID(), policy.PolicyType(), policy.AccessKey(), policy.SecretKey(),
			policy.Bucket(), policy.Area(), entity.PolicyOff).Error
		assert.NoError(suite.T(), err)
	}
}

func (suite *PolicySuite) AfterTest(suiteName, testName string) {
	err := suite.db.Migrator().DropTable(&Policy{})
	assert.NoError(suite.T(), err)
}

func (suite *PolicySuite) TestCreatePolicy() {
	// 正常数据
	policyFc := entity.PolicyFactoryConfig{SupportedPolicyTypes: []string{"qiniu"}}
	policyFactory, err := entity.NewPolicyFactory(policyFc)
	assert.NoError(suite.T(), err)

	policy, err := policyFactory.NewPolicy(3, "qiniu", "access3", "secret3",
		"bucket3", "area3", nil)
	assert.NoError(suite.T(), err)
	mysqlPolicy, err := suite.policyRepo.CreatePolicy(*policy)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint(3), mysqlPolicy.Id())

	// 相同账户已有该policy类型
	policy, err = policyFactory.NewPolicy(3, "qiniu", "access4", "secret4",
		"bucket4", "area4", nil)
	assert.NoError(suite.T(), err)
	_, err = suite.policyRepo.CreatePolicy(*policy)
	assert.Equal(suite.T(), err.Error(), "invalid policy type")
}
