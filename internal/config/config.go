package config

import (
	"os"
	"strconv"

	"github.com/shophub-project-2026/shophub/internal/db"
)

type Config struct {
	HTTPAddr  string
	HTTPPort  int
	DB        db.Config
	JWTSecret string
}

func Load() Config {
	port, err := strconv.Atoi(getEnv("SHOPHUB_HTTP_PORT", "8080"))
	if err != nil {
		port = 8080
	}
	dbPort, err := strconv.Atoi(getEnv("SHOPHUB_DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}
	return Config{
		HTTPAddr:  getEnv("SHOPHUB_HTTP_ADDR", "0.0.0.0"),
		HTTPPort:  port,
		JWTSecret: getEnv("SHOPHUB_JWT_SECRET", "change-me-in-production"),
		DB: db.Config{
			Host:     getEnv("SHOPHUB_DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("SHOPHUB_DB_NAME", "shophub_db"),
			User:     getEnv("SHOPHUB_DB_USER", "shophub_user"),
			Password: getEnv("SHOPHUB_DB_PASSWORD", "shophub_password"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
