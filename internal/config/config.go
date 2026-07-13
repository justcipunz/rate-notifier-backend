package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppPort             string
	DatabaseURL         string
	JWTSecret           string
	JWTTTL              time.Duration
	RateProviderURL     string
	RateFetchInterval   time.Duration
	RateProviderTimeout time.Duration
	CORSAllowedOrigin   string
}

func Load() (Config, error) {
	cfg := Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(os.Getenv("JWT_SECRET")),
		RateProviderURL:   strings.TrimSpace(os.Getenv("RATE_PROVIDER_URL")),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
	}

	cfg.JWTTTL = parseDuration("JWT_TTL", 12*time.Hour)
	cfg.RateFetchInterval = parseDuration("RATE_FETCH_INTERVAL", 5*time.Minute)
	cfg.RateProviderTimeout = parseDuration("RATE_PROVIDER_TIMEOUT", 10*time.Second)

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if cfg.RateProviderURL == "" {
		missing = append(missing, "RATE_PROVIDER_URL")
	}
	if _, ok := os.LookupEnv("RATE_FETCH_INTERVAL"); !ok {
		missing = append(missing, "RATE_FETCH_INTERVAL")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
