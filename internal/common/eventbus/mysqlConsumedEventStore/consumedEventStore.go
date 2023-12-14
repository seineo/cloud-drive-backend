package mysqlConsumedEventStore

import (
	"common/eventbus"
	"gorm.io/gorm"
)

type ConsumedEventStore struct {
	db *gorm.DB
}

type MySQLConsumedEvent struct {
	ID    int64  `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

func (MySQLConsumedEvent) TableName() string {
	return "consumed_events"
}

func (c *ConsumedEventStore) StoreConsumedEvent(event eventbus.ConsumedEvent) error {
	consumedEvent := MySQLConsumedEvent{
		ID:    event.EventID,
		Value: event.Value,
	}
	return c.db.Create(&consumedEvent).Error
}

func NewConsumedEventStore(db *gorm.DB) (eventbus.ConsumedEventStore, error) {
	if db == nil {
		panic("missing db")
	}
	return &ConsumedEventStore{db: db}, nil
}
