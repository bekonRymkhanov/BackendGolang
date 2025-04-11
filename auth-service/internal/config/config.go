package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	// Service info
	Version     string `mapstructure:"VERSION"`
	Environment string `mapstructure:"ENVIRONMENT"`

	// Server
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`

	// CORS
	CORSAllowOrigins []string `mapstructure:"CORS_ALLOW_ORIGINS"`

	// Database
	DatabaseURL       string `mapstructure:"DATABASE_URL"`
	DBMaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBConnMaxLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME"`

	// JWT
	JWTSecret            string `mapstructure:"JWT_SECRET"`
	JWTRefreshSecret     string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTExpiryMinutes     int    `mapstructure:"JWT_EXPIRY_MINUTES"`
	JWTRefreshExpiryDays int    `mapstructure:"JWT_REFRESH_EXPIRY_DAYS"`

	// Default Admin
	CreateDefaultAdmin bool `mapstructure:"CREATE_DEFAULT_ADMIN"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Загружаем .env файл если он существует, а если нет то просто игнорируем ошибку
	_ = viper.ReadInConfig()

	viper.SetDefault("VERSION", "1.0.0")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER_ADDRESS", ":8080")
	viper.SetDefault("CORS_ALLOW_ORIGINS", "*")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/auth_service?sslmode=disable")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 300)
	viper.SetDefault("JWT_EXPIRY_MINUTES", 60)
	viper.SetDefault("JWT_REFRESH_EXPIRY_DAYS", 7)
	viper.SetDefault("CREATE_DEFAULT_ADMIN", false)

	// то что нужно сто процентов указать в .env файлах
	requiredEnvs := []string{"JWT_SECRET", "JWT_REFRESH_SECRET"}
	missingEnvs := []string{}

	for _, env := range requiredEnvs {
		if !viper.IsSet(env) {
			missingEnvs = append(missingEnvs, env)
		}
	}

	if len(missingEnvs) > 0 {
		return nil, fmt.Errorf("required environment variables not set: %s", strings.Join(missingEnvs, ", "))
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if corsStr := viper.GetString("CORS_ALLOW_ORIGINS"); corsStr != "" {
		config.CORSAllowOrigins = strings.Split(corsStr, ",")
	}

	return config, nil
}
