package kafka

import (
	"common/eventbus"
	"context"
	"github.com/segmentio/kafka-go"
)

type EventProducer struct {
	dialer  *kafka.Dialer
	brokers []string
}

func (e *EventProducer) Publish(topic string, event eventbus.Event) error {
	w := kafka.NewWriter(kafka.WriterConfig{
		Dialer:  e.dialer,
		Brokers: e.brokers,
		Topic:   topic,
	})
	ctx := context.Background()
	eventBytes, err := event.Marshall()
	if err != nil {
		return err
	}
	w.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
	})
	w.Close()
	return nil
}

func NewEventProducer(dialer *kafka.Dialer, brokers []string) eventbus.Producer {
	return &EventProducer{dialer: dialer, brokers: brokers}
}
