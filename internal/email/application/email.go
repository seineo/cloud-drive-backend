package application

import (
	"common/eventbus"
	"common/eventbus/account"
	"common/eventbus/email"
	"common/eventbus/mysqlConsumedEventStore"
	"email/domain/service"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type EmailApplicationService interface {
	Run()
}

type applicationEmail struct {
	emailService  service.EmailService
	eventConsumer eventbus.Consumer
	eventProducer eventbus.Producer
	mysqlDB       *gorm.DB
}

func NewApplicationEmail(emailService service.EmailService, eventConsumer eventbus.Consumer,
	eventProducer eventbus.Producer, mysqlDB *gorm.DB) EmailApplicationService {
	return &applicationEmail{emailService: emailService, eventConsumer: eventConsumer, eventProducer: eventProducer, mysqlDB: mysqlDB}
}

func respondConsumedCodeEvent(producer eventbus.Producer, codeEvent account.CodeGenerated) error {
	respEvent := email.NewCodeEmailSentEvent(codeEvent.Email, codeEvent.Code, codeEvent.GetID())
	respEventBytes, err := respEvent.Marshall()
	if err != nil {
		return err
	}
	err = producer.Publish("email", respEventBytes)
	if err != nil {
		return err
	}
	return nil
}

func (a *applicationEmail) codeHandler(eventBytes []byte, eventData map[string]interface{}) {
	logrus.Infof("get event: %v", eventData)
	eventName, exists := eventData["eventName"]
	if !exists {
		logrus.Error("eventName is not as a key in event")
		return
	}
	if eventName == "codeGenerated" {
		codeEvent := account.CodeGenerated{}
		err := json.Unmarshal(eventBytes, &codeEvent)
		if err != nil {
			logrus.WithError(err).Errorln("unable to unmarshal event to codeGenerated")
			return
		}
		// TODO 在TODO1和TODO2之间要设置事务
		// 查询与插入event到消费事件表，如果消费过，发布一个消费过的通知
		tx := a.mysqlDB.Begin()
		consumedEventStore, err := mysqlConsumedEventStore.NewConsumedEventStore(tx)
		if err != nil {
			logrus.WithError(err).Error("unable to create consumed event store")
			tx.Rollback()
			return
		}
		err = consumedEventStore.StoreConsumedEvent(eventbus.ConsumedEvent{
			EventID: codeEvent.GetID(),
			Value:   string(eventBytes),
		})
		if err != nil {
			mysqlErr := err.(*mysql.MySQLError)
			if mysqlErr.Number == uint16(1062) { // 重复消息
				respErr := respondConsumedCodeEvent(a.eventProducer, codeEvent)
				if respErr != nil {
					logrus.WithError(respErr).Error("unable to respond with consumed codeGenerated event")
				}
				return // 只要重复，无论是否回复成功都要结束
			} else {
				logrus.WithError(err).Error("unable to store consumed event")
				return
			}
		}
		// 通过领域服务发送邮件
		err = a.emailService.SendVerificationCode(codeEvent.Email, codeEvent.Code)
		if err != nil {
			logrus.WithError(err).Errorln("unable to send verification code")
			tx.Rollback()
			return
		}
		if err := tx.Commit().Error; err != nil {
			logrus.WithError(err).Errorln("unable to set commit transaction")
			return
		}
		// 告知发布方已消费
		respErr := respondConsumedCodeEvent(a.eventProducer, codeEvent)
		if respErr != nil {
			logrus.WithError(respErr).Error("unable to respond with consumed codeGenerated event")
		}
	}
}

func (a *applicationEmail) Run() {
	logrus.Info("email service starts……")
	a.eventConsumer.Subscribe("account", a.codeHandler)
	err := a.eventConsumer.StartConsuming("account", time.Now())
	if err != nil {
		logrus.WithError(err).Fatalln("event consume error")
		return
	}
}
