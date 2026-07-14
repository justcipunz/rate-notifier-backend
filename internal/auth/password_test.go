package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword(hash, "secret123"); err != nil {
		t.Fatalf("CheckPassword returned error for valid password: %v", err)
	}

	if err := CheckPassword(hash, "wrong-password"); err == nil {
		t.Fatal("CheckPassword succeeded for invalid password")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword(strings.Repeat("a", 7)); err == nil {
		t.Fatal("expected error for 7 ASCII characters")
	}

	if err := ValidatePassword(strings.Repeat("a", 8)); err != nil {
		t.Fatalf("ValidatePassword returned error for 8 ASCII characters: %v", err)
	}

	if err := ValidatePassword(strings.Repeat("а", 4)); err == nil {
		t.Fatal("expected error for 4 Cyrillic characters")
	}

	if err := ValidatePassword(strings.Repeat("а", 8)); err != nil {
		t.Fatalf("ValidatePassword returned error for 8 Cyrillic characters: %v", err)
	}

	long := strings.Repeat("a", MaxPasswordBytes+1)
	if err := ValidatePassword(long); err == nil {
		t.Fatal("expected error for long password")
	}
}
