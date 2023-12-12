package application

import (
	"common/eventbus"
	"common/eventbus/account"
	"email/domain/service"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
)

type EmailApplicationService interface {
	Run()
}

type applicationEmail struct {
	emailService  service.EmailService
	eventConsumer eventbus.Consumer
}

func NewApplicationEmail(emailService service.EmailService, eventConsumer eventbus.Consumer) EmailApplicationService {
	return &applicationEmail{emailService: emailService, eventConsumer: eventConsumer}
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
		// 通过领域服务发送邮件
		err = a.emailService.SendVerificationCode(codeEvent.Email, codeEvent.Code)
		if err != nil {
			logrus.WithError(err).Errorln("unable to send verification code")
			return
		}
	}
}

func (a *applicationEmail) Run() {
	logrus.Info("email service starts……")
	a.eventConsumer.Subscribe("account", a.codeHandler)
	err := a.eventConsumer.StartConsuming("account", time.Now().Add(-time.Hour))
	if err != nil {
		logrus.WithError(err).Fatalln("event consume error")
		return
	}
}
