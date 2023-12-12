package eventbus

import "time"

type Event interface {
	GetID() int64
	GetName() string
	GetOccurTime() time.Time
	Marshall() ([]byte, error)
}

type EventStore interface {
	StoreEvent(event Event) error
	GetEvent(eventID int64) (Event, error)
}

type EventBase struct {
	EventID   int64     `json:"eventID"`
	EventName string    `json:"eventName"`
	OccurTime time.Time `json:"occurTime"`
}

func NewEventBase(eventName string) EventBase {
	return EventBase{
		EventID:   0,
		EventName: eventName,
		OccurTime: time.Now(),
	}
}
