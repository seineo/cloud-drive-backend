package config

type Config struct {
	ProjectName string
	ProjectURL  string
	Email       *EmailConfig
	MySQL       *MySQLConfig
	Redis       *RedisConfig
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		InitConfig()
	}
	return config
}

func InitConfig() {
	projectName := "Only云盘"
	projectURL := "localhost:4200"

	emailConfig := &EmailConfig{
		SMTPHost:      "smtp.qq.com",
		SMTPPort:      465,
		FromEmail:     "lyw_seineo@qq.com",
		FromEmailUser: "lyw_seineo@qq.com",
		FromEmailPsw:  "kvdqefpbeigmbbai",
	}

	mysqlConfig := &MySQLConfig{
		User:     "root",
		Password: "Li342204.",
		Protocol: "unix",
		Address:  "/tmp/mysql.sock",
		Database: "cloud_drive",
	}

	redisConfig := &RedisConfig{
		Network:        "tcp",
		Addr:           "localhost:6379",
		Password:       "",
		DB:             0,
		IdleConnection: 10,
		AuthKey:        "secret",
	}

	config = &Config{
		ProjectName: projectName,
		ProjectURL:  projectURL,
		Email:       emailConfig,
		MySQL:       mysqlConfig,
		Redis:       redisConfig,
	}
}
