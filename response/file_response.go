package response

import "time"

type FileResponse struct {
	Hash      string    `json:"hash" binding:"required"`
	Name      string    `json:"name" binding:"required"`
	Type      string    `json:"type"`
	Size      uint      `json:"size" binding:"required"`
	CreatedAt time.Time `json:"createdAt" binding:"required"`
}
