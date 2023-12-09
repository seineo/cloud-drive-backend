package repo

import (
	"account/domain/account/repository"
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type CodeSuite struct {
	suite.Suite
	rdb  *redis.Client
	repo repository.CodeRepository
}

func TestCodeSuite(t *testing.T) {
	suite.Run(t, new(CodeSuite))
}

func (suite *CodeSuite) SetupSuite() {
	suite.rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	assert.NotNil(suite.T(), suite.rdb)
	repo, err := NewCodeRepo(suite.rdb, context.Background())
	assert.NoError(suite.T(), err)
	suite.repo = repo
}

func (suite *CodeSuite) TearDownSuite() {
	err := suite.rdb.Close()
	assert.NoError(suite.T(), err)
}

func (suite *CodeSuite) BeforeTest(suiteName, testName string) {
	// 插入测试数据
	ctx := context.Background()
	err1 := suite.rdb.Set(ctx, "codeKey1", "code1", 10*time.Minute).Err()
	assert.NoError(suite.T(), err1)
	err2 := suite.rdb.Set(ctx, "codeKey2", "code2", 10*time.Minute).Err()
	assert.NoError(suite.T(), err2)
}

func (suite *CodeSuite) AfterTest(suiteName, testName string) {
	// 删除测试数据库的所有数据
	ctx := context.Background()
	err := suite.rdb.FlushDB(ctx).Err()
	assert.NoError(suite.T(), err)
}

func (suite *CodeSuite) TestSetCode() {
	result, err := suite.rdb.Exists(context.Background(), "newCodeKey").Result()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), result, int64(0))

	err = suite.repo.SetCode("newCodeKey", "newCode", time.Minute)
	assert.NoError(suite.T(), err)

	result, err = suite.rdb.Exists(context.Background(), "newCodeKey").Result()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), result, int64(1))
}

func (suite *CodeSuite) TestGetCode() {
	code, err := suite.repo.GetCode("codeKey1")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), code, "code1")
	// 不存在键
	code, err = suite.repo.GetCode("not-exists")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), code, "")
}
