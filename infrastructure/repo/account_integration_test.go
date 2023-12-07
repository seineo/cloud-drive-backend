package repo

import (
	"CloudDrive/common/dao"
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type AccountSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.AccountRepo
}

func (suite *AccountSuite) SetupSuite() {
	// 连接数据库
	db, err := dao.InitMySQLConn("root:TestPassword123.@tcp(localhost:3306)/cloud_drive?parseTime=true")
	assert.NoError(suite.T(), err)
	suite.db = db
}

func (suite *AccountSuite) TearDownSuite() {
	// 断开数据连接
	err := dao.CloseMySQLConn(suite.db)
	assert.NoError(suite.T(), err)
}

func (suite *AccountSuite) BeforeTest(suiteName, testName string) {
	// 创建表格
	err := suite.db.AutoMigrate(&account{})
	assert.NoError(suite.T(), err)
	repo, err := NewAccountRepo(suite.db)
	assert.NoError(suite.T(), err)
	suite.repo = repo
	// 插入测试数据
	initialAccounts := []*entity.Account{
		entity.NewAccountWithID(1, "1@test.com", "1", "1"),
		entity.NewAccountWithID(2, "2@test.com", "2", "2"),
	}
	for _, accP := range initialAccounts {
		suite.T().Logf("account: %v\n", *accP)
		err := suite.db.Exec("insert into accounts (email, nickname, password) values (?, ?, ?)",
			accP.GetEmail(), accP.GetNickname(), accP.GetPassword()).Error
		assert.NoError(suite.T(), err)
	}
}

func (suite *AccountSuite) AfterTest(suiteName, testName string) {
	// 删除表格
	err := suite.db.Migrator().DropTable(&account{})
	assert.NoError(suite.T(), err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(AccountSuite))
}

func (suite *AccountSuite) TestCreateAccount() {
	acc, err := suite.repo.Create(*entity.NewAccountWithID(3, "3@test.com", "3", "3"))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint(3), acc.GetID())

	// 使用重复邮箱创建
	accountDuplicate, err := suite.repo.Create(*entity.NewAccountWithID(0, "1@test.com", "test", "test"))
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), (*entity.Account)(nil), accountDuplicate)
}

func (suite *AccountSuite) TestQueryAccount() {
	account1, err := suite.repo.Get(1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "1@test.com", account1.GetEmail())

	// 使用不存在的id查询
	accountErr, err := suite.repo.Get(99)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), (*entity.Account)(nil), accountErr)

	account1, err = suite.repo.GetByEmail("1@test.com")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "1", account1.GetNickname())

	// 邮件不存在时，报错
	accountEmpty, err := suite.repo.GetByEmail("99")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), (*entity.Account)(nil), accountEmpty)
}

func (suite *AccountSuite) TestUpdateAccount() {
	toUpdate1 := entity.NewAccountWithID(1, "1@test.com", "newName", "newPassword")
	_, err := suite.repo.Update(*toUpdate1)
	assert.NoError(suite.T(), err)
	updated1, err := suite.repo.Get(toUpdate1.GetID())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), *toUpdate1, *updated1)

	// 使用不存在的id更新，应该不报错，但是后面查询不到该id的记录
	toUpdate2 := entity.NewAccountWithID(99, "1@test.com", "newName", "newPassword")
	_, err = suite.repo.Update(*toUpdate1)
	assert.NoError(suite.T(), err)
	updated2, err := suite.repo.Get(toUpdate2.GetID())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), (*entity.Account)(nil), updated2)
}

func (suite *AccountSuite) TestDeleteAccount() {
	err := suite.repo.Delete(1)
	assert.NoError(suite.T(), err)
	err = suite.repo.Delete(2)
	assert.NoError(suite.T(), err)

	// 使用不存在的id删除，会报错unprocessable
	err = suite.repo.Delete(99)
	assert.Error(suite.T(), err)
}
