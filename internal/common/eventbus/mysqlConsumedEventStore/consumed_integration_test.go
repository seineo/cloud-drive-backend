package mysqlConsumedEventStore

import (
	"common/dao"
	"common/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type ConsumedEventSuite struct {
	suite.Suite
	db    *gorm.DB
	store eventbus.ConsumedEventStore
}

var initialConsumedEvents = []*eventbus.ConsumedEvent{
	eventbus.NewConsumedEvent(1, "test1"),
	eventbus.NewConsumedEvent(2, "test2"),
}

func (suite *ConsumedEventSuite) SetupSuite() {
	// 连接数据库
	db, err := dao.InitMySQLConn("root:TestPassword123.@tcp(localhost:3306)/cloud_drive?parseTime=true")
	assert.NoError(suite.T(), err)
	suite.db = db
	suite.store, err = NewConsumedEventStore(db)
	assert.NoError(suite.T(), err)
}

func (suite *ConsumedEventSuite) TearDownSuite() {
	// 断开数据连接
	err := dao.CloseMySQLConn(suite.db)
	assert.NoError(suite.T(), err)
}

func (suite *ConsumedEventSuite) BeforeTest(suiteName, testName string) {
	// 创建表格
	err := suite.db.AutoMigrate(&MySQLConsumedEvent{})
	assert.NoError(suite.T(), err)
	// 插入测试数据
	for _, event := range initialConsumedEvents {
		err = suite.db.Exec("insert into consumed_events (id, value) values (?, ?)",
			event.EventID, event.Value).Error
		assert.NoError(suite.T(), err)
	}
}

func (suite *ConsumedEventSuite) AfterTest(suiteName, testName string) {
	// 删除表格
	err := suite.db.Migrator().DropTable(&MySQLConsumedEvent{})
	assert.NoError(suite.T(), err)
}

func TestConsumedEventStoreSuite(t *testing.T) {
	suite.Run(t, new(ConsumedEventSuite))
}

func (suite *ConsumedEventSuite) TestStoreConsumedEvent() {
	err := suite.store.StoreConsumedEvent(*eventbus.NewConsumedEvent(3, "test3"))
	assert.NoError(suite.T(), err)
	// 重复id
	err = suite.store.StoreConsumedEvent(*eventbus.NewConsumedEvent(1, "test4"))
	assert.Error(suite.T(), err)
	// 重复value
	err = suite.store.StoreConsumedEvent(*eventbus.NewConsumedEvent(5, "test3"))
	assert.NoError(suite.T(), err)
}
