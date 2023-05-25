package service

import (
	"archive/zip"
	"io"
	"os"
)

//// user-defined function to deal with file when walkDir pass by it
//type myWalkDirFunc func(fileInfo *model.File, err error) error
//
//// descends file/dir at dirPath/fileName and calls walkDirFn when pass by each file
//// note that when walkDir meets error, it lets walkDirFn deal with that.
//func walkDir(userID uint, fileHash string, dirPath string, fileName string, walkDirFn myWalkDirFunc) error {
//	file, err := model.GetFileMetadata(fileHash) // info includes file type and location etc.
//	if err != nil {
//		return err
//	}
//	// call walkDirFn the first time for root file.
//	// If there's an error during walkDirFn, or it is a single file, return
//	if err := walkDirFn(file, nil); err != nil || file.FileType != "dir" {
//		return err
//	}
//	// filetype is dir
//	filesMetadata, err := model.GetFilesMetadata(userID, filepath.Join(dirPath, fileName))
//	if err != nil {
//		// second call for root file, to report ReadDir error
//		err = walkDirFn(file, err)
//		return err
//	}
//	for _, file1 := range filesMetadata {
//		err := walkDir(userID, file1.Hash, file1.DirPath, file1.Name, walkDirFn)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func ArchiveDir(fileInfo *model.File, dirHash string, dstPath string) error {
//	// create a zip file and zip.Writer
//	f, err := os.Create(dstPath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//
//	writer := zip.NewWriter(f)
//	defer writer.Close()
//	// go through all the files of the srcPath
//	walker := func(fileInfo *model.File, err error) error {
//		if err != nil {
//			return err
//		}
//		path := filepath.Join(dirPath, fileInfo.Name)
//		log.Debugf("walk file %s", path)
//		// get relative path
//		relPath, err := filepath.Rel(dirPath, path)
//		if err != nil {
//			log.WithError(err).Error("failed to get relative path")
//			return err
//		}
//
//		// create file header
//		header := &zip.FileHeader{
//			Name:   relPath,
//			Method: zip.Deflate,
//		}
//		if fileInfo.FileType == "dir" { //  直接返回nil会忽略空目录，需要在这里创建一下目录再返回
//			header.Name += "/"
//			header.SetMode(0755)
//			_, err = writer.CreateHeader(header)
//			return err
//		}
//		// file type is not directory
//
//		// write file header to zip
//		zipFile, err := writer.CreateHeader(header)
//		if err != nil {
//			log.WithError(err).Error("failed to write file header")
//			return err
//		}
//		// write file content to zip
//		file, err := os.Open(fileInfo.Location)
//		if err != nil {
//			return err
//		}
//		defer file.Close()
//		_, err = io.Copy(zipFile, file)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//	err = walkDir(userID, fileHash, dirPath, fileName, walker)
//	return err
//}

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
