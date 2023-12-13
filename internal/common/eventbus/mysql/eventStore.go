package mysql

import (
	"common/eventbus"
	"gorm.io/gorm"
)

type EventStore struct {
	db *gorm.DB
}

func NewMySQLEventStore(db *gorm.DB) eventbus.EventStore {
	return &EventStore{db: db}
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
		Status: eventbus.EventUnprocessed,
		Value:  string(eventBytes),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *EventStore) SetEventProcessed(eventID int64) error {
	err := m.db.Model(&Event{}).Where("id = ?", eventID).Update("status", eventbus.EventProcessed).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *EventStore) GetUnprocessedEvents() ([]string, error) {
	var events []Event
	if err := m.db.Where("status = ?", eventbus.EventUnprocessed).Find(&events).Error; err != nil {
		return nil, err
	}
	return toStrings(events), nil
}
