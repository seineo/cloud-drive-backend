package application

import (
	"common/eventbus"
	"github.com/sirupsen/logrus"
)

type CronEventManager interface {
	PublishEvents()
}

type cronEventManager struct {
	eventProducer eventbus.Producer
	eventStore    eventbus.EventStore
}

func NewCronEventManager(eventProducer eventbus.Producer, eventStore eventbus.EventStore) CronEventManager {
	return &cronEventManager{eventProducer: eventProducer, eventStore: eventStore}
}

func (c *cronEventManager) PublishEvents() {
	// 获取没有处理的事件
	eventStrs, err := c.eventStore.GetUnconsumedEvents()
	if err != nil {
		logrus.WithError(err).Error("unable to get unconsumed events")
	}
	logrus.Infof("number of unconsumed events: %v", len(eventStrs))
	// 逐个发布
	for _, eventStr := range eventStrs {
		err := c.eventProducer.Publish("account", []byte(eventStr))
		if err != nil {
			logrus.WithError(err).Error("unable to publish events")
		}
	}
}
