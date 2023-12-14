package mysqlEventStore

import (
	"common/eventbus"
	"gorm.io/gorm"
)

type EventStore struct {
	db *gorm.DB
}

func NewMySQLEventStore(db *gorm.DB) (eventbus.EventStore, error) {
	if db == nil {
		panic("missing db")
	}
	if err := db.AutoMigrate(Event{}); err != nil {
		return nil, err
	}
	return &EventStore{db: db}, nil
}

type Event struct {
	ID     int64                `gorm:"primaryKey"`
	Status eventbus.EventStatus `gorm:"not null;index"`
	Value  string               `gorm:"not null"`
}

//// TableName 自定义表名
//func (Event) TableName() string {
//	return "events"
//}

func toStrings(mysqlEvents []Event) []string {
	var events []string
	for _, mysqlEvent := range mysqlEvents {
		events = append(events, mysqlEvent.Value)
	}
	return events
}

func (m *EventStore) StoreEvent(event eventbus.Event) error {
	eventBytes, err := event.Marshall()
	if err != nil {
		return err
	}
	err = m.db.Create(&Event{
		ID:     event.GetID(),
		Status: eventbus.EventUnconsumed,
		Value:  string(eventBytes),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *EventStore) SetEventConsumed(eventID int64) error {
	err := m.db.Model(&Event{}).Where("id = ?", eventID).Update("status", eventbus.EventConsumed).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *EventStore) GetUnconsumedEvents() ([]string, error) {
	var events []Event
	if err := m.db.Where("status = ?", eventbus.EventUnconsumed).Find(&events).Error; err != nil {
		return nil, err
	}
	return toStrings(events), nil
}
