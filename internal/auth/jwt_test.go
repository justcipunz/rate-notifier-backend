package auth

import (
	"testing"
	"time"
)

func TestTokenManagerGenerateAndParse(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)

	token, err := manager.Generate(42, "user@example.com")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("unexpected user id: %d", claims.UserID)
	}

	if claims.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", claims.Email)
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)

	if _, err := manager.Parse("invalid-token"); err == nil {
		t.Fatal("Parse succeeded for invalid token")
	}
}
