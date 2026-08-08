package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	currentVersion = 1
	versionPrefix  = "v1."
)

var additionalData = []byte("fortyone:integration-credentials:v1")

// Box encrypts integration credentials with a versioned AES-GCM envelope.
// Values without a version prefix are treated as legacy plaintext so callers
// can rotate existing installations lazily after a deployment.
type Box struct {
	aead cipher.AEAD
	rand io.Reader
}

// OpenResult includes the plaintext and the envelope version. Version zero
// identifies a legacy plaintext value that should be re-encrypted on write.
type OpenResult struct {
	Plaintext []byte
	Version   int
}

func New(secret string) (*Box, error) {
	return newWithReader(secret, rand.Reader)
}

func newWithReader(secret string, random io.Reader) (*Box, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("secretbox: secret is required")
	}
	if random == nil {
		return nil, errors.New("secretbox: random source is required")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("secretbox: create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secretbox: create gcm: %w", err)
	}
	return &Box{aead: aead, rand: random}, nil
}

func (b *Box) Seal(plaintext []byte) (string, error) {
	if b == nil || b.aead == nil {
		return "", errors.New("secretbox: box is not configured")
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(b.rand, nonce); err != nil {
		return "", fmt.Errorf("secretbox: generate nonce: %w", err)
	}
	ciphertext := b.aead.Seal(nil, nonce, plaintext, additionalData)
	envelope := append(nonce, ciphertext...)
	return versionPrefix + base64.RawURLEncoding.EncodeToString(envelope), nil
}

func (b *Box) Open(value string) (OpenResult, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return OpenResult{}, errors.New("secretbox: encrypted value is empty")
	}
	if !strings.HasPrefix(value, "v") {
		return OpenResult{Plaintext: []byte(value), Version: 0}, nil
	}
	if !strings.HasPrefix(value, versionPrefix) {
		return OpenResult{}, errors.New("secretbox: unsupported envelope version")
	}
	if b == nil || b.aead == nil {
		return OpenResult{}, errors.New("secretbox: box is not configured")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(value, versionPrefix))
	if err != nil {
		return OpenResult{}, fmt.Errorf("secretbox: decode envelope: %w", err)
	}
	if len(raw) <= b.aead.NonceSize() {
		return OpenResult{}, errors.New("secretbox: encrypted value is too short")
	}
	nonce := raw[:b.aead.NonceSize()]
	ciphertext := raw[b.aead.NonceSize():]
	plaintext, err := b.aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return OpenResult{}, fmt.Errorf("secretbox: decrypt envelope: %w", err)
	}
	return OpenResult{Plaintext: plaintext, Version: currentVersion}, nil
}

func CurrentVersion() int {
	return currentVersion
}
