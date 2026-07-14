package auth

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const MaxPasswordBytes = 72

var (
	ErrPasswordTooShort = errors.New("Пароль должен содержать не менее 8 символов")
	ErrPasswordTooLong  = errors.New("Пароль должен содержать не более 72 байт")
)

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
	if utf8.RuneCountInString(password) < 8 {
		return ErrPasswordTooShort
	}
	if len([]byte(password)) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}

	return nil
}
