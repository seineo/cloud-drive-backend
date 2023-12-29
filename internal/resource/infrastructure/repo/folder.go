package repo

import (
	"fmt"
	"gorm.io/gorm"
	"resource/domain/entity"
	"resource/domain/repository"
)

// TODO accountID policyID外键
type Folder struct {
	gorm.Model
	AccountID    uint
	PolicyID     uint
	ParentFolder *uint
	Name         string
	SubFolders   []Folder `gorm:"foreignkey:ParentFolder"`
}

type folderRepo struct {
	db *gorm.DB
}

func fromDomainFolder(folder entity.Folder) *Folder {
	return &Folder{
		AccountID:    folder.AccountID(),
		PolicyID:     folder.PolicyID(),
		ParentFolder: folder.ParentFolder(),
		Name:         folder.Name(),
	}
}

func toDomainFolder(folder Folder) *entity.Folder {
	return entity.UnmarshallFolder(folder.ID, folder.AccountID, folder.PolicyID, folder.ParentFolder, folder.Name)
}

func toDomainFolders(folders []Folder) []*entity.Folder {
	result := []*entity.Folder{}
	for _, folder := range folders {
		result = append(result, entity.UnmarshallFolder(folder.ID, folder.AccountID,
			folder.PolicyID, folder.ParentFolder, folder.Name))
	}
	return result
}

func (f *folderRepo) CreateFolder(folder entity.Folder) (*entity.Folder, error) {
	dbFolder := fromDomainFolder(folder)
	if err := f.db.Create(&dbFolder).Error; err != nil {
		return nil, err
	}
	return toDomainFolder(*dbFolder), nil
}

func (f *folderRepo) GetSubFolders(folderID uint) ([]*entity.Folder, error) {
	dbFolders := []Folder{}
	if err := f.db.Model(&Folder{}).Where("parent_folder = ?", folderID).Find(&dbFolders).Error; err != nil {
		return nil, err
	}
	return toDomainFolders(dbFolders), nil
}

func NewFolderRepo(db *gorm.DB) (repository.FolderRepo, error) {
	if db == nil {
		panic("missing db")
	}
	err := db.AutoMigrate(&Folder{})
	if err != nil {
		return nil, fmt.Errorf("unable to migrate folder model: %w", err)
	}
	return &folderRepo{db: db}, nil
}
