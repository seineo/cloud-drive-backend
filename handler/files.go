package handler

import (
	"CloudDrive/middleware"
	"CloudDrive/model"
	"CloudDrive/request"
	"CloudDrive/response"
	"CloudDrive/service"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var FileStoragePath = configs.Storage.DiskStoragePath
var TempFileStoragePath = configs.Storage.DiskTempStoragePath
var MaxUploadSize = configs.MaxUploadSize
var ArchiveThreshold = configs.ArchiveThreshold

var chunkMutex sync.Mutex // write currentChunks in redis

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files", middleware.AuthCheck)
	group.POST("dir", createDir)
	group.POST("file", uploadFile)
	group.GET("dir/:dirHash", downloadDir)
	group.GET("file/:fileHash", downloadFile)

	group.GET("metadata/dir/:dirHash", getFilesMetadata)
	group.GET("metadata/file/:fileHash", fileExists)
	////group.POST("share/*dirPath", shareFiles)
	group.POST("chunks", uploadFileChunk)
	group.POST("chunks/:fileHash", mergeFileChunks)
	group.GET("chunks/:fileHash", getMissedChunks)
}

func createDir(c *gin.Context) {
	// bind request data
	var dirRequest request.DirectoryRequest
	if err := c.Bind(&dirRequest); err != nil {
		c.JSON(400, gin.H{"message": "request data is invalid", "description": err.Error()})
		return
	}
	// get user info
	session := sessions.Default(c)
	userID := session.Get("userID")
	// store directory metadata
	err := model.StoreDirMetadata(&dirRequest, userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to store file metadata", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"dir": dirRequest})
}

// upload a file given the file content and json-format metadata in form data
func uploadFile(c *gin.Context) {
	// get request metadata in json format
	var fileInfo request.FileRequest
	jsonStr := c.PostForm("metadata")
	err := json.Unmarshal([]byte(jsonStr), &fileInfo)
	if err != nil {
		c.JSON(400, gin.H{"message": "failed to unmarshal file metadata", "description": err.Error()})
		return
	}

	// get file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"message": "failed to upload file", "description": err.Error()})
		return
	}
	// check file size
	if fileInfo.FileSize > MaxUploadSize {
		c.JSON(400, gin.H{"message": fmt.Sprintf("Uploaded file %s is too big", file.Filename)})
		return
	}
	// store file content if not exists
	fileStoragePath := filepath.Join(FileStoragePath, fileInfo.FileHash)
	exists, err := model.FileExists(fileInfo.FileHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to check whether file exists", "description": err.Error()})
		return
	}
	if !exists {
		if err := c.SaveUploadedFile(file, fileStoragePath); err != nil {
			c.JSON(500, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
			return
		}
	}
	// store file metadata to database regardless of file existence
	err = model.StoreFileMetadata(&fileInfo, fileStoragePath, exists)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to store file metadata", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"file": fileInfo})
}

// get metadata of all files under given directory
func getFilesMetadata(c *gin.Context) {
	dirHash := c.Param("dirHash")
	// get user info
	session := sessions.Default(c)
	userID := session.Get("userID")
	log.WithFields(logrus.Fields{
		"dirHash": dirHash,
		"userID":  userID,
	}).Info("trying to get file metadata")
	// get metadata of all the files under the directory
	files, dirs, err := model.GetFilesMetadata(dirHash)
	if err != nil {
		c.JSON(500, gin.H{"message": fmt.Sprintf("failed to get files and dirs under dir %s", dirHash), "description": err.Error()})
		return
	}
	// construct files in response
	var fileResponses []response.FileResponse
	for _, dir := range dirs {
		fileResponses = append(fileResponses, response.FileResponse{
			Hash:      dir.Hash,
			Name:      dir.Name,
			Type:      "dir",
			Size:      0,
			CreatedAt: dir.CreatedAt,
		})
	}
	for _, file := range files {
		fileResponses = append(fileResponses, response.FileResponse{
			Hash:      file.Hash,
			Name:      file.Name,
			Type:      file.Type,
			Size:      file.Size,
			CreatedAt: file.CreatedAt,
		})
	}
	c.JSON(200, gin.H{"files": fileResponses})
	return
}

// download the whole directory and return zipped result
func downloadDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	path := c.Query("path")
	err := service.ArchiveDir(path, dirHash, filepath.Join(TempFileStoragePath, dirHash))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to archive directory", "description": err.Error()})
		return
	}
	log.WithFields(logrus.Fields{
		"hash":       dirHash,
		"path":       path,
		"zippedPath": filepath.Join(TempFileStoragePath, dirHash),
	}).Info("directory archived")
	file, err := os.Open(filepath.Join(TempFileStoragePath, dirHash))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
		return
	}
	defer func() { // delete the temporal zipped file
		err := os.Remove(filepath.Join(TempFileStoragePath, dirHash))
		if err != nil {
			log.WithFields(logrus.Fields{
				"filePath": filepath.Join(TempFileStoragePath, dirHash),
			}).Error("failed to remove temporal zipped file")
		}
	}()
	// download zip file and name it with extension `zip`
	slice := strings.Split(path, "/")
	zipName := slice[len(slice)-1] + ".zip"
	log.Debug("zipName: ", zipName)
	c.Header("Content-Type", "application/octet-stream") // binary stream
	c.Header("Content-Disposition", "attachment; filename="+zipName)
	c.Header("Content-Encoding", "zip")
	io.Copy(c.Writer, file)
}

