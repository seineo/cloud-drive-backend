package kafkaEventManager

import (
	"common/eventbus"
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type EventProducer struct {
	dialer  *kafka.Dialer
	brokers []string
}

func (e *EventProducer) Publish(topic string, eventBytes []byte) error {
	w := kafka.NewWriter(kafka.WriterConfig{
		Dialer:  e.dialer,
		Brokers: e.brokers,
		Topic:   topic,
	})
	defer w.Close()
	ctx := context.Background()
	err := w.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
	})
	//conn, err := kafka.DialLeader(context.Background(), "tcp", "factual-marmot-8450-us1-kafka.upstash.io:9092", topic, 0)
	//if err != nil {
	//	return err
	//}
	//_, err = conn.WriteMessages(kafka.Message{Value: eventBytes})
	if err != nil {
		return err
	}
	logrus.Infof("publish event: %v", string(eventBytes))
	return nil
}

func NewEventProducer(dialer *kafka.Dialer, brokers []string) eventbus.Producer {
	return &EventProducer{dialer: dialer, brokers: brokers}
}
