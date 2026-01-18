package validate

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
)

// Email validates and standardizes an email address.
// It returns the normalized email (lowercase) and an error if invalid.
func Email(email string) (string, error) {
	// Trim spaces
	email = strings.TrimSpace(email)

	// Convert to lowercase
	email = strings.ToLower(email)

	// Parse email using Go's mail package
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", ErrInvalidEmail
	}

	return addr.Address, nil
}