// download normal file
// if target file exceeds size threshold, return zipped result
// otherwise return the file itself
func downloadFile(c *gin.Context) {
	fileHash := c.Param("fileHash")
	fileName := c.Query("fileName")
	if fileHash == "" || fileName == "" {
		c.JSON(400, gin.H{"message": "fileHash and fileName cannot be empty"})
		return
	}
	// get file metadata
	fileInfo, err := model.GetFileMetadata(fileHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get file metadata", "description": err.Error()})
	}
	// if size exceeds the threshold, we zip file and name the zipped file by file name.
	// Not for image, video and audio files since they have been archived to some extent
	isArchived := false
	log.Debug("file size: ", fileInfo.Size)
	log.Debug("threshold: ", ArchiveThreshold)
	if fileInfo.Size > ArchiveThreshold &&
		!strings.HasPrefix(fileInfo.FileType, "image") && !strings.HasPrefix(fileInfo.FileType, "audio") &&
		!strings.HasPrefix(fileInfo.FileType, "video") {
		isArchived = true
		err = service.ArchiveFile(fileInfo.Location, fileName, filepath.Join(TempFileStoragePath, fileInfo.Hash))
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to archive file", "description": err.Error()})
			return
		}
		log.WithFields(logrus.Fields{
			"fileHash":   fileInfo.Hash,
			"fileName":   fileName,
			"zippedPath": filepath.Join(TempFileStoragePath, fileInfo.Hash),
		}).Info("file archived")
	}
	// write response header
	c.Header("Content-Type", "application/octet-stream") // binary stream
	// return the file
	if isArchived {
		file, err := os.Open(filepath.Join(TempFileStoragePath, fileInfo.Hash))
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
			return
		}
		defer func() { // delete the temporal zipped file
			err := os.Remove(filepath.Join(TempFileStoragePath, fileInfo.Hash))
			if err != nil {
				log.WithFields(logrus.Fields{
					"filePath": filepath.Join(TempFileStoragePath, fileInfo.Hash),
				}).Error("failed to remove temporal zipped file")
			}
		}()
		// download zip file and name it with extension `zip`
		zipName := strings.Split(fileName, ".")[0] + ".zip"
		log.Debug("zipName: ", zipName)
		c.Header("Content-Disposition", "attachment; filename="+zipName)
		c.Header("Content-Encoding", "zip")
		io.Copy(c.Writer, file)
	} else {
		file, err := os.Open(fileInfo.Location)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
			return
		}
		defer file.Close()
		c.Header("Content-Disposition", "attachment; filename="+fileName)
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
//func shareFiles(c *gin.Context) {
//	// get current user
//	session := sessions.Default(c)
//	ownerID := session.Get("userID")
//	user, err := model.GetUserByID(ownerID.(uint))
//	if err != nil {
//		c.JSON(500, gin.H{"message": "failed to find user by id", "description": err.Error()})
//		return
//	}
//	// common fields
//	dirPath := c.Param("dirPath")
//	isLimited := c.PostForm("isLimited")
//	expiredTime := c.PostForm("expiredTime")
//	userRole, err := strconv.Atoi(c.PostForm("userRole"))
//	if err != nil {
//		c.JSON(400, gin.H{"message": "user role should be 0 or 1"})
//		return
//	}
//	fileNames := c.PostFormArray("fileNames")
//
//	var share model.Share
//
//	// get file hash
//	var filesHash []string
//	for i := 0; i < len(fileNames); i++ {
//		fileMetadata, err := model.GetFileMetadata(ownerID.(uint), dirPath, fileNames[i])
//		if err != nil {
//			c.JSON(500, gin.H{"message": "failed to get file metadata", "description": err.Error()})
//			return
//		}
//		filesHash = append(filesHash, fileMetadata.Hash)
//	}
//
//	if isLimited == "true" {
//		emails := c.PostFormArray("emails")
//		content := c.PostForm("content")
//		// send emails to users, and generate share info
//		for i := 0; i < len(emails); i++ {
//			var sharedIDs []string
//			var sharedLinks []string
//			for j := 0; j < len(fileNames); j++ {
//				// generate shared links, each file for each email has a unique link
//				sharedID := uuid.NewString()
//				sharedIDs = append(sharedIDs, sharedID)
//				sharedLink := fmt.Sprintf("%s/files/%s", config.GetConfig().ProjectURL, sharedID)
//				sharedLinks = append(sharedLinks, sharedLink)
//			}
//			// send email
//			err := service.SendShareEmails(user.Name, user.Email, emails[i], content, fileNames, sharedLinks)
//			if err != nil {
//				c.JSON(500, gin.H{"message": "failed to send file sharing email", "description": err.Error()})
//				return
//			}
//			// generate share info
//			for j := 0; j < len(fileNames); j++ {
//				sharedUser, _ := model.GetUserByEmail(emails[i])
//				share = model.Share{
//					SharedID:    sharedIDs[j],
//					FileHash:    filesHash[j],
//					OwnerID:     ownerID.(uint),
//					UserID:      service.GetSharedUserIDPtr(sharedUser),
//					UserRole:    uint(userRole),
//					Password:    nil,
//					AccessCount: 0,
//					SharedTime:  time.Now(),
//					ExpiredTime: service.GetShareExpiredTimePtr(expiredTime),
//					IsLimited:   true,
//				}
//			}
//		}
//	} else if isLimited == "false" { // shared to the public
//		password := c.PostForm("password")
//		for j := 0; j < len(fileNames); j++ {
//			sharedID := uuid.NewString()
//			share = model.Share{
//				SharedID:    sharedID,
//				FileHash:    filesHash[j],
//				OwnerID:     ownerID.(uint),
//				UserID:      nil,
//				UserRole:    uint(userRole),
//				Password:    service.GetPasswordPtr(password),
//				AccessCount: 0,
//				SharedTime:  time.Now(),
//				ExpiredTime: service.GetShareExpiredTimePtr(expiredTime),
//				IsLimited:   false,
//			}
//		}
//	} else {
//		c.JSON(400, gin.H{"message": "field `isLimited` can only either be `true` or `false`"})
//		return
//	}
//	// store share info to database
//	if err = model.CreateShare(&share); err != nil {
//		c.JSON(500, gin.H{"message": "failed to store share info", "description": err.Error()})
//		return
//	}
//	c.JSON(200, gin.H{"share": share})
//}

