package eventbus

import (
	"common/server"
	"time"
)

type Event interface {
	GetID() int64
	GetName() string
	GetOccurTime() time.Time
	Marshall() ([]byte, error)
}

type EventBase struct {
	EventID   int64     `json:"eventID"`
	EventName string    `json:"eventName"`
	OccurTime time.Time `json:"occurTime"`
}

func NewEventBase(eventName string) EventBase {
	return EventBase{
		EventID:   server.GenerateID(),
		EventName: eventName,
		OccurTime: time.Now(),
	}
}
