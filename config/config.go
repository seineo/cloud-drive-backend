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
	projectURL := "localhost:4200"

	emailConfig := &EmailConfig{
		SMTPHost:      "smtp.qq.com",
		SMTPPort:      465,
		FromEmail:     "lyw_seineo@qq.com",
		FromEmailUser: "lyw_seineo@qq.com",
		FromEmailPsw:  "kvdqefpbeigmbbai",
	}

	authCodeExpiredTime := 15 * time.Minute

	config = &Config{
		ProjectName:         projectName,
		ProjectURL:          projectURL,
		AuthCodeExpiredTime: authCodeExpiredTime,
		Email:               emailConfig,
		Storage:             initStorageConfig(),
		Log:                 initLogger(),
	}
}
