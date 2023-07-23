package request

type DirectoryRequest struct {
	Hash    string `json:"hash" binding:"required"`
	Name    string `json:"name" binding:"required"`
	DirHash string `json:"dirHash" binding:"required"`
}

type FileRequest struct {
	FileHash string `json:"fileHash" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
	FileType string `json:"fileType"`
	DirHash  string `json:"dirHash" binding:"required"`
	FileSize uint   `json:"fileSize" binding:"required"`
}

type ChunkRequest struct {
	FileHash    string `json:"fileHash" binding:"required"`
	ChunkHash   string `json:"chunkHash" binding:"required"`
	Index       uint   `json:"index" binding:"required"` // start from 1 to avoid binding required error when index is 0
	TotalChunks uint   `json:"totalChunks" binding:"required"`
}

type CurrentChunks struct {
	TotalChunks uint            `json:"totalChunks"`
	Indexes     map[uint]string `json:"indexes"`
}
