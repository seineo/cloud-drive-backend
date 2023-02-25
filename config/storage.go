package config

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
