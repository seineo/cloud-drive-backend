package service

import (
	"account/adapters/http/types"
	"account/domain/account/entity"
	"account/domain/account/service"
	"account/infrastructure/repo"
	"common/eventbus/mysqlEventStore"
	"common/slugerror"
	"context"
	"github.com/alexedwards/argon2id"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
	"time"
)

var DefaultVerificationService service.VerificationService
var DefaultAccountService service.AccountService

type ApplicationAccount interface {
	Create(user types.AccountSignUpRequest) (*entity.Account, error)
	Login(user types.AccountLoginRequest) (*entity.Account, error)
	Get(accountID uint) (*entity.Account, error)
	Update(accountID uint, user types.AccountUpdateRequest) error
	Delete(accountID uint) error
	SendVerificationCode(user types.AccountCodeRequest) (string, error)
	GetVerificationCode(email string) (string, error)
}

type applicationAccount struct {
	accountFc    entity.AccountFactoryConfig
	countFactory *entity.CodeFactory
	mysqlDB      *gorm.DB
	rdb          *redis.Client
}

func NewApplicationAccount(accountFc entity.AccountFactoryConfig, countFactory *entity.CodeFactory, mysqlDB *gorm.DB, rdb *redis.Client) ApplicationAccount {
	return &applicationAccount{accountFc: accountFc, countFactory: countFactory, mysqlDB: mysqlDB, rdb: rdb}
}

func (a *applicationAccount) SendVerificationCode(user types.AccountCodeRequest) (string, error) {
	// mysqlEventStore 事务
	tx := a.mysqlDB.Begin()
	eventStore, err := mysqlEventStore.NewMySQLEventStore(tx)
	if err != nil {
		return "", err
	}
	codeRepo, err := repo.NewCodeRepo(a.rdb, context.Background())
	if err != nil {
		return "", err
	}
	verificationService := service.NewVerificationService(codeRepo, a.countFactory, eventStore)
	code, err := verificationService.GenerateAuthCode(user.Email, 15*time.Minute)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	return code, tx.Commit().Error
}

// GetDefaultVerificationService 配置无事务的默认的verification service
func (a *applicationAccount) GetDefaultVerificationService() (service.VerificationService, error) {
	if DefaultVerificationService == nil {
		eventStore, err := mysqlEventStore.NewMySQLEventStore(a.mysqlDB)
		if err != nil {
			return nil, err
		}
		codeRepo, err := repo.NewCodeRepo(a.rdb, context.Background())
		if err != nil {
			return nil, err
		}
		DefaultVerificationService = service.NewVerificationService(codeRepo, a.countFactory, eventStore)
	}
	return DefaultVerificationService, nil
}

func (a *applicationAccount) GetDefaultAccountService() (service.AccountService, error) {
	if DefaultAccountService == nil {
		accountRepo, err := repo.NewAccountRepo(a.mysqlDB)
		if err != nil {
			return nil, err
		}
		DefaultAccountService = service.NewAccountService(accountRepo, a.accountFc)
	}
	return DefaultAccountService, nil
}

func (a *applicationAccount) GetVerificationCode(email string) (string, error) {
	verificationService, err := a.GetDefaultVerificationService()
	if err != nil {
		return "", err
	}
	return verificationService.GetAuthCode(email)
}

func (a *applicationAccount) Get(accountID uint) (*entity.Account, error) {
	accountService, err := a.GetDefaultAccountService()
	if err != nil {
		return nil, err
	}
	return accountService.GetAccountByID(accountID)
}

func (a *applicationAccount) Login(user types.AccountLoginRequest) (*entity.Account, error) {
	accountService, err := a.GetDefaultAccountService()
	if err != nil {
		return nil, err
	}
	// 查看邮箱对应账号是否存在
	account, err := accountService.GetAccountByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	// 查看账号密码是否匹配
	match, err := argon2id.ComparePasswordAndHash(user.Password, account.GetPassword())
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, slugerror.NewSlugError(slugerror.ErrUnauthorized, "invalid password", "password not match")
	}
	return account, nil
}

func (a *applicationAccount) Create(user types.AccountSignUpRequest) (*entity.Account, error) {
	accountService, err := a.GetDefaultAccountService()
	if err != nil {
		return nil, err
	}
	account, err := accountService.NewAccount(user.Email, user.Nickname, user.Password)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (a *applicationAccount) Update(accountID uint, user types.AccountUpdateRequest) error {
	accountService, err := a.GetDefaultAccountService()
	if err != nil {
		return err
	}
	if len(user.Email) > 0 {
		if err := accountService.ChangeEmail(accountID, user.Email); err != nil {
			return err
		}
	} else if len(user.Nickname) > 0 {
		if err := accountService.ChangeNickname(accountID, user.Nickname); err != nil {
			return err
		}
	} else if len(user.Password) > 0 {
		if err := accountService.ChangePassword(accountID, user.Password); err != nil {
			return err
		}
	} else {
		return slugerror.NewSlugError(slugerror.ErrInvalidInput,
			"empty request data", "The request needs to have a field that is not empty")
	}
	return nil
}

func (a *applicationAccount) Delete(accountID uint) error {
	accountService, err := a.GetDefaultAccountService()
	if err != nil {
		return err
	}
	return accountService.DeleteAccount(accountID)
}
