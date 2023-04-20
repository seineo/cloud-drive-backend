package config

type StorageConfig struct {
	Redis               *RedisConfig
	MySQL               *MySQLConfig
	DiskStoragePath     string
	DiskTempStoragePath string
}

type RedisConfig struct {
	Network        string
	Addr           string
	Password       string
	DB             int
	IdleConnection int
	AuthKey        string
}

type MySQLConfig struct {
	User     string
	Password string
	Protocol string
	Address  string
	Database string
}

func initStorageConfig() *StorageConfig {
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

	//const DiskStoragePath = "/var/OnlyCloudDrive/"
	const DiskStoragePath = "/Users/liyuewei/Desktop/files"
	const DiskTempStoragePath = "/Users/liyuewei/Desktop/tempFiles"

	return &StorageConfig{
		Redis:               redisConfig,
		MySQL:               mysqlConfig,
		DiskStoragePath:     DiskStoragePath,
		DiskTempStoragePath: DiskTempStoragePath,
	}
}