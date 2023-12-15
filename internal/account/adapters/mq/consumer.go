package mq

import (
	"account/config"
	"common/eventbus"
	"common/eventbus/email"
	"common/eventbus/kafkaEventManager"
	"common/eventbus/mysqlEventStore"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

type MQConsumer struct { // 作为发布方，接受消费方的回应消息
	consumer   eventbus.Consumer
	eventStore eventbus.EventStore
}

func NewMQConsumer(configs *config.Config) *MQConsumer {
	// kafka注入
	mechanism, err := scram.Mechanism(scram.SHA256, configs.KafkaUsername, configs.KafkaPassword)
	if err != nil {
		log.Fatalln(err)
	}
	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}
	// mysql
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true",
		configs.DBUser, configs.DBPassword, configs.DBProtocol, configs.DBAddr, configs.DBDatabase)
	logrus.WithFields(logrus.Fields{
		"dsn": dsn,
	}).Debug(" dsn")
	mysqlDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("unable to connect mysqlEventStore database")
	}
	eventStore, err := mysqlEventStore.NewMySQLEventStore(mysqlDB)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create event store")
	}
	mqConsumer := &MQConsumer{}
	mqConsumer.consumer = kafkaEventManager.NewEventConsumer(dialer, []string{configs.KafkaBroker})
	mqConsumer.eventStore = eventStore
	return mqConsumer
}

// TODO event handler应该都放在application的service中
func (m *MQConsumer) codeEmailHandler(eventBytes []byte, eventData map[string]interface{}) {
	logrus.Infof("get event: %v", eventData)
	eventName, exists := eventData["eventName"]
	if !exists {
		logrus.Error("eventName is not as a key in event")
		return
	}
	if eventName == "codeEmailSent" {
		codeEmailEvent := email.CodeEmailSent{}
		err := json.Unmarshal(eventBytes, &codeEmailEvent)
		if err != nil {
			logrus.WithError(err).Errorln("unable to unmarshal event to codeEmailSent")
			return
		}
		// 设置codeEmailEvent事件已消费
		if err := m.eventStore.SetEventConsumed(codeEmailEvent.ConsumedEventID); err != nil {
			logrus.WithError(err).Errorln("unable to set event consumed")
			return
		}
	}
}

func (m *MQConsumer) Run() {
	m.consumer.Subscribe("email", m.codeEmailHandler)
	err := m.consumer.StartConsuming("email", time.Now())
	if err != nil {
		logrus.WithError(err).Fatalln("event consume error")
		return
	}
}
