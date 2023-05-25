package response

import "time"

type FileResponse struct {
	Hash     string    `json:"hash" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	Type     string    `json:"type" binding:"required"`
	Size     uint      `json:"size" binding:"required"`
	CreateAt time.Time `json:"createAt" binding:"required"`
}

type DirResponse struct {
	Hash     string    `json:"hash" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	CreateAt time.Time `json:"createAt" binding:"required"`
}
