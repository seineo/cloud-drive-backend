package handler

import (
	"CloudDrive/config"
	"CloudDrive/middleware"
	"CloudDrive/model"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"time"
)

var DirStoragePath = config.GetConfig().Storage.DiskStoragePath

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files", middleware.AuthCheck)
	group.POST(":dirPath", uploadFile)
	group.GET(":dirPath", getFiles)
	group.GET("metadata/:dirPath", getFilesMetadata)
}

func uploadFile(c *gin.Context) {
	// get request data
	fileName := c.PostForm("fileName")
	hash := c.PostForm("hash")
	size := c.PostForm("size")
	contentType := c.PostForm("contentType")
	dirPath := c.Param("dirPath")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"message": "failed to upload file", "description": err.Error()})
		return
	}
	fileHandler, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to open uploaded file", "description": err.Error()})
		return
	}
	defer fileHandler.Close()
	session := sessions.Default(c)
	userID := session.Get("userID")
	// check whether file already exists
	exists, err := model.FileExists(hash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to check whether file exists", "description": err.Error()})
		return
	}
	if exists {
		c.JSON(409, gin.H{"message": "file exists"})
		return
	} else {
		// store file
		fileStoragePath := filepath.Join(DirStoragePath, fileName)
		if err := c.SaveUploadedFile(file, fileStoragePath); err != nil {
			c.JSON(500, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
			return
		}
		//store file metadata
		metadata := model.File{
			Hash:       hash,
			Name:       fileName,
			UserID:     userID.(uint),
			FileType:   contentType,
			Size:       size,
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

func getFiles(c *gin.Context) {
	//dirPath := c.Param("dirPath")
	//fileName := c.Query("fileName")

}
