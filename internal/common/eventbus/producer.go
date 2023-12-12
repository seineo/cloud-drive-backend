package eventbus

type Producer interface {
	Publish(topic string, event Event) error
}
