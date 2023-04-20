package handler

import (
	"CloudDrive/config"
	"CloudDrive/middleware"
	"CloudDrive/model"
	"CloudDrive/service"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	group.POST("share/*dirPath", shareFiles)
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

// get metadata of all files under given directory
func getFilesMetadata(c *gin.Context) {
	dirPath := c.Param("dirPath")
	// get user info
	session := sessions.Default(c)
	userID := session.Get("userID")
	log.WithFields(logrus.Fields{
		"dirPath": dirPath,
		"userID":  userID,
	}).Info("trying to get file metadata")
	// get metadata of all the files under the directory
	files, err := model.GetFilesMetadata(userID.(uint), dirPath)
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
	// get user info
	session := sessions.Default(c)
	userID := session.Get("userID")
	// get file metadata
	fileInfo, err := model.GetFileMetadata(userID.(uint), dirPath, fileName)
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
		err = service.ArchiveFile(userID.(uint), dirPath, fileName, filepath.Join(TempFileStoragePath, fileInfo.Hash))
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

// share all contents under directories or specific files
// if files are shared to registered users (type: limit),
//
//	it needs mail list, corresponding mail content, expired time and user role;
//
// otherwise (shared to the public)
//
//	we need user role, expired time and password(optional)
func shareFiles(c *gin.Context) {
	// get current user
	session := sessions.Default(c)
	ownerID := session.Get("userID")
	user, err := model.GetUserByID(ownerID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to find user by id", "description": err.Error()})
		return
	}
	// common fields
	dirPath := c.Param("dirPath")
	isLimited := c.PostForm("isLimited")
	expiredTime := c.PostForm("expiredTime")
	userRole, err := strconv.Atoi(c.PostForm("userRole"))
	if err != nil {
		c.JSON(400, gin.H{"message": "user role should be 0 or 1"})
		return
	}
	fileNames := c.PostFormArray("fileNames")

	var share model.Share

	// get file hash
	var filesHash []string
	for i := 0; i < len(fileNames); i++ {
		fileMetadata, err := model.GetFileMetadata(ownerID.(uint), dirPath, fileNames[i])
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to get file metadata", "description": err.Error()})
			return
		}
		filesHash[i] = fileMetadata.Hash
	}

	if isLimited == "true" {
		emails := c.PostFormArray("emails")
		content := c.PostForm("emailContent")
		// send emails to users, and generate share info
		for i := 0; i < len(emails); i++ {
			var sharedIDs []string
			var sharedLinks []string
			for j := 0; j < len(fileNames); j++ {
				// generate shared links, each file for each email has a unique link
				sharedID := uuid.NewString()
				sharedIDs = append(sharedIDs, sharedID)
				sharedLink := fmt.Sprintf("%s/files/%s", config.GetConfig().ProjectURL, sharedID)
				sharedLinks = append(sharedLinks, sharedLink)
			}
			// send email
			err := service.SendShareEmails(user.Name, user.Email, emails[i], content, fileNames, sharedLinks)
			if err != nil {
				c.JSON(500, gin.H{"message": "failed to send file sharing email", "description": err.Error()})
				return
			}
			// generate share info
			for j := 0; j < len(fileNames); j++ {
				sharedUser, _ := model.GetUserByEmail(emails[i])
				share = model.Share{
					SharedID:    sharedIDs[j],
					FileHash:    filesHash[j],
					OwnerID:     ownerID.(uint),
					UserID:      service.GetSharedUserIDPtr(sharedUser),
					UserRole:    uint(userRole),
					Password:    nil,
					AccessCount: 0,
					SharedTime:  time.Now(),
					ExpiredTime: service.GetShareExpiredTimePtr(expiredTime),
					IsLimited:   true,
				}
			}
		}
	} else if isLimited == "false" { // shared to the public
		password := c.PostForm("password")
		for j := 0; j < len(fileNames); j++ {
			sharedID := uuid.NewString()
			share = model.Share{
				SharedID:    sharedID,
				FileHash:    filesHash[j],
				OwnerID:     ownerID.(uint),
				UserID:      nil,
				UserRole:    uint(userRole),
				Password:    service.GetPasswordPtr(password),
				AccessCount: 0,
				SharedTime:  time.Now(),
				ExpiredTime: service.GetShareExpiredTimePtr(expiredTime),
				IsLimited:   false,
			}
		}
	} else {
		c.JSON(400, gin.H{"message": "field `isLimited` can only either be `true` or `false`"})
		return
	}
	// store share info to database
	if err = model.CreateShare(&share); err != nil {
		c.JSON(500, gin.H{"message": "failed to store share info", "description": err.Error()})
	}
	c.JSON(200, gin.H{"share": share})
}
