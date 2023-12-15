package eventbus

type ConsumedEventStore interface {
	StoreConsumedEvent(event ConsumedEvent) error
}

type ConsumedEvent struct {
	EventID int64
	Value   string
}

func NewConsumedEvent(eventID int64, value string) *ConsumedEvent {
	return &ConsumedEvent{EventID: eventID, Value: value}
}
