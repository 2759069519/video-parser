package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Env      string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
}

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "development")

	if env == "production" {
		if password := os.Getenv("DB_PASSWORD"); password == "" {
			return nil, fmt.Errorf("DB_PASSWORD is required in production mode")
		}
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" || sslmode == "disable" {
			return nil, fmt.Errorf("DB_SSLMODE must be set to 'require' or higher in production mode")
		}
	}

	return &Config{
		Env: env,
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "video_parser"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
