package config

import (
	"os"
)

type Config struct {
	FriendServiceURL string
	Port             string
	DB               DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func LoadConfig() Config {
	return Config{
		FriendServiceURL: getEnv("FRIEND_SERVICE_URL", "http://localhost:8001"),
		Port:             getEnv("PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "159.223.84.254"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "maindb"),
			User:     getEnv("DB_USER", "dbadmin"),
			Password: getEnv("DB_PASSWORD", "cgroup123"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
