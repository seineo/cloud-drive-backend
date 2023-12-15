package main

import (
	"common/eventbus/kafkaEventManager"
	"common/logs"
	"crypto/tls"
	"email/application"
	"email/config"
	"email/domain/service"
	"email/infrastructure"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	logs.Init()
	// 读取配置
	configs, err := config.LoadConfig("config")
	if err != nil {
		logrus.Fatal(err.Error())
	}
	// gomail
	mailDialer := gomail.NewDialer(configs.SMTPAddr, configs.SMTPPort, configs.SMTPUser, configs.SMTPPassword)
	emailSender := infrastructure.NewGoMailer(mailDialer)

	emailService := service.NewEmailService(map[string]string{"code": configs.SMTPSender}, emailSender, configs)
	// kafkaEventManager
	mechanism, err := scram.Mechanism(scram.SHA256, configs.KafkaUsername, configs.KafkaPassword)
	if err != nil {
		log.Fatalln(err)
	}
	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	consumer := kafkaEventManager.NewEventConsumer(dialer, []string{configs.KafkaBroker})
	producer := kafkaEventManager.NewEventProducer(dialer, []string{configs.KafkaBroker})

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
	as := application.NewApplicationEmail(emailService, consumer, producer, mysqlDB)
	as.Run()
}
