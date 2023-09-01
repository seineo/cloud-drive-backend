package service

import (
	"CloudDrive/model"
	"CloudDrive/response"
	"archive/zip"
	"github.com/robfig/cron/v3"
	"io"
	"os"
	"path/filepath"
)

// FileInfo is a data structure used by Walk to pass file information
type FileInfo struct {
	IsDir    bool
	Hash     string
	Name     string
	Location string // not empty only when !isDir
}

// MyWalkFunc is a user-defined function called by Walk to visit each file or directory.
// Errors in Walk will be passed to MyWalkFunc to deal with,
// and the errors thrown by MyWalkFunc will be thrown by Walk then.
type MyWalkFunc func(path string, fileInfo FileInfo, err error) error

// Walk descends path and calls walkFn for each file or directory.
// (This function is implemented with reference to the filepath.Walk function in the standard library.)
func Walk(path string, fileInfo FileInfo, walkFn MyWalkFunc) error {
	if !fileInfo.IsDir {
		return walkFn(path, fileInfo, nil)
	}
	files, dirs, err := model.GetFilesMetadata(fileInfo.Hash, false, "", "")
	if err = walkFn(path, fileInfo, err); err != nil {
		return err
	}
	for _, file := range files {
		filePath := filepath.Join(path, file.Name)
		subFileInfo := FileInfo{
			IsDir:    false,
			Hash:     file.FileHash,
			Name:     file.Name,
			Location: file.Location,
		}
		if err = walkFn(filePath, subFileInfo, nil); err != nil {
			return err
		}
	}
	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir.Name)
		subDirInfo := FileInfo{
			IsDir: true,
			Hash:  dir.Hash,
			Name:  dir.Name,
		}
		if err = Walk(dirPath, subDirInfo, walkFn); err != nil {
			return err
		}
	}
	return nil
}

// ArchiveDir archives a directory to dstPath, given its path(as zip root) and hash.
func ArchiveDir(root string, hash string, dstPath string) error {
	// create a zip file and zip.Writer
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()
	// find the directory from db
	dir, err := model.GetDirMetadata(hash)
	if err != nil {
		return err
	}
	dirInfo := FileInfo{
		IsDir: true,
		Hash:  hash,
		Name:  dir.Name,
	}
	// define MyWalkFn
	walker := func(path string, fileInfo FileInfo, err error) error {
		if err != nil { // throw the error that Walk passes
			return err
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// create file header
		header := &zip.FileHeader{
			Name:   relativePath,
			Method: zip.Deflate,
		}
		if fileInfo.IsDir { //  if isDir, we will ignore empty directory if return nil directly
			header.Name += "/"
			header.SetMode(0755)
			_, err = writer.CreateHeader(header)
			return err
		}
		// file type is not directory
		zipFile, err := writer.CreateHeader(header)
		if err != nil {
			log.WithError(err).Error("failed to write file header")
			return err
		}
		// write file content to zip
		file, err := os.Open(fileInfo.Location)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}
		return nil
	}
	err = Walk(root, dirInfo, walker)
	return err
}

// ArchiveFile archives single file given its storage location, name shown for users and destination path
func ArchiveFile(location string, fileName string, dstPath string) error {
	// create a zip file and zip.Writer
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// create file header
	header := &zip.FileHeader{
		Name:   fileName,
		Method: zip.Deflate,
	}
	// write file header to zip
	zipFile, err := writer.CreateHeader(header)
	if err != nil {
		log.WithError(err).Error("failed to write file header")
		return err
	}
	// write file content to zip
	file, err := os.Open(location)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(zipFile, file)

	return nil
}

func ScheduleDeleteStaleFiles() {
	c := cron.New()
	c.AddFunc(configs.File.StaleTimeCron, model.DeleteStaleFiles)
	c.Start()
}

func Convert2FileResponse(files []model.UserFileInfo, dirs []model.Directory) []response.FileResponse {
	fileResponses := []response.FileResponse{}
	for _, dir := range dirs {
		fileResponses = append(fileResponses, response.FileResponse{
			DirectoryHash: *dir.ParentHash,
			FileHash:      dir.Hash,
			Name:          dir.Name,
			Type:          "dir",
			Size:          0,
			IsStarred:     dir.IsStarred,
			CreatedAt:     dir.CreatedAt,
			DeletedAt:     dir.DeletedAt,
		})
	}
	for _, file := range files {
		fileResponses = append(fileResponses, response.FileResponse{
			DirectoryHash: file.DirectoryHash,
			FileHash:      file.FileHash,
			Name:          file.Name,
			Type:          file.Type,
			Size:          file.Size,
			IsStarred:     file.IsStarred,
			CreatedAt:     file.CreatedAt,
			DeletedAt:     file.DeletedAt,
		})
	}
	return fileResponses
}
