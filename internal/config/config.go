package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type APIConfig struct {
	AppPort           string
	DatabaseURL       string
	JWTSecret         string
	JWTTTL            time.Duration
	CORSAllowedOrigin string
}

type WorkerConfig struct {
	DatabaseURL         string
	RateProviderURL     string
	RateFetchInterval   time.Duration
	RateProviderTimeout time.Duration
}

func LoadAPI() (APIConfig, error) {
	cfg := APIConfig{
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(os.Getenv("JWT_SECRET")),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
	}

	ttl, err := parsePositiveDuration("JWT_TTL", 12*time.Hour)
	if err != nil {
		return APIConfig{}, err
	}
	cfg.JWTTTL = ttl

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return APIConfig{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func LoadWorker() (WorkerConfig, error) {
	cfg := WorkerConfig{
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RateProviderURL: strings.TrimSpace(os.Getenv("RATE_PROVIDER_URL")),
	}

	interval, err := parsePositiveDuration("RATE_FETCH_INTERVAL", 5*time.Minute)
	if err != nil {
		return WorkerConfig{}, err
	}
	cfg.RateFetchInterval = interval

	timeout, err := parsePositiveDuration("RATE_PROVIDER_TIMEOUT", 10*time.Second)
	if err != nil {
		return WorkerConfig{}, err
	}
	cfg.RateProviderTimeout = timeout

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.RateProviderURL == "" {
		missing = append(missing, "RATE_PROVIDER_URL")
	}

	if len(missing) > 0 {
		return WorkerConfig{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func parsePositiveDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %s: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("duration %s must be greater than zero", key)
	}

	return parsed, nil
}
