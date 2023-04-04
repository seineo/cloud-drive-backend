package model

import "time"

type File struct {
	Hash        string `gorm:"primaryKey"` // MD5 hash value of file content, as primary key
	Name        string
	UserID      uint
	ContentType string // dir, pdf, img, video...
	Size        string
	DirPath     string // virtual directory path shown for users
	Location    string // real file storage path
	CreateTime  time.Time
}

func StoreFileMetadata(file *File) error {
	return db.Create(file).Error
}

func GetFilesMetadata(dirPath string) ([]File, error) {
	var files []File
	err := db.Where("dir_path = ？", dirPath).Find(&files).Error
	return files, err
}

func GetFileMetadata(dirPath string, fileName string) (*File, error) {
	var file File
	err := db.Where("dir_path = ？and name = ?", dirPath, fileName).Find(&file).Error
	return &file, err
}

func GetFileLocation(dirPath string, fileName string) (string, error) {
	var file File
	err := db.Where("dir_path = ? and name = ?", dirPath, fileName).First(&file).Error
	return file.Location, err
}

func FileExists(hash string) (bool, error) {
	var file File
	result := db.Where("hash = ?", hash).First(&file)
	return result.RowsAffected == 1, result.Error
}
