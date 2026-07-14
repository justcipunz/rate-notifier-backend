package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const MaxPasswordBytes = 72

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func ValidatePassword(password string) error {
	if len([]byte(password)) < 8 {
		return fmt.Errorf("password must be at least 8 bytes")
	}
	if len([]byte(password)) > MaxPasswordBytes {
		return fmt.Errorf("password must be at most 72 bytes")
	}

	return nil
}
