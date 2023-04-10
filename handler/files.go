package handler

import (
	"CloudDrive/config"
	"CloudDrive/middleware"
	"CloudDrive/model"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var DirStoragePath = config.GetConfig().Storage.DiskStoragePath

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files", middleware.AuthCheck)
	group.POST("data/*dirPath", uploadFile)
	group.GET("data/*dirPath", getFiles)
	// we don't need metadata of specific file, since front end would show all files in a directory
	group.GET("metadata/*dirPath", getFilesMetadata)
}

func uploadFile(c *gin.Context) {
	// get request data
	fileName := c.PostForm("fileName")
	hash := c.PostForm("hash")
	contentType := c.PostForm("contentType")
	dirPath := c.Param("dirPath")
	if fileName == "" || hash == "" || contentType == "" || dirPath == "" {
		c.JSON(400, gin.H{"message": "form data cannot be empty"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"message": "failed to upload file", "description": err.Error()})
		return
	}
	// check file size
	if file.Size > config.GetConfig().MaxUploadSize {
		c.JSON(400, gin.H{"message": fmt.Sprintf("Uploaded file %s is too big", file.Filename)})
		return
	}
	// check whether file already exists
	exists, err := model.FileExists(hash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to check whether file exists", "description": err.Error()})
		return
	}
	// if exists, conflict
	if exists {
		c.JSON(409, gin.H{"message": "file exists"})
		return
	} else { // if not exists, store the file in stream
		fileStoragePath := filepath.Join(DirStoragePath, hash)
		if err := c.SaveUploadedFile(file, fileStoragePath); err != nil {
			c.JSON(500, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
			return
		}
		// get user info
		session := sessions.Default(c)
		userID := session.Get("userID")
		//store file metadata
		metadata := model.File{
			Hash:       hash,
			Name:       fileName,
			UserID:     userID.(uint),
			FileType:   contentType,
			Size:       file.Size,
			DirPath:    dirPath,
			Location:   fileStoragePath,
			CreateTime: time.Now(),
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

// download directory or file, both need its directory path and name
func getFiles(c *gin.Context) {
	//dirPath := c.Param("dirPath")
	//fileName := c.Query("fileName")
	//if dirPath == "" || fileName == "" {
	//	c.JSON(400, gin.H{"message": "dirPath and fileName cannot be empty"})
	//	return
	//}
	//service.ArchiveFile(dirPath, fileName, )
}
