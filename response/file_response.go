package response

import (
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

type DirTraceResponse struct {
	Name string `json:"name" binding:"required"`
	Hash string `json:"hash" binding:"required"`
}
