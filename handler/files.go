package handler

import (
	"CloudDrive/config"
	"crypto/md5"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"path/filepath"
)

var DirStoragePath = config.GetConfig().Storage.DiskStoragePath

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files")
	group.POST(":dirPath", uploadFile)
	group.GET("", getDirFiles)
	group.GET(":filePath", getFile)
}

func uploadFile(c *gin.Context) { // TODO 使用form的dir字段判断是不是新建文件夹
	// get uploaded file data
	file, err := c.FormFile("file")
	fileName := c.PostForm("fileName")
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
	// hash file using MD5 (fast)
	hashMD5 := md5.New()
	if _, err := io.Copy(hashMD5, fileHandler); err != nil {
		c.JSON(500, gin.H{"message": "failed to hash uploaded file", "description": err.Error()})
		return
	}
	log.WithFields(logrus.Fields{
		"hash":     hashMD5,
		"fileName": fileName,
	}).Info("get md5 hash of file content")
	// store file
	filePath := filepath.Join(DirStoragePath, fileName)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(400, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
		return
	}
	//store file metadata
	//dirPath := c.Param("dirPath")

}

func getDirFiles(c *gin.Context) {

}

func getFile(c *gin.Context) {

}
