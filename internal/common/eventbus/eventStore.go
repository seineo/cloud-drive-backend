package eventbus

type EventStore interface {
	StoreEvent(event Event) error
	SetEventConsumed(eventID int64) error
	GetUnconsumedEvents() ([]string, error) // 每个event是字符串形式
}

type EventStatus string

const EventConsumed = "consumed"
const EventUnconsumed = "unconsumed"
