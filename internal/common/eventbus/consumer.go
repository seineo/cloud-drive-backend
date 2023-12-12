package eventbus

import "time"

type EventHandler func(eventBytes []byte, eventData map[string]interface{})

// Consumer 一个topic只有一个消费者的情况， 多个需要考虑并发（可以另开一个接口）
type Consumer interface {
	Subscribe(topic string, handler EventHandler)
	GetTopics() []string
	StartConsuming(topic string, startTime time.Time) error // 按topic订阅消费，需要处理多个topic，则需要自行开多个协程
	Stop() error
}
