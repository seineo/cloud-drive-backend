package model

import (
	"CloudDrive/request"
	"errors"
	"fmt"
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
	IsStarred  bool
	CreatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	ParentHash *string        `gorm:"size:255"`
	SubDirs    []Directory    `gorm:"foreignKey:ParentHash;"`                                 // cascade deletion will only delete one-level subdirectories
	Files      []File         `gorm:"many2many:directory_files;constraint:OnDelete:CASCADE;"` // delete related user files after deleting real files or directories
}

type DirectoryFile struct {
	DirectoryHash string
	FileHash      string
	FileName      string
	UserID        uint
	IsStarred     bool // false -> don't filter, true -> filter
	CreatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type UserFileInfo struct {
	DirectoryHash string
	FileHash      string
	Name          string
	Type          string
	Size          uint
	IsStarred     bool
	Location      string
	CreatedAt     time.Time
	DeletedAt     gorm.DeletedAt
}

var RefCountError = errors.New("reference count of the file is already 0")
var ColumnNotExistsError = errors.New("")

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
func StoreFileMetadata(fileRequest *request.FileRequest, fileStoragePath string, exists bool, userID uint) error {
	directoryFile := DirectoryFile{
		DirectoryHash: fileRequest.DirHash,
		FileHash:      fileRequest.FileHash,
		FileName:      fileRequest.FileName,
		UserID:        userID,
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
func GetFilesMetadata(dirHash string, isStarred bool, sort string, order string) ([]UserFileInfo, []Directory, error) {
	var filesInfo []UserFileInfo
	var dirs []Directory
	var parentDir Directory

	// find subdirectories
	if err := db.Where("hash = ?", dirHash).First(&parentDir).Error; err != nil {
		return nil, nil, err
	}
	dirQuery := db.Model(&parentDir).Session(&gorm.Session{})
	if isStarred {
		dirQuery = dirQuery.Where("is_starred = ?", isStarred)
	}
	if len(sort) != 0 && len(order) != 0 {
		dirQuery = dirQuery.Order(fmt.Sprintf("%s %s", sort, order))
	}
	if err := dirQuery.Association("SubDirs").Find(&dirs); err != nil {
		return nil, nil, err
	}

	//left join tables `directory_files` and `files` to get files info
	fileSubQuery := db.Select("*").Table("directory_files").
		Where("directory_hash = ?", dirHash).Session(&gorm.Session{})
	if isStarred {
		fileSubQuery = fileSubQuery.Where("is_starred = ?", isStarred)
	}
	fileQuery := db.Debug().
		Select("directory_hash, file_hash, file_name as name, file_type as type, size, location, is_starred, query.created_at, query.deleted_at").
		Table("(?) as query", fileSubQuery).
		Joins("left join files on query.file_hash = files.hash").Session(&gorm.Session{})
	if len(sort) != 0 && len(order) != 0 {
		fileQuery = fileQuery.Order(fmt.Sprintf("%s %s", sort, order))
	}
	if err := fileQuery.Find(&filesInfo).Error; err != nil {
		return nil, nil, err
	}
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
	//configs.Storage.DiskStoragePath
	staleTime := time.Now().Add(-configs.File.StaleTime)
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

func GetAllUnderTrashDir(dirHash string) ([]Directory, []DirectoryFile, error) {
	var files []DirectoryFile
	var dirs []Directory
	var parentDir Directory
	if err := db.Unscoped().Where("hash = ?", dirHash).First(&parentDir).Error; err != nil {
		return nil, nil, err
	}
	if err := db.Unscoped().Model(&parentDir).Association("SubDirs").Find(&dirs); err != nil {
		return nil, nil, err
	}
	if err := db.Unscoped().Where("directory_hash = ?", dirHash).Find(&files).Error; err != nil {
		return nil, nil, err
	}
	for _, dir := range dirs {
		curDirs, curFiles, err := GetAllUnderTrashDir(dir.Hash)
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
		for _, file := range files {
			if err := DeleteFile(file.DirectoryHash, file.FileHash); err != nil {
				return err
			}
		}
		if len(dirs) != 0 {
			if err := tx.Debug().Delete(&dirs).Error; err != nil {
				return err
			}
		}
		if err := tx.Debug().Delete(&Directory{Hash: dirHash}).Error; err != nil {
			return err
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

func StarDir(dirHash string) error {
	return db.Model(&Directory{Hash: dirHash}).Update("is_starred", true).Error
}

func UnstarDir(dirHash string) error {
	return db.Model(&Directory{Hash: dirHash}).Update("is_starred", false).Error
}

func StarFile(dirHash string, fileHash string) error {
	return db.Model(&DirectoryFile{}).Where("directory_hash = ? and file_hash = ?", dirHash, fileHash).
		Update("is_starred", true).Error
}

func UnstarFile(dirHash string, fileHash string) error {
	return db.Model(&DirectoryFile{}).Where("directory_hash = ? and file_hash = ?", dirHash, fileHash).
		Update("is_starred", false).Error
}

func GetStarredFiles(userID uint) ([]UserFileInfo, []Directory, error) {
	var files []UserFileInfo
	var dirs []Directory
	if err := db.Where("user_id = ? and is_starred = 1", userID).Find(&dirs).Error; err != nil {
		return nil, nil, err
	}
	fileSubQuery := db.Select("*").Table("directory_files").Where("user_id = ? and is_starred = 1", userID)
	fileQuery := db.Debug().
		Select("directory_hash, file_hash, file_name as name, file_type as type, size, location, is_starred, query.created_at, query.deleted_at").
		Table("(?) as query", fileSubQuery).
		Joins("left join files on query.file_hash = files.hash")
	if err := fileQuery.Find(&files).Error; err != nil {
		return nil, nil, err
	}
	return files, dirs, nil
}

func TraceTrashAncestorDir(dirHash string) (string, error) {
	var curDir Directory
	var parentDir Directory
	if err := db.Unscoped().Where("hash = ?", dirHash).First(&curDir).Error; err != nil {
		return "", err
	}
	// search parent directory until it has not been deleted
	for true {
		if curDir.ParentHash == nil { // root directory
			break
		} else {
			if err := db.Unscoped().Where("hash = ? and deleted_at is not null", *curDir.ParentHash).
				First(&parentDir).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					break
				} else {
					return "", err
				}
			} else {
				curDir = parentDir
				parentDir = Directory{}
			}
		}
	}
	return curDir.Hash, nil
}

func GetTrashFiles(userID uint) ([]UserFileInfo, []Directory, error) {
	var filesInfo []UserFileInfo
	var dirs []Directory
	dirSubQuery := db.Unscoped().Select("*").Table("directories").
		Where(" user_id = ? and deleted_at is not null", userID)
	dirQuery := db.Unscoped().Select("query.hash, query.user_id,  query.name,"+
		" query.is_starred, query.created_at, query.deleted_at, query.parent_hash").
		Where("directories.deleted_at is null").
		Table("(?) as query", dirSubQuery).
		Joins("left join directories on query.parent_hash = directories.hash")
	if err := dirQuery.Find(&dirs).Error; err != nil {
		return nil, nil, err
	}
	fileSubQuery1 := db.Unscoped().Select("*").
		Where("user_id = ? and deleted_at is not null", userID).
		Table("directory_files")
	fileSubQuery2 := db.Unscoped().Select("query1.directory_hash, query1.file_hash, query1.file_name,"+
		" query1.user_id, query1.is_starred, query1.created_at, query1.deleted_at").
		Where("directories.deleted_at is null").
		Table("(?) as query1", fileSubQuery1).
		Joins("left join directories on query1.directory_hash = directories.hash")
	fileQuery := db.Unscoped().
		Select("directory_hash, file_hash, file_name as name, file_type as type, size, location, is_starred, query2.created_at, query2.deleted_at").
		Table("(?) as query2", fileSubQuery2).
		Joins("left join files on query2.file_hash = files.hash")
	if err := fileQuery.Find(&filesInfo).Error; err != nil {
		return nil, nil, err
	}
	return filesInfo, dirs, nil
}

//func GetTrashFiles(userID uint) ([]UserFileInfo, []Directory, error) {
//	files, dirs, err := getTrashFiles(userID)
//	trashDirs := []Directory{}
//	ancestorMap := make(map[string]string)
//	for _, dir := range dirs {
//
//	}
//}

func DeleteTrashFile(dirHash string, fileHash string) error {
	return db.Unscoped().Where("directory_hash = ? and file_hash = ?",
		dirHash, fileHash).Delete(&DirectoryFile{}).Error
}

func DeleteTrashDir(dirHash string) error {
	dirs, files, err := GetAllUnderTrashDir(dirHash)
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, file := range files {
			if err := tx.Unscoped().Where("directory_hash = ? and file_hash = ?",
				file.DirectoryHash, file.FileHash).Delete(&DirectoryFile{}).Error; err != nil {
				return err
			}
		}
		if len(dirs) != 0 {
			if err := tx.Unscoped().Delete(&dirs).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Delete(&Directory{Hash: dirHash}).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func ClearTrashFiles(userID uint) error {
	filesInfo, dirs, err := GetTrashFiles(userID)
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, file := range filesInfo {
			if err = tx.Unscoped().Where("directory_hash = ? and file_hash = ?",
				file.DirectoryHash, file.FileHash).Delete(&DirectoryFile{}).Error; err != nil {
				return err
			}
		}
		if len(dirs) != 0 {
			if err := tx.Unscoped().Delete(&dirs).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

func RestoreTrashFile(dirHash string, fileHash string) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&DirectoryFile{}).Where("directory_hash = ? and file_hash = ?",
			dirHash, fileHash).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// increase reference count of file and restore it if it has been deleted
		result := tx.Model(&File{}).Where("hash = ?", fileHash).
			Update("ref_count", gorm.Expr("ref_count + ?", 1))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			if err := tx.Unscoped().Model(&File{}).Where("hash = ?", fileHash).
				Updates(map[string]interface{}{
					"ref_count":  1,
					"deleted_at": nil,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func RestoreTrashDir(dirHash string) error {
	dirs, files, err := GetAllUnderTrashDir(dirHash)
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, file := range files {
			if err := RestoreTrashFile(file.DirectoryHash, file.FileHash); err != nil {
				return err
			}
		}
		for _, dir := range dirs {
			if err := tx.Unscoped().Model(&Directory{}).Where("hash = ?", dir.Hash).
				Update("deleted_at", nil).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Model(&Directory{}).Where("hash = ?", dirHash).
			Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}
