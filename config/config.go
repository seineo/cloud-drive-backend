package config

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Config struct {
	ProjectName         string
	ProjectURL          string
	AuthCodeExpiredTime time.Duration
	Email               *EmailConfig
	Storage             *StorageConfig
	Log                 *logrus.Logger
	MaxUploadSize       int64 // max size of uploaded file that we allow
	ArchiveThreshold    int64
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		initConfig()
	}
	return config
}

func initConfig() {
	projectName := "Only云盘"
	// it should be standard url with http/https, otherwise url in email will be webdoc format
	projectURL := "http://localhost:4200"

	emailConfig := &EmailConfig{
		SMTPHost:      "smtp.qq.com",
		SMTPPort:      465,
		FromEmail:     "lyw_seineo@qq.com",
		FromEmailUser: "lyw_seineo@qq.com",
		FromEmailPsw:  "kvdqefpbeigmbbai",
	}

	const AUTH_CODE_EXPIRED = 15 * time.Minute
	const MAX_UPLOAD_SIZE = 1 << 40   // 1GB
	const ARCHIVE_THRESHOLD = 1 << 20 // 128MB

	config = &Config{
		ProjectName:         projectName,
		ProjectURL:          projectURL,
		AuthCodeExpiredTime: AUTH_CODE_EXPIRED,
		Email:               emailConfig,
		Storage:             initStorageConfig(),
		Log:                 initLogger(),
		MaxUploadSize:       MAX_UPLOAD_SIZE,
		ArchiveThreshold:    ARCHIVE_THRESHOLD,
	}
}
