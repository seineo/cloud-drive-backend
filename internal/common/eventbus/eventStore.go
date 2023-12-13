package eventbus

type EventStore interface {
	StoreEvent(event Event) error
	SetEventProcessed(eventID int64) error
	GetUnprocessedEvents() ([]string, error) // 每个event是字符串形式
}

type EventStatus string

const EventProcessed = "processed"
const EventUnprocessed = "unprocessed"
