package repo

import (
	"common/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"resource/domain/entity"
	"resource/domain/repository"
	"testing"
)

type FolderSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.FolderRepo
}

func (suite *FolderSuite) SetupSuite() {
	// 连接数据库
	db, err := dao.InitMySQLConn("root:TestPassword123.@tcp(localhost:3306)/cloud_drive?parseTime=true")
	assert.NoError(suite.T(), err)
	suite.db = db
}

func (suite *FolderSuite) TearDownSuite() {
	// 断开数据连接
	err := dao.CloseMySQLConn(suite.db)
	assert.NoError(suite.T(), err)
}

func (suite *FolderSuite) BeforeTest(suiteName, testName string) {
	// 创建表格
	err := suite.db.AutoMigrate(&Folder{})
	assert.NoError(suite.T(), err)
	repo, err := NewFolderRepo(suite.db)
	assert.NoError(suite.T(), err)
	suite.repo = repo

	// 插入测试数据
	parentID := uint(1)
	initialFolders := []*entity.Folder{
		entity.NewFolder(1, 1, nil, "root"),
		entity.NewFolder(1, 1, &parentID, "sub1"),
	}
	for _, folder := range initialFolders {
		err = suite.db.Exec("insert into folders (account_id, policy_id, parent_folder, name) values (?, ?, ?, ?)",
			folder.AccountID(), folder.PolicyID(), folder.ParentFolder(), folder.Name()).Error
		assert.NoError(suite.T(), err)
	}
}

func (suite *FolderSuite) AfterTest(suiteName, testName string) {
	// 删除表格
	err := suite.db.Migrator().DropTable(&Folder{})
	assert.NoError(suite.T(), err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(FolderSuite))
}

func (suite *FolderSuite) TestCreateFolder() {
	parentFolder := uint(1)
	folder1, err1 := suite.repo.CreateFolder(*entity.NewFolder(1, 1, &parentFolder, "sub2"))
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), uint(3), folder1.Id())
	assert.Equal(suite.T(), uint(3), folder1.Id())

}

func (suite *FolderSuite) TestGetSubFolders() {
	parentFolder := uint(1)
	folder1, err1 := suite.repo.CreateFolder(*entity.NewFolder(1, 1, &parentFolder, "sub2"))
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), uint(3), folder1.Id())
	folder2, err2 := suite.repo.CreateFolder(*entity.NewFolder(1, 1, &parentFolder, "sub2"))
	assert.NoError(suite.T(), err2)
	assert.Equal(suite.T(), uint(4), folder2.Id())

	// 确定根目录有3个子目录
	folders, err3 := suite.repo.GetSubFolders(1)
	assert.NoError(suite.T(), err3)
	assert.Equal(suite.T(), 3, len(folders))
}