// check whether file exists using hash
func fileExists(c *gin.Context) {
	hash := c.Param("fileHash")
	exists, err := model.FileExists(hash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to check whether file exists", "description": err.Error()})
		return
	}
	if exists {
		c.JSON(200, gin.H{"exist": true})
		return
	} else {
		c.JSON(200, gin.H{"exist": false})
		return
	}
}

// store file chunks on disk and related info in redis
func uploadFileChunk(c *gin.Context) {
	// get request chunk info
	var chunkRequest request.ChunkRequest
	jsonStr := c.PostForm("metadata")
	err := json.Unmarshal([]byte(jsonStr), &chunkRequest)
	if err != nil {
		c.JSON(400, gin.H{"message": "failed to unmarshal file metadata", "description": err.Error()})
		return
	}
	var tempFileDir = filepath.Join(TempFileStoragePath, chunkRequest.FileHash)
	// store chunk info in redis
	chunkMutex.Lock()
	defer chunkMutex.Unlock()
	val, err := rdb.Get(ctx, chunkRequest.FileHash).Result()
	var chunkInfo request.CurrentChunks
	if err == redis.Nil { // not exists,  the first chunk of the file
		chunkInfo = request.CurrentChunks{
			TotalChunks: chunkRequest.TotalChunks,
			Indexes:     map[uint]string{chunkRequest.Index: chunkRequest.ChunkHash},
		}
		// create directory for chunks of this file
		os.Mkdir(tempFileDir, 0755)
	} else if err != nil { // the first chunk of the file
		c.JSON(500, gin.H{"message": "failed to read file chunk info from redis", "description": err.Error()})
		return
	} else { // not the first chunk of the file
		err = json.Unmarshal([]byte(val), &chunkInfo)
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to unmarshal chunk info", "description": err.Error()})
			return
		}
		chunkInfo.Indexes[chunkRequest.Index] = chunkRequest.ChunkHash
	}
	chunkJson, err := json.Marshal(chunkInfo)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to marshal chunk info", "description": err.Error()})
		return
	}
	err = rdb.Set(ctx, chunkRequest.FileHash, string(chunkJson), 0).Err()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to set chunk info", "description": err.Error()})
		return
	}
	// store file chunk on disk
	chunk, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(400, gin.H{"message": "miss chunk data", "description": err.Error()})
		return
	}
	if err := c.SaveUploadedFile(chunk, filepath.Join(tempFileDir, chunkRequest.ChunkHash)); err != nil {
		c.JSON(500, gin.H{"message": "failed to store uploaded file", "description": err.Error()})
		return
	}
	//blob, err := base64.StdEncoding.DecodeString(chunkRequest.Blob)
	//if err != nil {
	//	c.JSON(500, gin.H{"message": "failed to decode chunk blob data", "description": err.Error()})
	//	return
	//}
	//err = os.WriteFile(filepath.Join(tempFileDir, chunkRequest.ChunkHash), blob, 0644)
	//if err != nil {
	//	c.JSON(500, gin.H{"message": "failed to store file chunk", "description": err.Error()})
	//	return
	//}
	log.WithFields(logrus.Fields{
		"fileHash":  chunkRequest.FileHash,
		"chunkHash": chunkRequest.ChunkHash,
	}).Infof("stored file chunk")
	c.JSON(200, gin.H{"chunkIndex": chunkRequest.Index, "fileHash": chunkRequest.FileHash})
}

