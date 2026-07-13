package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort string
}

func Load() Config {
	cfg := Config{
		AppPort: getEnv("APP_PORT", "8080"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
