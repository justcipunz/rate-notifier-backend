package config

import (
	"strings"
	"testing"
)

func TestLoadAPI_UsesFallbackDurations(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/rate_notifier?sslmode=disable")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("LoadAPI returned error: %v", err)
	}

	if cfg.JWTTTL.String() != "12h0m0s" {
		t.Fatalf("unexpected JWTTTL: %s", cfg.JWTTTL)
	}
}

func TestLoadAPI_ReturnsErrorOnInvalidDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/rate_notifier?sslmode=disable")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_TTL", "hello")

	_, err := LoadAPI()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_TTL") {
		t.Fatalf("expected JWT_TTL error, got %v", err)
	}
}

func TestLoadWorker_UsesFallbackDurations(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/rate_notifier?sslmode=disable")
	t.Setenv("RATE_PROVIDER_URL", "https://example.com/rates.json")

	cfg, err := LoadWorker()
	if err != nil {
		t.Fatalf("LoadWorker returned error: %v", err)
	}

	if cfg.RateFetchInterval.String() != "5m0s" {
		t.Fatalf("unexpected RateFetchInterval: %s", cfg.RateFetchInterval)
	}
	if cfg.RateProviderTimeout.String() != "10s" {
		t.Fatalf("unexpected RateProviderTimeout: %s", cfg.RateProviderTimeout)
	}
}

func TestLoadWorker_ReturnsErrorOnZeroDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/rate_notifier?sslmode=disable")
	t.Setenv("RATE_PROVIDER_URL", "https://example.com/rates.json")
	t.Setenv("RATE_FETCH_INTERVAL", "0s")

	_, err := LoadWorker()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "RATE_FETCH_INTERVAL") {
		t.Fatalf("expected RATE_FETCH_INTERVAL error, got %v", err)
	}
}

func TestLoadWorker_ReturnsErrorOnMissingRequiredValue(t *testing.T) {
	t.Setenv("RATE_PROVIDER_URL", "https://example.com/rates.json")

	_, err := LoadWorker()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected DATABASE_URL error, got %v", err)
	}
}
