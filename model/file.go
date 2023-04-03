package model

import "time"

type File struct {
	Hash        string // MD5 hash value of file content, as primary key
	Name        string
	User        string
	ContentType string // dir, pdf, img, video...  TODO 是否要使用https://github.com/h2non/filetype
	Size        string
	DirPath     string // virtual directory path shown for users
	Location    string // real file storage path
	CreateTime  time.Time
}

func StoreFileMetadata() {

}

func GetFiles() {

}

func GetFileLocation() {

}
