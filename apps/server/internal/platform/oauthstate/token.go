// Package oauthstate provides the opaque token primitive used by provider
// authorization flows. Persistence and identity binding remain the
// responsibility of the provider-specific state store.
package oauthstate

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const TokenSize = 32

var (
	ErrInvalidToken = errors.New("invalid OAuth state token")
	ErrRandomSource = errors.New("OAuth state random source is not configured")
)

// Token is a canonical, URL-safe encoding of 256 bits of entropy. Its digest
// is suitable for persistence; callers must never persist String() itself.
type Token struct {
	raw     [TokenSize]byte
	encoded string
}

// New generates a fresh OAuth state token from a cryptographically secure
// source. Passing nil fails closed instead of silently falling back to a weak
// source.
func New(source io.Reader) (Token, error) {
	if source == nil {
		return Token{}, ErrRandomSource
	}
	var raw [TokenSize]byte
	if _, err := io.ReadFull(source, raw[:]); err != nil {
		return Token{}, fmt.Errorf("generate OAuth state token: %w", err)
	}
	return tokenFromRaw(raw), nil
}

// NewRandom generates a token using crypto/rand.Reader.
func NewRandom() (Token, error) {
	return New(rand.Reader)
}

// Parse accepts only the canonical unpadded base64url representation emitted
// by New. Rejecting alternate encodings keeps token identity unambiguous.
func Parse(value string) (Token, error) {
	value = strings.TrimSpace(value)
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) != TokenSize {
		return Token{}, ErrInvalidToken
	}
	var fixed [TokenSize]byte
	copy(fixed[:], raw)
	token := tokenFromRaw(fixed)
	if token.encoded != value {
		return Token{}, ErrInvalidToken
	}
	return token, nil
}

func tokenFromRaw(raw [TokenSize]byte) Token {
	return Token{
		raw:     raw,
		encoded: base64.RawURLEncoding.EncodeToString(raw[:]),
	}
}

func (t Token) String() string {
	return t.encoded
}

// Digest returns a copy of the SHA-256 digest of the decoded random bytes.
func (t Token) Digest() []byte {
	digest := sha256.Sum256(t.raw[:])
	return append([]byte(nil), digest[:]...)
}
