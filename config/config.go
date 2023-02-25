package config

type Config struct {
	Email *EmailConfig
	MySQL *MySQLConfig
	Redis *RedisConfig
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		InitConfig()
	}
	return config
}

func InitConfig() {
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
		Network:        "",
		Addr:           "localhost:6379",
		Password:       "",
		DB:             0,
		IdleConnection: 10,
		AuthKey:        "secret",
	}

	config = &Config{
		Email: emailConfig,
		MySQL: mysqlConfig,
		Redis: redisConfig,
	}
}
