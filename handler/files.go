package handler

import (
	"CloudDrive/config"
	"CloudDrive/middleware"
	"CloudDrive/model"
	"CloudDrive/service"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var FileStoragePath = config.GetConfig().Storage.DiskStoragePath
var TempFileStoragePath = config.GetConfig().Storage.DiskTempStoragePath
var MaxUploadSize = config.GetConfig().MaxUploadSize
var ArchiveThreshold = config.GetConfig().ArchiveThreshold

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files", middleware.AuthCheck)
	group.POST("data/*dirPath", uploadFile)
	group.GET("data/*dirPath", getFiles)
	// we don't need metadata of specific file, since front end would show all files in a directory
	group.GET("metadata/*dirPath", getFilesMetadata)
}

// upload file or create a directory given its directory path in url and its file/dir name in form data
func uploadFile(c *gin.Context) {
	// get request form data
	fileName := c.PostForm("fileName")
	hash := c.PostForm("hash")
	fileType := c.PostForm("fileType")
	dirPath := c.Param("dirPath")
	if fileName == "" || hash == "" || fileType == "" || dirPath == "" {
		c.JSON(400, gin.H{"message": "form data cannot be empty"})
		return
	}
	// check whether file already exists
	exists, err := model.FileExists(hash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to check whether file exists", "description": err.Error()})
		return
	}
	var metadata model.File
	// if exists, conflict
	if exists {
		c.JSON(409, gin.H{"message": "file exists"})
		return
	} else {
		// get user info
		session := sessions.Default(c)
		userID := session.Get("userID")
		// store different metadata depending on file type
		if fileType == "dir" {
			metadata = model.File{
				Hash:       hash,
				Name:       fileName,
				UserID:     userID.(uint),
				FileType:   fileType,
				Size:       0,
				DirPath:    dirPath,
				Location:   "",
				CreateTime: time.Now(),
			}
		} else {
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(400, gin.H{"message": "failed to upload file", "description": err.Error()})
				return
			}
			// check file size
			if file.Size > MaxUploadSize {
				c.JSON(400, gin.H{"message": fmt.Sprintf("Uploaded file %s is too big", file.Filename)})
				return
			}
			fileStoragePath := filepath.Join(FileStoragePath, hash)
			if err := c.SaveUploadedFile(file, fileStoragePath); err != nil {
				c.JSON(500, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
				return
			}
			//store file metadata
			metadata = model.File{
				Hash:       hash,
				Name:       fileName,
				UserID:     userID.(uint),
				FileType:   fileType,
				Size:       file.Size,
				DirPath:    dirPath,
				Location:   fileStoragePath,
				CreateTime: time.Now(),
			}
		}
		err = model.StoreFileMetadata(&metadata)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to store file metadata", "description": err.Error()})
			return
		}
		c.JSON(200, gin.H{"file": metadata})
	}
}

func getFilesMetadata(c *gin.Context) {
	dirPath := c.Param("dirPath")
	log.WithFields(logrus.Fields{
		"dirPath": dirPath,
	}).Info("get dirPath")
	// get metadata of all the files under the directory
	files, err := model.GetFilesMetadata(dirPath)
	if err != nil {
		c.JSON(500, gin.H{"message": fmt.Sprintf("failed to get files under dir %s", dirPath),
			"description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"files": files})
	return

}

// download directory or normal file, both need its directory path and name
// if target is dir or file exceeds specific size, return the zipped result
// else return the file itself
func getFiles(c *gin.Context) {
	dirPath := c.Param("dirPath")
	fileName := c.Query("fileName")
	if dirPath == "" || fileName == "" {
		c.JSON(400, gin.H{"message": "dirPath and fileName cannot be empty"})
		return
	}
	// get file metadata
	fileInfo, err := model.GetFileMetadata(dirPath, fileName)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get file metadata", "description": err.Error()})
	}
	// if file is dir or its size exceeds the threshold, we zip file and name the zipped file by file hash
	// not for image, video and audio files since they have been archived to some extent
	isArchived := false
	if fileInfo.FileType == "dir" || (fileInfo.Size > ArchiveThreshold &&
		!strings.HasPrefix(fileInfo.FileType, "image") && !strings.HasPrefix(fileInfo.FileType, "audio") &&
		!strings.HasPrefix(fileInfo.FileType, "video")) {
		isArchived = true
		err = service.ArchiveFile(dirPath, fileName, filepath.Join(TempFileStoragePath, fileInfo.Hash))
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to archive file", "description": err.Error()})
			return
		}
		log.WithFields(logrus.Fields{
			"fileName":   fileName,
			"zippedPath": filepath.Join(TempFileStoragePath, fileInfo.Hash),
		}).Info("file archived")
	}
	// write response header
	c.Header("Content-Disposition", "attachment; filename="+fileName) // download named by filename
	c.Header("Content-Type", "application/octet-stream")              // binary stream
	// return the file
	if isArchived {
		file, err := os.Open(filepath.Join(TempFileStoragePath, fileInfo.Hash))
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
			return
		}
		defer file.Close()
		c.Header("Content-Encoding", "zip")
		io.Copy(c.Writer, file)
	} else {
		file, err := os.Open(fileInfo.Location)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
			return
		}
		defer file.Close()
		io.Copy(c.Writer, file)
	}
}
