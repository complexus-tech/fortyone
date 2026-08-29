package fortyone

import (
	"strings"
	"testing"
)

func TestNewIdempotencyKeyIsRandomAndContractValid(t *testing.T) {
	t.Parallel()
	first, err := NewIdempotencyKey()
	if err != nil {
		t.Fatalf("NewIdempotencyKey() error = %v", err)
	}
	second, err := NewIdempotencyKey()
	if err != nil {
		t.Fatalf("NewIdempotencyKey() second error = %v", err)
	}
	if first == second || len(first) != 64 {
		t.Fatalf("generated keys are not independent 32-byte hex values")
	}
	if err := ValidateIdempotencyKey(first); err != nil {
		t.Fatalf("ValidateIdempotencyKey() error = %v", err)
	}
}

func TestValidateIdempotencyKeyRejectsInvalidHeaderValues(t *testing.T) {
	t.Parallel()
	for _, value := range []string{
		"too-short",
		strings.Repeat("a", maximumIdempotencyKeyBytes+1),
		strings.Repeat("a", minimumIdempotencyKeyBytes) + " ",
		strings.Repeat("a", minimumIdempotencyKeyBytes) + "\n",
		strings.Repeat("a", minimumIdempotencyKeyBytes) + "é",
	} {
		if err := ValidateIdempotencyKey(value); err == nil {
			t.Fatalf("ValidateIdempotencyKey(%q) error = nil", value)
		}
	}
}
