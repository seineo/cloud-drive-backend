package response

import (
	"CloudDrive/model"
	"gorm.io/gorm"
	"time"
)

type FileResponse struct {
	DirectoryHash string         `json:"directoryHash" binding:"required"`
	FileHash      string         `json:"fileHash" binding:"required"`
	Name          string         `json:"name" binding:"required"`
	Type          string         `json:"type"`
	Size          uint           `json:"size" binding:"required"`
	IsStarred     bool           `json:"isStarred" binding:"required"`
	CreatedAt     time.Time      `json:"createdAt" binding:"required"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt" binding:"required"`
}

func Convert2FileResponse(files []model.UserFileInfo, dirs []model.Directory) []FileResponse {
	fileResponses := []FileResponse{}
	for _, dir := range dirs {
		fileResponses = append(fileResponses, FileResponse{
			DirectoryHash: *dir.ParentHash,
			FileHash:      dir.Hash,
			Name:          dir.Name,
			Type:          "dir",
			Size:          0,
			IsStarred:     dir.IsStarred,
			CreatedAt:     dir.CreatedAt,
			DeletedAt:     dir.DeletedAt,
		})
	}
	for _, file := range files {
		fileResponses = append(fileResponses, FileResponse{
			DirectoryHash: file.DirectoryHash,
			FileHash:      file.FileHash,
			Name:          file.Name,
			Type:          file.Type,
			Size:          file.Size,
			IsStarred:     file.IsStarred,
			CreatedAt:     file.CreatedAt,
			DeletedAt:     file.DeletedAt,
		})
	}
	return fileResponses
}
