package model

import (
	"CloudDrive/request"
	"CloudDrive/response"
	"errors"
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
	Hash       string `gorm:"primaryKey"`
	UserID     uint
	Name       string
	Path       string
	CreateAt   time.Time
	DeleteAt   gorm.DeletedAt `gorm:"index"`
	ParentHash *string        `gorm:"size:191"` // virtual directory path shown for users
	SubDirs    []Directory    `gorm:"foreignKey:ParentHash"`
}

type DirectoryFile struct {
	DirectoryHash string `gorm:"primaryKey"`
	FileHash      string `gorm:"primaryKey"`
	FileName      string
	CreateAt      time.Time
	DeleteAt      gorm.DeletedAt `gorm:"index"`
}

// StoreDirMetadata stores directory metadata and adds association with its parent directory.
func StoreDirMetadata(dir *Directory) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		// insert file data into file table
		if err := tx.Create(dir).Error; err != nil {
			return err
		}
		// find parent directory
		var parentDir Directory
		if err := db.Where("hash = ?", dir.ParentHash).First(&parentDir).Error; err != nil {
			return err
		}
		// update association
		if err := tx.Model(&parentDir).Association("SubDirs").Append(dir); err != nil {
			return err
		}
		return nil
	})
	return err

}

// StoreFileMetadata stores file metadata to database.
// If file exists, insert file metadata into table `directory_files` only,
// otherwise insert into two tables: `file`s and `directory_files` using transaction.
// Ref link: https://stackoverflow.com/questions/65012939/custom-fields-in-many2many-jointable
func StoreFileMetadata(fileRequest *request.FileRequest, fileStoragePath string, exists bool) error {
	if exists {
		err := db.Create(&DirectoryFile{
			DirectoryHash: fileRequest.DirHash,
			FileHash:      fileRequest.FileHash,
			FileName:      fileRequest.FileName,
			CreateAt:      time.Now(),
		}).Error
		if err != nil {
			return err
		}
	} else {
		file := &File{
			Hash:     fileRequest.FileHash,
			FileType: fileRequest.FileType,
			Size:     fileRequest.FileSize,
			Location: fileStoragePath,
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			// insert file data into file table
			if err := tx.Create(file).Error; err != nil {
				return err
			}
			// insert link between directory and file into association table
			err := tx.Create(&DirectoryFile{
				DirectoryHash: fileRequest.DirHash,
				FileHash:      fileRequest.FileHash,
				FileName:      fileRequest.FileName,
				CreateAt:      time.Now(),
			}).Error
			if err != nil {
				return err
			}
			return nil
		})
		return err
	}
	return nil
}

// GetFilesMetadata returns the list of file metadata and subdirectory metadata when directory is found.
// Note that directory hash is generated using uuid, so we treat it unique.
func GetFilesMetadata(dirHash string) ([]response.FileResponse, []response.DirResponse, error) {
	var files []response.FileResponse
	var dirs []Directory
	var parentDir Directory

	// find subdirectories
	if err := db.Where("hash = ?", dirHash).First(&parentDir).Error; err != nil {
		return nil, nil, err
	}
	if err := db.Model(&parentDir).Association("SubDirs").Find(&dirs); err != nil {
		return nil, nil, err
	}

	// construct directories in response
	var dirRepsonse []response.DirResponse
	for _, dir := range dirs {
		dirRepsonse = append(dirRepsonse, response.DirResponse{
			Hash:     dir.Hash,
			Name:     dir.Name,
			CreateAt: dir.CreateAt,
		})
	}

	// left join tables `directory_files` and `files` to get files
	subQuery := db.Select("directory_hash, file_hash, file_name, create_at").Table("directory_files").
		Where("directory_hash = ?", dirHash)
	db.Debug().
		Select("file_hash as hash, file_name as name, file_type as type, size, create_at").
		Table("(?) as query", subQuery).
		Joins("left join files on query.file_hash = files.hash").Find(&files)

	return files, dirRepsonse, nil
}

// GetFileMetadata gets file metadata from database
func GetFileMetadata(hash string) (*File, error) {
	var file File
	err := db.Where("hash = ?", hash).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FileExists checks whether file exists given hash name of the file.
func FileExists(hash string) (bool, error) {
	var file File
	result := db.Where("hash = ?", hash).First(&file)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	} else {
		return result.RowsAffected == 1, result.Error
	}
}

//// DeleteFilesMetadata given directory path, delete the metadata of files under it
//func DeleteFilesMetadata(userID uint, dirPath string) error {
//	err := db.Where("user_id = ? and dir_path = ?", userID, dirPath).Delete(&File{}).Error
//	return err
//}
