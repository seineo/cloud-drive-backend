package main

import (
	"account/adapters/http"
	"account/adapters/mq"
	"account/application"
	"account/config"
	"common/eventbus/kafkaEventManager"
	"common/eventbus/mysqlEventStore"
	"common/logs"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func CronPublishEvents(configs *config.Config) {
	// kafka注入
	var dialer *kafka.Dialer
	if configs.KafkaUsername == "" {
		logrus.Info("use plain mechanism here")
		dialer = &kafka.Dialer{}
	} else {
		mechanism, err := scram.Mechanism(scram.SHA256, configs.KafkaUsername, configs.KafkaPassword)
		if err != nil {
			logrus.Fatalln(err)
		}
		dialer = &kafka.Dialer{
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		}
	}
	// mysql
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=true",
		configs.DBUser, configs.DBPassword, configs.DBProtocol, configs.DBAddr, configs.DBDatabase)
	logrus.WithFields(logrus.Fields{
		"dsn": dsn,
	}).Debug(" dsn")
	mysqlDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("fail to connect mysqlEventStore database")
	}
	eventProducer := kafkaEventManager.NewEventProducer(dialer, []string{configs.KafkaBroker})
	eventStore, err := mysqlEventStore.NewMySQLEventStore(mysqlDB)
	if err != nil {
		logrus.WithError(err).Fatal("unable to set up event store")
	}
	// 运行定时任务
	cronEventManager := application.NewCronEventManager(eventProducer, eventStore)
	c := cron.New()
	c.AddFunc("@every 10s", cronEventManager.PublishEvents)
	c.Start()
}

func main() {
	logs.Init()
	// 读取配置
	configs, err := config.LoadConfig("config")
	if err != nil {
		logrus.Fatal(err.Error())
	}
	// 定时检查领域事件并发送
	CronPublishEvents(configs)
	// 消费消息队列的数据
	mqConsumer := mq.NewMQConsumer(configs)
	go mqConsumer.Run()
	// 运行http服务器
	engine := gin.Default()
	httpServer := http.NewHttpServer(configs, engine)
	httpServer.Run()
}