// merge uploaded chunks of a file and store it into another file
func mergeFileChunks(c *gin.Context) {
	fileHash := c.Param("fileHash")
	var merge request.FileRequest
	err := c.Bind(&merge)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid request data", "description": err.Error()})
		return
	}
	// set source files and target file
	chunkDir := filepath.Join(TempFileStoragePath, fileHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to read chunk directory", "description": err.Error()})
		return
	}
	targetPath := filepath.Join(FileStoragePath, fileHash)
	targetFile, err := os.OpenFile(targetPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to open target storage file", "description": err.Error()})
		return
	}
	// get chunk order info from redis
	chunkMutex.Lock()
	chunkJson, err := rdb.Get(ctx, fileHash).Result()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to store file chunk", "description": err.Error()})
		return
	}
	var chunkInfo request.CurrentChunks
	err = json.Unmarshal([]byte(chunkJson), &chunkInfo)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to unmarshal chunk info", "description": err.Error()})
		return
	}
	chunkMutex.Unlock()
	// check whether all chunks are uploaded
	for i := 1; uint(i) <= chunkInfo.TotalChunks; i++ {
		_, ok := chunkInfo.Indexes[uint(i)]
		if !ok {
			c.JSON(400, gin.H{"message": "failed to merge chunks", "missedChunk": i})
			return
		}
	}
	// read file in order and write into target file
	for i := 1; uint(i) <= chunkInfo.TotalChunks; i++ {
		chunkHash := chunkInfo.Indexes[uint(i)]
		chunkData, err := os.ReadFile(filepath.Join(chunkDir, chunkHash))
		if err != nil {
			c.JSON(500, gin.H{"message": "failed to read chunk file", "description": err.Error()})
			return
		}
		_, err = targetFile.Write(chunkData)
		if err != nil {
			os.Remove(targetPath)
			c.JSON(500, gin.H{"message": "failed to write all the chunk data to target", "description": err.Error()})
			return
		}
	}
	// delete chunk info in redis
	err = rdb.Del(ctx, fileHash).Err()
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to delete chunk info in redis", "description": err.Error()})
		return
	}
	// delete temporal chunks
	os.RemoveAll(chunkDir)
	// store file info in mysql
	session := sessions.Default(c)
	userID := session.Get("userID")
	err = model.StoreFileMetadata(&merge, targetPath, false)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to store merged file info", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"fileName": merge.FileName, "userID": userID})
}

// To achieve breakpoint resume, obtain the not yet uploaded file chunks.
func getMissedChunks(c *gin.Context) {
	fileHash := c.Param("fileHash")
	// get indexes of uploaded chunks from redis
	chunkMutex.Lock()
	defer chunkMutex.Unlock()
	chunkJson, err := rdb.Get(ctx, fileHash).Result()
	if err == redis.Nil { // not exists
		c.JSON(200, gin.H{"exists": false})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"message": "failed to get file chunk", "description": err.Error()})
		return
	}
	var chunkInfo request.CurrentChunks
	err = json.Unmarshal([]byte(chunkJson), &chunkInfo)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to unmarshal chunk info", "description": err.Error()})
		return
	}
	// find missed chunks
	missedChunks := []uint{}
	for i := 1; uint(i) <= chunkInfo.TotalChunks; i++ {
		equal := false
		var index uint
		for index = range chunkInfo.Indexes {
			if index == uint(i) {
				equal = true
				break
			}
		}
		if !equal {
			missedChunks = append(missedChunks, index)
		}
	}
	c.JSON(200, gin.H{"exists": true, "missedChunks": missedChunks})
}
