package model

import (
	"CloudDrive/request"
	"gorm.io/gorm"
	"time"
)

type File struct {
	Hash     string `gorm:"primaryKey"` // hash value of file content
	FileType string // dir, pdf, img, video...
	Size     uint
	Location string // real file storage path
}

type Directory struct {
	Hash    string `gorm:"primaryKey"`
	UserID  uint
	Name    string
	DirPath string // virtual directory path shown for users
	Files   []File `gorm:"many2many:directory_files;"`
}

type DirectoryFile struct {
	DirectoryHash string `gorm:"primaryKey"`
	FileHash      string `gorm:"primaryKey"`
	FileName      string
	CreateAt      time.Time
	DeleteAt      gorm.DeletedAt `gorm:"index"`
}

func (df *DirectoryFile) BeforeCreate(db *gorm.DB) error {
	df.CreateAt = time.Now()
	return nil
}

func (df *DirectoryFile) BeforeDelete(db *gorm.DB) error {
	df.DeleteAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	return nil
}

func StoreDirMetadata(dir *Directory) error {
	return db.Create(dir).Error
}

// StoreFileMetadata inserts file metadata into two tables: `file`s and `directory_files` using transaction.
// Ref link: https://stackoverflow.com/questions/65012939/custom-fields-in-many2many-jointable
func StoreFileMetadata(fileRequest *request.FileRequest, fileStoragePath string) error {
	file := &File{
		Hash:     fileRequest.FileHash,
		FileType: fileRequest.FileType,
		Size:     fileRequest.FileSize,
		Location: fileStoragePath,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// insert file data into file table
		err := db.Create(file).Error
		if err != nil {
			return err
		}
		// insert link between directory and file into association table
		err = db.Create(&DirectoryFile{
			DirectoryHash: fileRequest.DirHash,
			FileHash:      fileRequest.FileHash,
			FileName:      fileRequest.FileName,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

//// GetFilesMetadata when directory is found, return the file list
//// note that directory hash is generated using uuid, so we treat it unique
//func GetFilesMetadata(dirHash string) ([]File, error) {
//	var files []File
//	var file File
//	// find directory first
//	if err := db.Where("hash = ?", dirHash).First(&file).Error; err != nil {
//		return nil, err
//	}
//	var dirPath string
//	if file.DirPath != "" {
//		dirPath = file.DirPath + "/" + file.Name
//	} else {
//		dirPath = file.Name
//	}
//	// find files under that directory
//	if err := db.Where("dir_path = ?", dirPath).Find(&files).Error; err != nil {
//		return nil, err
//	}
//	return files, nil
//}

//
//// GetFileMetadata when file is found, return the pointer to the file,
//// note that we  care about the path of the file in different user cloud drives.
//func GetFileMetadata(userID uint, dirPath string) (*File, error) {
//	var file File // it will initialize with default fields!
//	if err := db.Where("hash = ?", fileHash).First(&file).Error; err != nil {
//		return nil, err
//	}
//	return &file, nil
//}
//
//// GetFileLocation return the storage path for given file
//func GetFileLocation(userID uint, dirPath string, fileName string) (string, error) {
//	var file File
//	if err := db.Where("user_id = ? and dir_path = ? and name = ?", userID, dirPath, fileName).First(&file).Error; err != nil {
//		return "", err
//	}
//	return file.Location, nil
//}
//
//// FileExists given hash name of the file, check whether file exists
//func FileExists(hash string) (bool, error) {
//	var file File
//	result := db.Where("hash = ?", hash).First(&file)
//	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
//		return false, nil
//	} else {
//		return result.RowsAffected == 1, result.Error
//	}
//}
//
//// DeleteFilesMetadata given directory path, delete the metadata of files under it
//func DeleteFilesMetadata(userID uint, dirPath string) error {
//	err := db.Where("user_id = ? and dir_path = ?", userID, dirPath).Delete(&File{}).Error
//	return err
//}
