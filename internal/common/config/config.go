package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	ProjectName   string `mapstructure:"PROJECT_NAME"`
	ProjectURL    string `mapstructure:"PROJECT_URL"`
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisIdleConn int    `mapstructure:"REDIS_IDLE_CONN"`
	RedisNetwork  string `mapstructure:"REDIS_NETWORK"`
	RedisKey      string `mapstructure:"REDIS_KEY"`
	DBAddr        string `mapstructure:"DB_ADDR"`
	DBUser        string `mapstructure:"DB_USER"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBProtocol    string `mapstructure:"DB_PROTOCOL"`
	DBDatabase    string `mapstructure:"DB_DATABASE"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // 读取环境变量，覆盖默认值
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	config := Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
