package service

import (
	"CloudDrive/model"
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// user-defined function to deal with file when walkDir pass by it
type myWalkDirFunc func(fileInfo *model.File, err error) error

// descends file/dir at dirPath/fileName and calls walkDirFn when pass by each file
// note that when walkDir meets error, it lets walkDirFn deal with that.
func walkDir(dirPath string, fileName string, walkDirFn myWalkDirFunc) error {
	file, err := model.GetFileMetadata(dirPath, fileName) // info includes file type and location etc.
	if err != nil {
		return err
	}
	// call walkDirFn the first time for root file.
	// If there's an error during walkDirFn, or it is a single file, return
	if err := walkDirFn(file, nil); err != nil || file.FileType != "dir" {
		return err
	}
	// filetype is dir
	filesMetadata, err := model.GetFilesMetadata(filepath.Join(dirPath, fileName))
	if err != nil {
		// second call for root file, to report ReadDir error
		err = walkDirFn(file, err)
		return err
	}
	for _, file1 := range filesMetadata {
		err := walkDir(file1.DirPath, file1.Name, walkDirFn)
		if err != nil {
			return err
		}
	}
	return nil
}

// ArchiveFile archive single file or a directory in zip format
func ArchiveFile(dirPath string, fileName string, dstPath string) error {
	// create a zip file and zip.Writer
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()
	rootDir := dirPath
	// go through all the files of the srcPath
	walker := func(fileInfo *model.File, err error) error {
		if err != nil {
			return err
		}
		path := filepath.Join(fileInfo.DirPath, fileInfo.Name)
		log.Debugf("walk file %s", path)
		if fileInfo.FileType == "dir" { //  直接返回nil会忽略空目录，需要在这里创建一下目录再返回
			_, err := writer.Create(path[len(rootDir):])
			return err
		}
		file, err := os.Open(fileInfo.Location)
		if err != nil {
			return err
		}
		defer file.Close()

		// create path in zip should use zip root related path instead of absolute path,
		// otherwise it will create all the parent directory
		// TODO 可以获取srcPath的长度，然后取path[len(srcPath:] 就获得了相对路径
		zipFile, err := writer.Create(path[len(rootDir):])
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	}
	err = walkDir(dirPath, fileName, walker)
	return err
}
