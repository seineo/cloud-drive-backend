package request

type DirectoryRequest struct {
	Hash    string `json:"hash" binding:"required"`
	Name    string `json:"name" binding:"required"`
	DirPath string `json:"dirPath" binding:"required"`
}

type FileRequest struct {
	FileHash string `json:"fileHash" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
	FileType string `json:"fileType" binding:"required"`
	DirHash  string `json:"dirHash" binding:"required"`
	FileSize uint   `json:"fileSize" binding:"required"`
	Exists   bool   `json:"exists" binding:"required"`
}
