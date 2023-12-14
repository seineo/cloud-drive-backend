package eventbus

type Producer interface {
	Publish(topic string, eventBytes []byte) error
}
