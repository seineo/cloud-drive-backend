package service

import (
	"resource/domain/entity"
	"resource/domain/repository"
)

type FolderService interface {
	CreateFolder(accountID uint, policyID uint, parentFolder *uint, name string) (*entity.Folder, error)
	GetSubFolders(folderID uint) ([]*entity.Folder, error)
}

type folderService struct {
	folderRepo repository.FolderRepo
}

func (f *folderService) CreateFolder(accountID uint, policyID uint, parentFolder *uint, name string) (*entity.Folder, error) {
	account := entity.NewFolder(accountID, policyID, parentFolder, name)
	folder, err := f.folderRepo.CreateFolder(*account)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func (f *folderService) GetSubFolders(folderID uint) ([]*entity.Folder, error) {
	folders, err := f.folderRepo.GetSubFolders(folderID)
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func NewFolderService(folderRepo repository.FolderRepo) FolderService {
	return &folderService{folderRepo: folderRepo}
}
