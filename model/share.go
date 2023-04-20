package model

import (
	"time"
)

type Share struct {
	SharedID    string `gorm:"primaryKey"` // uuid for shared link
	FileHash    string
	OwnerID     uint    // the user that file belongs to
	UserID      *uint   // shared user id, empty if shared
	UserRole    uint    // 0 for editor, 1 for viewer
	Password    *string // password for accessing the file, can be nil
	AccessCount uint    // count user access times
	SharedTime  time.Time
	ExpiredTime *time.Time // nil when no time limit
	IsLimited   bool
	// 再加一个字段为 isLimited bool类型，当该字段true且userID为空，说明分享给限定用户，但是他还没注册，不允许访问；当注册过后，在更新UserID
	// 当该字段为false，userID应当为空， password可以为空也可以不为空
}

func CreateShare(share *Share) error {
	return db.Create(share).Error
}
