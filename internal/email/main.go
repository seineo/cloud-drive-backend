package main

import (
	kafkaEventManager "common/eventbus/kafka"
	"common/logs"
	"crypto/tls"
	"email/application"
	"email/config"
	"email/domain/service"
	"email/infrastructure"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
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
	// kafka
	mechanism, err := scram.Mechanism(scram.SHA256, configs.KafkaUsername, configs.KafkaPassword)
	if err != nil {
		log.Fatalln(err)
	}
	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	consumer := kafkaEventManager.NewEventConsumer(dialer, []string{configs.KafkaBroker})
	as := application.NewApplicationEmail(emailService, consumer)
	as.Run()
}
