package kafka

import (
	"common/eventbus"
	"context"
	"encoding/json"
	"errors"
	"github.com/segmentio/kafka-go"
	"time"
)

type EventConsumer struct {
	dialer     *kafka.Dialer
	brokers    []string
	handlerMap map[string][]eventbus.EventHandler
	reader     *kafka.Reader
}

func (e *EventConsumer) Stop() error {
	err := e.reader.Close()
	if err != nil {
		return err
	}
	return nil
}

func (e *EventConsumer) GetTopics() []string {
	keys := make([]string, len(e.handlerMap))
	i := 0
	for k := range e.handlerMap {
		keys[i] = k
		i++
	}
	return keys
}

func (e *EventConsumer) Subscribe(topic string, handler eventbus.EventHandler) {
	e.handlerMap[topic] = append(e.handlerMap[topic], handler)
}

func (e *EventConsumer) StartConsuming(topic string, startTime time.Time) error {
	var errs error
	e.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: e.brokers,
		Topic:   topic,
		Dialer:  e.dialer,
	})
	defer func(r *kafka.Reader) {
		err := r.Close()
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}(e.reader)
	// 只消费从startTime开始的消息
	if err := e.reader.SetOffsetAt(context.Background(), startTime); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second) // 每个使用context的函数独立计时
	defer cancelFunc()
	for {
		message, readErr := e.reader.ReadMessage(ctx) // 阻塞直到新消息，或超时，或关闭读取
		if readErr != nil {
			errs = errors.Join(errs, readErr)
			break
		}
		// 将这个消息unmarshal为map
		var eventData map[string]interface{}
		jsonErr := json.Unmarshal(message.Value, &eventData)
		if jsonErr != nil {
			errs = errors.Join(errs, jsonErr)
			break
		}
		// 输送消息给每个订阅该topic的handler
		for _, handler := range e.handlerMap[topic] {
			handler(eventData)
		}
	}
	return errs
}

func NewEventConsumer(dialer *kafka.Dialer, brokers []string) eventbus.Consumer {
	return &EventConsumer{
		dialer:     dialer,
		brokers:    brokers,
		handlerMap: make(map[string][]eventbus.EventHandler),
	}
}
