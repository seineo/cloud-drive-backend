package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type ProjectConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtpHost"`
	SMTPPort int    `yaml:"smtpPort"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type FileConfig struct {
	StaleTime     time.Duration `yaml:"staleTime"`
	StaleTimeCron string        `yaml:"staleTimeCron"`
}

type LocalConfig struct {
	TempStoragePath string `yaml:"tempStoragePath"`
	StoragePath     string `yaml:"storagePath"`
}

type RedisConfig struct {
	Addr           string `yaml:"addr"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	IdleConnection int    `yaml:"idleConnection"`
	Network        string `yaml:"network"`
	AuthKey        string `yaml:"authKey"`
}

type DatabaseConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Database string `yaml:"database"`
}

type Config struct {
	Project  ProjectConfig  `yaml:"project"`
	Email    EmailConfig    `yaml:"email"`
	File     FileConfig     `yaml:"file"`
	Local    LocalConfig    `yaml:"local"`
	Redis    RedisConfig    `yaml:"redis"`
	Database DatabaseConfig `yaml:"database"`
}

func LoadConfig(paths []string) Config {
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	config := Config{}
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err.Error())
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("failed to load config: %v\n", err.Error())
	}
	return config
}
