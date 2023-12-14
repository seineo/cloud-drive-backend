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
	codeRepo   repository.CodeRepository
	factory    *entity.CodeFactory // 验证码工厂中有mutex锁，需要用指针传递
	eventStore eventbus.EventStore
}

func NewVerificationService(codeRepo repository.CodeRepository, factory *entity.CodeFactory,
	eventStore eventbus.EventStore) VerificationService {
	return &verificationService{codeRepo: codeRepo, factory: factory, eventStore: eventStore}
}

// GenerateAuthCode 生成验证码，并存储，以及发布领域事件。 （需要事务）
func (v *verificationService) GenerateAuthCode(email string, expiration time.Duration) (string, error) {
	// 生成code
	codeObj := v.factory.NewVerificationCode(email)

	// 存储领域事件，而后清空entity的事件
	event := codeObj.GetEvent()
	if err := v.eventStore.StoreEvent(event); err != nil {
		return "", err
	}
	codeObj.ClearEvent()
	// 存储code  （如果是redis这不同的数据库，则需要放在最后，方便回滚，否则需要额外的补偿操作）
	if err := v.codeRepo.SetCode(email, codeObj.Get(), expiration); err != nil {
		return "", err
	}
	return codeObj.Get(), nil
}

func (v *verificationService) GetAuthCode(email string) (string, error) {
	return v.codeRepo.GetCode(email)
}
