package entity

type Folder struct {
	id           uint
	accountID    uint
	policyID     uint
	parentFolder *uint
	name         string
}

func (f *Folder) Id() uint {
	return f.id
}

func (f *Folder) AccountID() uint {
	return f.accountID
}

func (f *Folder) PolicyID() uint {
	return f.policyID
}

func (f *Folder) ParentFolder() *uint {
	return f.parentFolder
}

func (f *Folder) Name() string {
	return f.name
}

func NewFolder(userID uint, policyID uint, parentFolder *uint, name string) *Folder {
	return &Folder{
		accountID:    userID,
		policyID:     policyID,
		parentFolder: parentFolder,
		name:         name,
	}
}

// UnmarshallFolder 从仓储实体映射回来领域实体，因为本函数不做参数验证和参数转换
func UnmarshallFolder(id uint, userID uint, policyID uint, parentFolder *uint, name string) *Folder {
	return &Folder{
		id:           id,
		accountID:    userID,
		policyID:     policyID,
		parentFolder: parentFolder,
		name:         name,
	}
}
