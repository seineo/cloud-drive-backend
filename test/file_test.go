package test

import (
	"CloudDrive/model"
	"CloudDrive/service"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//	     zip-root/
//	/       \      \      \
//
// file1    file2    dir1/   dir2/
//
//	|         |        \
//
// content1  content2   file3
var files = []model.File{
	{
		Hash:       randString(10),
		Name:       "zip-root",
		UserID:     0,
		FileType:   "dir",
		Size:       0,
		DirPath:    "/root",
		Location:   "/Users/liyuewei/Desktop/zip-root",
		CreateTime: time.Now(),
	},
	{
		Hash:       randString(10),
		Name:       "file1",
		UserID:     0,
		FileType:   "file",
		Size:       0,
		DirPath:    "/root/zip-root",
		Location:   "/Users/liyuewei/Desktop/zip-root/file1",
		CreateTime: time.Now(),
	},
	{
		Hash:       randString(10),
		Name:       "file2",
		UserID:     0,
		FileType:   "file",
		Size:       0,
		DirPath:    "/root/zip-root",
		Location:   "/Users/liyuewei/Desktop/zip-root/file2",
		CreateTime: time.Now(),
	},
	{
		Hash:       randString(10),
		Name:       "dir1",
		UserID:     0,
		FileType:   "dir",
		Size:       0,
		DirPath:    "/root/zip-root",
		Location:   "/Users/liyuewei/Desktop/zip-root/dir1",
		CreateTime: time.Now(),
	},
	{
		Hash:       randString(10),
		Name:       "dir2",
		UserID:     0,
		FileType:   "dir",
		Size:       0,
		DirPath:    "/root/zip-root",
		Location:   "/Users/liyuewei/Desktop/zip-root/dir2",
		CreateTime: time.Now(),
	},
	{
		Hash:       randString(10),
		Name:       "file3",
		UserID:     0,
		FileType:   "file",
		Size:       0,
		DirPath:    "/root/zip-root/dir1",
		Location:   "/Users/liyuewei/Desktop/zip-root/dir1/file3",
		CreateTime: time.Now(),
	},
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// given path and file type, `create file or directory
func create(p string, fileType string) (*os.File, error) {
	if fileType == "dir" {
		if err := os.MkdirAll(p, 0770); err != nil {
			return nil, err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
			return nil, err
		}
		return os.Create(p)
	}
	return nil, nil
}

func prepareFiles() {
	for _, file := range files {
		// store metadata to mysql
		err := model.StoreFileMetadata(&file)
		if err != nil {
			log.WithError(err).Error("failed to store file metadata to mysql")
			return
		}
		// create file at given location and write some strings
		fileHandler, err := create(file.Location, file.FileType)
		if err != nil {
			log.WithError(err).Error("failed to create file")
			return
		}
		if file.FileType == "file" {
			_, err := fileHandler.WriteString(file.Name)
			if err != nil {
				log.WithError(err).Error("failed to write string to file")
				return
			}
		}
	}
}

func clearFiles() {
	for _, file := range files {
		// delete metadata from mysql
		if err := model.DeleteFilesMetadata(0, file.DirPath); err != nil {
			log.WithError(err).Error("failed to delete file metadata from mysql")
			return
		}
		if err := os.RemoveAll(file.Location); err != nil {
			log.WithError(err).Error("failed to delete file in disk")
			return
		}
	}
}

func TestArchiveFile(t *testing.T) {
	// prepare data: create file in disk and store metadata in mysql
	prepareFiles()
	// test function
	err := service.ArchiveFile(0, "/root", "zip-root", "/Users/liyuewei/Desktop/zip-result.zip")
	if err != nil {
		log.WithError(err).Error("failed to zip file")
		return
	}
	//clear data
	clearFiles()
}
