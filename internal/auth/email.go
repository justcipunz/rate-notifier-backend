package auth

import (
	"fmt"
	"net/mail"
	"strings"
)

func NormalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("email is empty")
	}

	address, err := mail.ParseAddress(value)
	if err != nil {
		return "", err
	}

	email := strings.ToLower(strings.TrimSpace(address.Address))
	if email == "" {
		return "", fmt.Errorf("email is empty")
	}

	if len(email) > 254 {
		return "", fmt.Errorf("email is too long")
	}

	return email, nil
}
