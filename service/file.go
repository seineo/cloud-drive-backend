package service

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ArchiveFile archive single file or a directory in zip format
func ArchiveFile(srcPath string, dstPath string) error {
	// create a zip file and zip.Writer
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()
	// go through all the files of the srcPath
	walker := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		log.Debug("path:", path)
		if d.IsDir() { // TODO 直接返回nil会忽略空目录，需要在这里创建一下目录再返回
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// create path in zip should use zip root related path instead of absolute path,
		// otherwise it will create all the parent directory
		// TODO 可以获取srcPath的长度，然后取path[len(srcPath:] 就获得了相对路径
		zipFile, err := writer.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	}
	err = filepath.WalkDir(srcPath, walker)
	return err
}
