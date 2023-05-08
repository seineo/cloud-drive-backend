package model

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type File struct {
	Hash       string `gorm:"primaryKey"` // MD5 hash value of file content, as primary key
	Name       string
	UserID     uint
	FileType   string // dir, pdf, img, video...
	Size       uint
	DirPath    string // virtual directory path shown for users
	Location   string // real file storage path
	CreateTime time.Time
}

func StoreFileMetadata(file *File) error {
	return db.Create(file).Error
}

// GetFilesMetadata when file is found, return the file list
// when not found, return empty list without error
func GetFilesMetadata(userID uint, fileHash string) ([]File, error) {
	var files []File
	var file File
	// find directory first
	if err := db.Where("hash = ?", fileHash).First(&file).Error; err != nil {
		return nil, err
	}
	dirPath := file.DirPath
	// find files under that directory
	if err := db.Where("user_id = ? and dir_path = ?", userID, dirPath).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileMetadata when file is found, return the pointer to the file,
// when not found, raise RecordNotFound error, it should be dealt with differently from other errors
func GetFileMetadata(userID uint, dirPath string, fileName string) (*File, error) {
	var file File // it will initialize with default fields!
	if err := db.Where("user_id = ? and dir_path = ? and name = ?", userID, dirPath, fileName).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFileLocation return the storage path for given file
func GetFileLocation(userID uint, dirPath string, fileName string) (string, error) {
	var file File
	if err := db.Where("user_id = ? and dir_path = ? and name = ?", userID, dirPath, fileName).First(&file).Error; err != nil {
		return "", err
	}
	return file.Location, nil
}

// FileExists given hash name of the file, check whether file exists
func FileExists(hash string) (bool, error) {
	var file File
	result := db.Where("hash = ?", hash).First(&file)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	} else {
		return result.RowsAffected == 1, result.Error
	}
}

// DeleteFilesMetadata given directory path, delete the metadata of files under it
func DeleteFilesMetadata(userID uint, dirPath string) error {
	err := db.Where("user_id = ? and dir_path = ?", userID, dirPath).Delete(&File{}).Error
	return err
}
