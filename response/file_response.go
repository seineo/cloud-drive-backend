package response

import "time"

type FileResponse struct {
	DirectoryHash string    `json:"directoryHash" binding:"required"`
	FileHash      string    `json:"fileHash" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	Type          string    `json:"type"`
	Size          uint      `json:"size" binding:"required"`
	IsStarred     bool      `json:"isStarred" binding:"required"`
	CreatedAt     time.Time `json:"createdAt" binding:"required"`
}
