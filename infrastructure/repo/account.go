package repo

import (
	"CloudDrive/common/slugerror"
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

var RecordNotFoundError = slugerror.NewSlugError(slugerror.ErrUnprocessable,
	"request unprocessable", "record not found in the database")

type account struct {
	gorm.Model
	Email    string `gorm:"unique; not null"`
	Nickname string `gorm:"not null"`
	Password string `gorm:"not null"`
}

type accountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) repository.AccountRepo {
	if db == nil {
		panic("missing db")
	}
	return &accountRepo{db: db}
}

// 领域的账号可能最初并无ID，数据库分配后才有，所以这里不获取他的ID
func fromDomainAccount(ea entity.Account) *account {
	return &account{
		Email:    ea.GetEmail(),
		Nickname: ea.GetNickname(),
		Password: ea.GetPassword(),
	}
}

func toDomainAccount(mysqlAccount account) *entity.Account {
	return entity.UnmarshallAccount(mysqlAccount.ID, mysqlAccount.Email, mysqlAccount.Nickname, mysqlAccount.Password)
}

func (repo *accountRepo) Get(accountID uint) (*entity.Account, error) {
	mysqlAccount := account{}
	if err := repo.db.First(&mysqlAccount, "id = ?", accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, RecordNotFoundError
		}
		return nil, fmt.Errorf("failed to get account by id: %w", err)
	}
	return toDomainAccount(mysqlAccount), nil
}

func (repo *accountRepo) GetByEmail(email string) (*entity.Account, error) {
	mysqlAccount := account{}
	if err := repo.db.Find(&mysqlAccount, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}
	return toDomainAccount(mysqlAccount), nil
}

func (repo *accountRepo) Create(ea entity.Account) (*entity.Account, error) {
	mysqlAccount := fromDomainAccount(ea)
	if err := repo.db.Create(mysqlAccount).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return toDomainAccount(*mysqlAccount), nil
}

func (repo *accountRepo) Update(ea entity.Account) (*entity.Account, error) {
	mysqlAccount := fromDomainAccount(ea)
	// 请注意Updates使用struct更新时默认只更新非零字段
	result := repo.db.Model(&account{}).Where("id = ?", ea.GetID()).Updates(*mysqlAccount)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, RecordNotFoundError
	}
	return toDomainAccount(*mysqlAccount), nil
}

func (repo *accountRepo) Delete(accountID uint) error {
	result := repo.db.Delete(&account{}, accountID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return RecordNotFoundError
	}
	return nil
}
