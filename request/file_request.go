package request

type DirectoryRequest struct {
	Hash    string `json:"hash" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Path    string `json:"path" binding:"required"`
	DirHash string `json:"dirHash" binding:"required"`
}

type FileRequest struct {
	FileHash string `json:"fileHash" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
	FileType string `json:"fileType" binding:"required"`
	DirHash  string `json:"dirHash" binding:"required"`
	FileSize uint   `json:"fileSize" binding:"required"`
}
