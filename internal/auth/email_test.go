package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail("  User@example.com  ")
	if err != nil {
		t.Fatalf("NormalizeEmail returned error: %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("unexpected email: %s", got)
	}
}

func TestNormalizeEmailWithDisplayName(t *testing.T) {
	got, err := NormalizeEmail("Vladimir <User@example.com>")
	if err != nil {
		t.Fatalf("NormalizeEmail returned error: %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("unexpected email: %s", got)
	}
}

func TestNormalizeEmailRejectsInvalidValue(t *testing.T) {
	if _, err := NormalizeEmail("not-an-email"); err == nil {
		t.Fatal("expected error for invalid email")
	}
}
