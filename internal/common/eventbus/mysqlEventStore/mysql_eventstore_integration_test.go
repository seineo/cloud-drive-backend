package mysqlEventStore

import (
	"common/dao"
	"common/eventbus"
	"common/eventbus/account"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type MySQLEventSuite struct {
	suite.Suite
	db    *gorm.DB
	store eventbus.EventStore
}

var initialEvents = []eventbus.Event{
	account.NewCodeGeneratedEvent("1@test.com", "123456"),
	account.NewCodeGeneratedEvent("2@test.com", "234567"),
}

func (suite *MySQLEventSuite) SetupSuite() {
	// 连接数据库
	db, err := dao.InitMySQLConn("root:TestPassword123.@tcp(localhost:3306)/cloud_drive?parseTime=true")
	assert.NoError(suite.T(), err)
	suite.db = db
	suite.store, err = NewMySQLEventStore(db)
	assert.NoError(suite.T(), err)
}

func (suite *MySQLEventSuite) TearDownSuite() {
	// 断开数据连接
	err := dao.CloseMySQLConn(suite.db)
	assert.NoError(suite.T(), err)
}

func (suite *MySQLEventSuite) BeforeTest(suiteName, testName string) {
	// 创建表格
	err := suite.db.AutoMigrate(&Event{})
	assert.NoError(suite.T(), err)
	// 插入测试数据
	for _, event := range initialEvents {
		jsonEvent, err := event.Marshall()
		assert.NoError(suite.T(), err)
		err = suite.db.Exec("insert into events (id, status, value) values (?, ?, ?)",
			event.GetID(), eventbus.EventUnconsumed, string(jsonEvent)).Error
		assert.NoError(suite.T(), err)
	}
}

func (suite *MySQLEventSuite) AfterTest(suiteName, testName string) {
	// 删除表格
	err := suite.db.Migrator().DropTable(&Event{})
	assert.NoError(suite.T(), err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(MySQLEventSuite))
}

func (suite *MySQLEventSuite) TestStoreEvent() {
	err := suite.store.StoreEvent(account.NewCodeGeneratedEvent("99@test.com", "11111"))
	assert.NoError(suite.T(), err)
}

func (suite *MySQLEventSuite) TestSetEventConsumed() {
	// 获取initialEvents[0]，并修改其状态
	initialEvent := initialEvents[0]
	err := suite.store.SetEventConsumed(initialEvent.GetID())
	assert.NoError(suite.T(), err)
}

func (suite *MySQLEventSuite) TestGetUnconsumedEvents() {
	events, err := suite.store.GetUnconsumedEvents()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(events))
	for i, event := range events {
		var codeEvent account.CodeGenerated
		err = json.Unmarshal([]byte(event), &codeEvent)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), codeEvent.EventID, initialEvents[i].GetID())
	}
}
