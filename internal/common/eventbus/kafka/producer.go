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
	defer w.Close()
	ctx := context.Background()
	eventBytes, err := event.Marshall()
	if err != nil {
		return err
	}
	err = w.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewEventProducer(dialer *kafka.Dialer, brokers []string) eventbus.Producer {
	return &EventProducer{dialer: dialer, brokers: brokers}
}
