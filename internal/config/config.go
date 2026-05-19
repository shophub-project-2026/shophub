package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr string
	HTTPPort int
}

func Load() Config {
	port, err := strconv.Atoi(getEnv("SHOPHUB_HTTP_PORT", "8080"))
	if err != nil {
		port = 8080
	}
	return Config{
		HTTPAddr: getEnv("SHOPHUB_HTTP_ADDR", "0.0.0.0"),
		HTTPPort: port,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
