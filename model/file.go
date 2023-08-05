package model

import (
	config2 "CloudDrive/config"
	"CloudDrive/request"
	"errors"
	"gorm.io/gorm"
	"time"
)

type File struct {
	Hash      string `gorm:"primaryKey;size:255"` // hash value of file content
	FileType  string // dir, pdf, img, video...
	Size      uint
	Location  string // real file storage path
	RefCount  uint
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Directory struct {
	Hash       string `gorm:"primaryKey;size:255"`
	UserID     uint
	Name       string
	CreatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	ParentHash *string        `gorm:"size:255"`
	SubDirs    []Directory    `gorm:"foreignKey:ParentHash;"`                                 // cascade deletion will only delete one-level subdirectories
	Files      []File         `gorm:"many2many:directory_files;constraint:OnDelete:CASCADE;"` // delete all user files after deleting real files
}

type DirectoryFile struct {
	DirectoryHash string
	FileHash      string
	FileName      string
	CreatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type FileInfo struct {
	Hash      string
	Name      string
	Type      string
	Size      uint
	Location  string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

var RefCountError = errors.New("reference count of the file is already 0")

// StoreDirMetadata stores directory metadata and adds association with its parent directory.
func StoreDirMetadata(directoryRequest *request.DirectoryRequest, userID uint) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		// insert file data into file table
		dir := Directory{
			Hash:       directoryRequest.Hash,
			UserID:     userID,
			Name:       directoryRequest.Name,
			ParentHash: &directoryRequest.DirHash,
		}
		if err := tx.Create(&dir).Error; err != nil {
			return err
		}
		// find parent directory
		var parentDir Directory
		if err := db.Where("hash = ?", directoryRequest.DirHash).First(&parentDir).Error; err != nil {
			return err
		}
		// update association
		if err := tx.Model(&parentDir).Association("SubDirs").Append(&dir); err != nil {
			return err
		}
		return nil
	})
	return err

}

// StoreFileMetadata stores file metadata to database.
// If file exists, insert file metadata into table `directory_files` and update reference count in table `files`,
// otherwise insert into two tables: `file`s and `directory_files` using transaction.
// Ref link: https://stackoverflow.com/questions/65012939/custom-fields-in-many2many-jointable
func StoreFileMetadata(fileRequest *request.FileRequest, fileStoragePath string, exists bool) error {
	directoryFile := DirectoryFile{
		DirectoryHash: fileRequest.DirHash,
		FileHash:      fileRequest.FileHash,
		FileName:      fileRequest.FileName,
	}
	if exists {
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&directoryFile).Error; err != nil {
				return err
			}
			// update reference count
			if err := tx.Model(&File{}).Where("hash = ?", fileRequest.FileHash).
				Update("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
				return err
			}
			return nil
		})
		return err
	} else {
		file := &File{
			Hash:     fileRequest.FileHash,
			FileType: fileRequest.FileType,
			Size:     fileRequest.FileSize,
			Location: fileStoragePath,
			RefCount: 1,
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			// insert file data into file table
			if err := tx.Create(file).Error; err != nil {
				return err
			}
			// insert link between directory and file into association table
			err := tx.Create(&directoryFile).Error
			if err != nil {
				return err
			}
			return nil
		})
		return err
	}
	return nil
}

// GetFilesMetadata returns the list of file metadata and subdirectory metadata.
// Note that directory hash is generated using uuid, so we treat it unique.
func GetFilesMetadata(dirHash string) ([]FileInfo, []Directory, error) {
	var filesInfo []FileInfo
	var dirs []Directory
	var parentDir Directory

	// find subdirectories
	if err := db.Where("hash = ?", dirHash).First(&parentDir).Error; err != nil {
		return nil, nil, err
	}
	if err := db.Model(&parentDir).Association("SubDirs").Find(&dirs); err != nil {
		return nil, nil, err
	}

	//left join tables `directory_files` and `files` to get files info
	subQuery := db.Select("directory_hash, file_hash, file_name, created_at, deleted_at").Table("directory_files").
		Where("directory_hash = ?", dirHash)
	db.Debug().
		Select("file_hash as hash, file_name as name, file_type as type, size, location, query.created_at").
		Table("(?) as query", subQuery).
		Joins("left join files on query.file_hash = files.hash").Find(&filesInfo)

	return filesInfo, dirs, nil
}

