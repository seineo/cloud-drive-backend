package handler

import (
	"CloudDrive/middleware"
	"CloudDrive/model"
	"CloudDrive/request"
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

//var MaxUploadSize = configs.MaxUploadSize
//var ArchiveThreshold = configs.ArchiveThreshold

var chunkMutex sync.Mutex // write currentChunks in redis

func RegisterFilesRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/files", middleware.AuthCheck)
	group.POST("dir", createDir)
	group.POST("file", uploadFile)
	group.GET("dir/:dirHash", downloadDir)
	group.DELETE("dir/:dirHash", deleteDir)
	group.GET("file/:fileHash", downloadFile)
	group.DELETE("file/:dirHash/:fileHash", deleteFile)

	group.GET("metadata/dir/:dirHash", getFilesMetadata)
	group.GET("metadata/dir/:dirHash/trace", getTraceDirs)
	group.GET("metadata/file/:fileHash", fileExists)

	group.GET("metadata/star", getStarredFiles)
	group.PUT("metadata/dir/:dirHash/star", starDir)
	group.DELETE("metadata/dir/:dirHash/star", unstarDir)
	group.PUT("metadata/file/:dirHash/:fileHash/star", starFile)
	group.DELETE("metadata/file/:dirHash/:fileHash/star", unstarFile)

	////group.POST("share/*dirPath", shareFiles)
	group.POST("chunks", uploadFileChunk)
	group.POST("chunks/:fileHash", mergeFileChunks)
	group.GET("chunks/:fileHash", getMissedChunks)

	group.GET("trash", getTrashFiles)
	group.DELETE("trash/:dirHash/:fileHash", deleteTrashFile)
	group.DELETE("trash/:dirHash", deleteTrashDir)
	group.DELETE("trash", clearTrashFiles)
	group.POST("/:dirHash/:fileHash/untrash", restoreTrashFile)
	group.POST("/:dirHash/untrash", restoreTrashDir)
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
	//// check file size
	//if fileInfo.FileSize > MaxUploadSize {
	//	c.JSON(400, gin.H{"message": fmt.Sprintf("Uploaded file %s is too big", file.Filename)})
	//	return
	//}
	// store file content if not exists
	fileStoragePath := filepath.Join(configs.Local.StoragePath, fileInfo.FileHash)
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
	session := sessions.Default(c)
	userID := session.Get("userID")
	// store file metadata to database regardless of file existence
	err = model.StoreFileMetadata(&fileInfo, fileStoragePath, exists, userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to store file metadata", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"file": fileInfo})
}

// get metadata of all files under given directory
func getFilesMetadata(c *gin.Context) {
	dirHash := c.Param("dirHash")
	star := c.Query("star")
	sort := c.Query("sort")
	order := c.Query("order")
	isStarred := false
	if star == "true" {
		isStarred = true
	}
	// get user info
	session := sessions.Default(c)
	userID := session.Get("userID")
	log.WithFields(logrus.Fields{
		"dirHash": dirHash,
		"userID":  userID,
	}).Info("trying to get file metadata")
	// get metadata of all the files under the directory
	files, dirs, err := model.GetFilesMetadata(dirHash, isStarred, sort, order)
	if err != nil {
		c.JSON(500, gin.H{"message": fmt.Sprintf("failed to get files and dirs under dir %s", dirHash), "description": err.Error()})
		return
	}
	// construct files in response
	fileResponses := service.Convert2FileResponse(files, dirs)
	c.JSON(200, fileResponses)
}

// download the whole directory and return zipped result
func downloadDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	path := c.Query("path")
	tempFileStoragePath := configs.Local.TempStoragePath
	err := service.ArchiveDir(path, dirHash, filepath.Join(tempFileStoragePath, dirHash))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to archive directory", "description": err.Error()})
		return
	}
	log.WithFields(logrus.Fields{
		"hash":       dirHash,
		"path":       path,
		"zippedPath": filepath.Join(tempFileStoragePath, dirHash),
	}).Info("directory archived")
	file, err := os.Open(filepath.Join(tempFileStoragePath, dirHash))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
		return
	}
	defer func() { // delete the temporal zipped file
		err := os.Remove(filepath.Join(tempFileStoragePath, dirHash))
		if err != nil {
			log.WithFields(logrus.Fields{
				"filePath": filepath.Join(tempFileStoragePath, dirHash),
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
	action := c.Query("action")
	if fileHash == "" || fileName == "" || action == "" {
		c.JSON(400, gin.H{"message": "fileHash, fileName and action cannot be empty"})
		return
	}
	// get file metadata
	fileInfo, err := model.GetFileMetadata(fileHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get file metadata", "description": err.Error()})
	}
	// if size exceeds the threshold, we zip file and name the zipped file by file name.
	// Not for image, video and audio files since they have been archived to some extent
	//isArchived := false
	//log.Debug("file size: ", fileInfo.Size)
	//if fileInfo.Size > ArchiveThreshold &&
	//	!strings.HasPrefix(fileInfo.FileType, "image") && !strings.HasPrefix(fileInfo.FileType, "audio") &&
	//	!strings.HasPrefix(fileInfo.FileType, "video") {
	//	isArchived = true
	//	err = service.ArchiveFile(fileInfo.Location, fileName, filepath.Join(TempFileStoragePath, fileInfo.Hash))
	//	if err != nil {
	//		c.JSON(500, gin.H{"message": "failed to archive file", "description": err.Error()})
	//		return
	//	}
	//	log.WithFields(logrus.Fields{
	//		"fileHash":   fileInfo.Hash,
	//		"fileName":   fileName,
	//		"zippedPath": filepath.Join(TempFileStoragePath, fileInfo.Hash),
	//	}).Info("file archived")
	//}
	// write response header
	c.Header("Content-Type", fileInfo.FileType)
	if fileInfo.FileType == "text/plain" {
		c.Header("Content-Type", fmt.Sprintf("%s;charset=utf-8", fileInfo.FileType))
	}
	// return the file
	//if isArchived {
	//	file, err := os.Open(filepath.Join(TempFileStoragePath, fileInfo.Hash))
	//	if err != nil {
	//		c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
	//		return
	//	}
	//	defer func() { // delete the temporal zipped file
	//		err := os.Remove(filepath.Join(TempFileStoragePath, fileInfo.Hash))
	//		if err != nil {
	//			log.WithFields(logrus.Fields{
	//				"filePath": filepath.Join(TempFileStoragePath, fileInfo.Hash),
	//			}).Error("failed to remove temporal zipped file")
	//		}
	//	}()
	//	// download zip file and name it with extension `zip`
	//	zipName := strings.Split(fileName, ".")[0] + ".zip"
	//	log.Debug("zipName: ", zipName)
	//	c.Header("Content-Disposition", "attachment; filename="+zipName)
	//	c.Header("Content-Encoding", "zip")
	//	io.Copy(c.Writer, file)
	//} else {
	file, err := os.Open(fileInfo.Location)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to open file", "description": err.Error()})
		return
	}
	defer file.Close()
	if action == "download" {
		c.Header("Content-Disposition", "attachment; filename="+fileName)
	} else if action == "preview" {
		c.Header("Content-Disposition", "inline; filename="+fileName)
	}
	io.Copy(c.Writer, file)
	//}
}

func deleteDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	if err := model.DeleteDir(dirHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to delete directory", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func deleteFile(c *gin.Context) {
	fileHash := c.Param("fileHash")
	dirHash := c.Param("dirHash")
	if err := model.DeleteFile(dirHash, fileHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to delete file", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
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
	var tempFileDir = filepath.Join(configs.Local.TempStoragePath, chunkRequest.FileHash)
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
		log.Debugf("index: %d, chunk hash: %s", chunkRequest.Index, chunkRequest.ChunkHash)
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
	chunkDir := filepath.Join(configs.Local.TempStoragePath, fileHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to read chunk directory", "description": err.Error()})
		return
	}
	targetPath := filepath.Join(configs.Local.StoragePath, fileHash)
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
	for i := 0; uint(i) < chunkInfo.TotalChunks; i++ {
		_, ok := chunkInfo.Indexes[uint(i)]
		if !ok {
			c.JSON(400, gin.H{"message": "failed to merge chunks", "missedChunk": i})
			return
		}
	}
	// read file in order and write into target file
	for i := 0; uint(i) < chunkInfo.TotalChunks; i++ {
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
	err = model.StoreFileMetadata(&merge, targetPath, false, userID.(uint))
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
	chunkArray := make([]bool, chunkInfo.TotalChunks)
	for index := range chunkInfo.Indexes {
		chunkArray[index] = true
	}
	for i := 0; i < len(chunkArray); i++ {
		if !chunkArray[i] {
			missedChunks = append(missedChunks, uint(i))
		}
	}
	c.JSON(200, gin.H{"exists": true, "missedChunks": missedChunks})
}

func starDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	if err := model.StarDir(dirHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to star a directory", "description": err.Error()})
		return
	}
	c.JSON(200, dirHash)
}

func unstarDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	if err := model.UnstarDir(dirHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to unstar a directory", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func starFile(c *gin.Context) {
	dirHash := c.Param("dirHash")
	fileHash := c.Param("fileHash")
	if err := model.StarFile(dirHash, fileHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to star a file", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"dirHash": dirHash, "fileHash": fileHash})
}

func unstarFile(c *gin.Context) {
	dirHash := c.Param("dirHash")
	fileHash := c.Param("fileHash")
	if err := model.UnstarFile(dirHash, fileHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to unstar a file", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func getStarredFiles(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	files, dirs, err := model.GetStarredFiles(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get starred files", "description": err.Error()})
		return
	}
	fileResponses := service.Convert2FileResponse(files, dirs)
	c.JSON(200, fileResponses)
}

func getTrashFiles(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	files, dirs, err := model.GetTrashFiles(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get trash files", "description": err.Error()})
		return
	}
	fileResponses := service.Convert2FileResponse(files, dirs)
	c.JSON(200, fileResponses)
}

func deleteTrashFile(c *gin.Context) {
	dirHash := c.Param("dirHash")
	fileHash := c.Param("fileHash")
	if err := model.DeleteTrashFile(dirHash, fileHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to delete trash file", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func deleteTrashDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	if err := model.DeleteTrashDir(dirHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to delete trash directory", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func clearTrashFiles(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if err := model.ClearTrashFiles(userID.(uint)); err != nil {
		c.JSON(500, gin.H{"message": "failed to clear trash files", "description": err.Error()})
		return
	}
	c.Writer.WriteHeader(204)
}

func restoreTrashFile(c *gin.Context) {
	dirHash := c.Param("dirHash")
	fileHash := c.Param("fileHash")
	if err := model.RestoreTrashFile(dirHash, fileHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to restore trash file", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"dirHash": dirHash, "fileHash": fileHash})
}

func restoreTrashDir(c *gin.Context) {
	dirHash := c.Param("dirHash")
	if err := model.RestoreTrashDir(dirHash); err != nil {
		c.JSON(500, gin.H{"message": "failed to restore trash directory", "description": err.Error()})
		return
	}
	c.JSON(200, gin.H{"dirHash": dirHash})
}

func getTraceDirs(c *gin.Context) {
	dirHash := c.Param("dirHash")
	dirs, err := model.TracePathDirs(dirHash)
	if err != nil {
		c.JSON(500, gin.H{"message": "failed to get directories in the path", "description": err.Error()})
		return
	}
	c.JSON(200, dirs)
}
