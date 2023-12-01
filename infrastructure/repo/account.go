package repo

import (
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"fmt"
	"gorm.io/gorm"
)

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

func fromDomainAccount(acc entity.Account) account {
	return account{
		Email:    acc.Email,
		Nickname: acc.Nickname,
		Password: acc.Password,
	}
}

func toDomainAccount(mysqlAccount account) *entity.Account {
	return &entity.Account{
		ID:       mysqlAccount.ID,
		Email:    mysqlAccount.Email,
		Nickname: mysqlAccount.Nickname,
		Password: mysqlAccount.Password,
	}
}

func (repo *accountRepo) Get(accountID uint) (*entity.Account, error) {
	mysqlAccount := account{}
	if err := repo.db.First(&mysqlAccount, "id = ?", accountID).Error; err != nil {
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

func (repo *accountRepo) Create(acc entity.Account) (*entity.Account, error) {
	mysqlAccount := fromDomainAccount(acc)
	if err := repo.db.Create(&mysqlAccount).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return toDomainAccount(mysqlAccount), nil
}

func (repo *accountRepo) Update(acc entity.Account) (*entity.Account, error) {
	mysqlAccount := fromDomainAccount(acc)
	// 请注意Updates使用struct更新时默认只更新非零字段
	if err := repo.db.Model(&account{}).Where("id = ?", acc.ID).Updates(mysqlAccount).Error; err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}
	return toDomainAccount(mysqlAccount), nil
}

func (repo *accountRepo) Delete(accountID uint) error {
	if err := repo.db.Delete(&account{}, accountID).Error; err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	return nil
}
