package service

import (
	"CloudDrive/common/validation"
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"time"
)

type VerificationService interface {
	SendAuthCode(email string, expiration time.Duration) (string, error)
	GetAuthCode(codeKey string) (string, error)
}

type verificationService struct {
	codeRepo repository.CodeRepository
	factory  *entity.CodeFactory // 验证码工厂中有mutex锁，需要用指针传递
}

func (v *verificationService) SendAuthCode(email string, expiration time.Duration) (string, error) {
	// 编码email和时间得到codeKey
	codeKey := validation.SHA256Hash(email, time.Now().String())
	// 生成code
	code := v.factory.NewVerificationCode()
	// 存储code
	if err := v.codeRepo.SetCode(codeKey, code.Get(), expiration); err != nil {
		return "", err
	}
	// TODO 领域事件：发送验证码邮件
	return code.Get(), nil
}

func (v *verificationService) GetAuthCode(codeKey string) (string, error) {
	return v.codeRepo.GetCode(codeKey)
}

func NewVerificationService(codeRepo repository.CodeRepository, factory *entity.CodeFactory) VerificationService {
	return &verificationService{
		codeRepo: codeRepo,
		factory:  factory,
	}
}
