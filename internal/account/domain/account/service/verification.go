package service

import (
	"account/domain/account/entity"
	"account/domain/account/repository"
	"common/eventbus"
	"time"
)

type VerificationService interface {
	GenerateAuthCode(email string, expiration time.Duration) (string, error)
	GetAuthCode(email string) (string, error)
}

type verificationService struct {
	codeRepo      repository.CodeRepository
	factory       *entity.CodeFactory // 验证码工厂中有mutex锁，需要用指针传递
	eventProducer eventbus.Producer
}

func NewVerificationService(codeRepo repository.CodeRepository, factory *entity.CodeFactory, eventProducer eventbus.Producer) VerificationService {
	return &verificationService{codeRepo: codeRepo, eventProducer: eventProducer, factory: factory}
}

func (v *verificationService) GenerateAuthCode(email string, expiration time.Duration) (string, error) {
	// 生成code
	codeObj := v.factory.NewVerificationCode(email)
	// 存储code
	if err := v.codeRepo.SetCode(email, codeObj.Get(), expiration); err != nil {
		return "", err
	}
	// TODO 存储领域事件到事件表，并清空entity的事件
	for _, event := range codeObj.GetEvents() {
		if err := v.eventProducer.Publish("account", event); err != nil {
			return "", err
		}
	}
	return codeObj.Get(), nil
}

func (v *verificationService) GetAuthCode(email string) (string, error) {
	return v.codeRepo.GetCode(email)
}
