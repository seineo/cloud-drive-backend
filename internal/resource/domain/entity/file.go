package entity

type File struct {
	id       uint
	folderID uint
	name     string
	size     uint
}

func NewFile(folderID uint, name string, size uint) *File {
	return &File{
		folderID: folderID,
		name:     name,
		size:     size,
	}
}
