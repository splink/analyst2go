package db

type DatabaseConfig struct {
	Host       string
	Name       string
	Port       string
	User       string
	Password   string
	DisableSSL bool
}
