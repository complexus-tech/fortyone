package fortyone

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const (
	minimumIdempotencyKeyBytes = 16
	maximumIdempotencyKeyBytes = 255
)

// NewIdempotencyKey returns a cryptographically random, URL/header-safe key.
// Persist the returned value with the operation until its retry lifecycle is
// complete; generating a fresh key for a retry defeats idempotency.
func NewIdempotencyKey() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return hex.EncodeToString(random), nil
}

// ValidateIdempotencyKey applies the public API's byte-level header contract.
func ValidateIdempotencyKey(value string) error {
	if len(value) < minimumIdempotencyKeyBytes || len(value) > maximumIdempotencyKeyBytes {
		return errors.New("FortyOne idempotency key must contain 16 to 255 visible ASCII characters")
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return errors.New("FortyOne idempotency key must contain 16 to 255 visible ASCII characters")
		}
	}
	return nil
}
