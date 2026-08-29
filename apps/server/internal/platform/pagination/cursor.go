package pagination

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	cursorVersion         = "v1"
	minimumCursorKeyBytes = 32
	maximumCursorBytes    = 8 << 10
	maximumPayloadBytes   = 4 << 10
)

var (
	ErrInvalidCursor            = errors.New("invalid cursor")
	ErrInvalidCursorKey         = errors.New("invalid cursor signing key")
	ErrUnsupportedCursorVersion = errors.New("unsupported cursor version")
	ErrCursorTooLarge           = errors.New("cursor is too large")
	ErrCursorSignature          = errors.New("invalid cursor signature")
)

var cursorKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,31}$`)

// This is a public cryptographic domain-separation label, not secret material.
const cursorKeyDerivationDomain = "fortyone.cursor.v1" // gitleaks:allow

// SigningKey identifies one cursor-signing generation. IDs are safe metadata;
// Secret must come from the managed secret store and must never be logged.
type SigningKey struct {
	ID     string
	Secret []byte
}

// DeriveSigningKey gives each cursor family independent HMAC material even
// when process configuration has a single root secret. The purpose is public
// metadata and must be a stable, non-empty protocol identifier.
func DeriveSigningKey(id string, rootSecret []byte, purpose string) (SigningKey, error) {
	if !cursorKeyIDPattern.MatchString(id) || len(rootSecret) == 0 || strings.TrimSpace(purpose) == "" || purpose != strings.TrimSpace(purpose) {
		return SigningKey{}, ErrInvalidCursorKey
	}
	digest := hmac.New(sha256.New, rootSecret)
	_, _ = digest.Write([]byte(cursorKeyDerivationDomain))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(purpose))
	return SigningKey{ID: id, Secret: digest.Sum(nil)}, nil
}

// CursorCodec encodes a domain-owned cursor payload. The active key signs new
// cursors; previous keys may verify cursors during bounded rotation overlap.
type CursorCodec[T any] struct {
	activeID string
	keys     map[string][]byte
}

func NewCursorCodec[T any](active SigningKey, previous ...SigningKey) (CursorCodec[T], error) {
	keys := make(map[string][]byte, len(previous)+1)
	for _, key := range append([]SigningKey{active}, previous...) {
		if !cursorKeyIDPattern.MatchString(key.ID) || len(key.Secret) < minimumCursorKeyBytes {
			return CursorCodec[T]{}, ErrInvalidCursorKey
		}
		if _, exists := keys[key.ID]; exists {
			return CursorCodec[T]{}, fmt.Errorf("%w: duplicate key id", ErrInvalidCursorKey)
		}
		keys[key.ID] = bytes.Clone(key.Secret)
	}

	return CursorCodec[T]{activeID: active.ID, keys: keys}, nil
}

func (c CursorCodec[T]) Encode(payload T) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode cursor payload: %w", err)
	}
	if len(payloadJSON) > maximumPayloadBytes {
		return "", ErrCursorTooLarge
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signed := strings.Join([]string{cursorVersion, c.activeID, payloadPart}, ".")
	secret, ok := c.keys[c.activeID]
	if !ok {
		return "", ErrInvalidCursorKey
	}
	signature := signCursor(secret, signed)
	token := signed + "." + base64.RawURLEncoding.EncodeToString(signature)
	if len(token) > maximumCursorBytes {
		return "", ErrCursorTooLarge
	}
	return token, nil
}

func (c CursorCodec[T]) Decode(token string) (T, error) {
	var zero T
	if token == "" || len(token) > maximumCursorBytes {
		if len(token) > maximumCursorBytes {
			return zero, ErrCursorTooLarge
		}
		return zero, ErrInvalidCursor
	}

	parts := strings.Split(token, ".")
	if len(parts) != 4 {
		return zero, ErrInvalidCursor
	}
	if parts[0] != cursorVersion {
		return zero, ErrUnsupportedCursorVersion
	}
	secret, ok := c.keys[parts[1]]
	if !ok {
		return zero, ErrCursorSignature
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil || len(signature) != sha256.Size {
		return zero, ErrCursorSignature
	}
	signed := strings.Join(parts[:3], ".")
	if !hmac.Equal(signature, signCursor(secret, signed)) {
		return zero, ErrCursorSignature
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return zero, ErrInvalidCursor
	}
	if len(payloadJSON) > maximumPayloadBytes {
		return zero, ErrCursorTooLarge
	}

	decoder := json.NewDecoder(bytes.NewReader(payloadJSON))
	decoder.DisallowUnknownFields()
	var payload T
	if err := decoder.Decode(&payload); err != nil {
		return zero, ErrInvalidCursor
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return zero, ErrInvalidCursor
	}
	return payload, nil
}

func signCursor(secret []byte, value string) []byte {
	digest := hmac.New(sha256.New, secret)
	_, _ = digest.Write([]byte(value))
	return digest.Sum(nil)
}