//// note: should run in a transaction
//func decreaseRefCount(tx *gorm.DB, fileHash string) error {
//	// find the file in table `files`
//	var file File
//	if err := tx.Where("hash = ?", fileHash).First(&file).Error; err != nil {
//		return err
//	}
//	// check reference count
//	if file.RefCount <= 0 {
//		return RefCountError
//	}
//	// decrease reference count in table `files`
//	if err := tx.Model(&File{}).Where("hash = ?", fileHash).
//		Update("ref_count", gorm.Expr("ref_count - ?", 1)).Error; err != nil {
//		return err
//	}
//	// if reference count equals 0, soft delete
//	if file.RefCount == 0 {
//		if err := tx.Delete(&file).Error; err != nil {
//			return err
//		}
//	}
//	return nil
//}

// DeleteFile deletes given file and reduces reference count of real file after deletion.
func DeleteFile(dirHash string, fileHash string) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		// soft delete in table `directory_files`, note in relation table we need where condition statement
		result := tx.Where("directory_hash = ? and file_hash = ?", dirHash, fileHash).Delete(&DirectoryFile{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 { // if it doesn't delete any row, then it should not decrease reference count
			return nil
		}
		// decrease reference count in table `files`
		if err := tx.Model(&File{}).Where("hash = ?", fileHash).
			Update("ref_count", gorm.Expr("ref_count - ?", 1)).Error; err != nil {
			return err
		}
		// check the reference count in table `files`
		var file File
		if err := tx.Where("hash = ?", fileHash).First(&file).Error; err != nil {
			return err
		}
		// check reference count, and it will roll back if error occurs
		if file.RefCount < 0 {
			return RefCountError
		}
		// if reference count equals 0, soft delete
		if file.RefCount == 0 {
			if err := tx.Delete(&file).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// DeleteStaleFiles regularly checks and deletes stale files.
func DeleteStaleFiles() {
	config := config2.GetConfig()
	//configs.Storage.DiskStoragePath
	staleTime := time.Now().Add(-config.FileStaleTime)
	var files []File
	if err := db.Unscoped().Where("deleted_at <= ?", staleTime).Find(&files).Error; err != nil {
		log.Errorf("failed to find stale files, error: %v\n", err.Error())
	}
	if len(files) != 0 {
		db.Unscoped().Delete(&files)
	}
}

// GetAllUnderDir returns all unique files and subdirectories under given directory.
func GetAllUnderDir(dirHash string) ([]Directory, []DirectoryFile, error) {
	var files []DirectoryFile
	var dirs []Directory
	var parentDir Directory
	if err := db.Where("hash = ?", dirHash).First(&parentDir).Error; err != nil {
		return nil, nil, err
	}
	if err := db.Model(&parentDir).Association("SubDirs").Find(&dirs); err != nil {
		return nil, nil, err
	}
	if err := db.Where("directory_hash = ?", dirHash).Find(&files).Error; err != nil {
		return nil, nil, err
	}
	for _, dir := range dirs {
		curDirs, curFiles, err := GetAllUnderDir(dir.Hash)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, curFiles...)
		dirs = append(dirs, curDirs...)
	}
	return dirs, files, nil
}

// DeleteDir deletes given directory and the subdirectory and files under it.
func DeleteDir(dirHash string) error {
	dirs, files, err := GetAllUnderDir(dirHash)
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Directory{Hash: dirHash}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&dirs).Error; err != nil {
			return err
		}
		for _, file := range files {
			if err := DeleteFile(file.DirectoryHash, file.FileHash); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func GetDirMetadata(hash string) (*Directory, error) {
	var dir Directory
	err := db.Where("hash = ?", hash).First(&dir).Error
	if err != nil {
		return nil, err
	}
	return &dir, nil
}

// GetFileMetadata gets file metadata from database.
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
