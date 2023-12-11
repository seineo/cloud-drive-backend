package application

import "common/eventbus"

type EmailService interface {
	Run()
}

type applicationEmail struct {
	eventConsumer eventbus.Consumer
}

func (a *applicationEmail) Run() {
	// TODO implement me
	panic("implement me")

	// TODO 解析MQ传过来的jsonData的Type，根据Type来调用函数
}
