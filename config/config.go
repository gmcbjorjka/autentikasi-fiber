package config

import "os"

type Config struct {
	AppPort   string
	AppHost   string
	MySQLHost string
	MySQLPort string
	MySQLUser string
	MySQLPass string
	MySQLDB   string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		AppPort:   getEnv("APP_PORT", "3000"),
		AppHost:   getEnv("APP_HOST", "0.0.0.0"),
		MySQLHost: getEnv("MYSQL_HOST", "127.0.0.1"),
		MySQLPort: getEnv("MYSQL_PORT", "3306"),
		MySQLUser: getEnv("MYSQL_USER", "root"),
		MySQLPass: getEnv("MYSQL_PASS", ""),
		MySQLDB:   getEnv("MYSQL_DB", "test"),
		JWTSecret: getEnv("JWT_SECRET", "changeme"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
