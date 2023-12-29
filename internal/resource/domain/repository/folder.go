package repository

import "resource/domain/entity"

type FolderRepo interface {
	CreateFolder(folder entity.Folder) (*entity.Folder, error)
	GetSubFolders(folderID uint) ([]*entity.Folder, error)
}
