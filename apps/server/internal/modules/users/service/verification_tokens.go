package users

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/validate"
)

const (
	verificationTokenVersion          int16 = 1
	verificationCodeUpperBound              = 1_000_000
	verificationTokenIssueLimit             = 3
	verificationTokenIssueWindow            = time.Hour
	verificationTokenMaximumTTL             = 15 * time.Minute
	verificationTokenCollisionRetries       = 5
)

var verificationTokenKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// VerificationTokenKey identifies one HMAC key generation. Current is used
// for new codes; Previous keys may remain enabled briefly while outstanding
// codes expire during rotation.
type VerificationTokenKey struct {
	ID     string
	Secret string
}

// VerificationTokenConfig configures a versioned verification-token keyring.
type VerificationTokenConfig struct {
	Current  VerificationTokenKey
	Previous []VerificationTokenKey
}

type verificationTokenKey struct {
	id     string
	secret []byte
}

// VerificationTokenManager generates CSPRNG codes and purpose-bound HMAC
// digests. It is deliberately separate from browser-session signing.
type VerificationTokenManager struct {
	current verificationTokenKey
	keys    []verificationTokenKey
	random  io.Reader
	now     func() time.Time
}

// NewVerificationTokenManager validates and constructs a token keyring.
func NewVerificationTokenManager(config VerificationTokenConfig) (*VerificationTokenManager, error) {
	return newVerificationTokenManager(config, rand.Reader, time.Now)
}

func newVerificationTokenManager(
	config VerificationTokenConfig,
	random io.Reader,
	now func() time.Time,
) (*VerificationTokenManager, error) {
	if random == nil {
		return nil, errors.New("verification token random source is required")
	}
	if now == nil {
		return nil, errors.New("verification token clock is required")
	}

	configuredKeys := append([]VerificationTokenKey{config.Current}, config.Previous...)
	keys := make([]verificationTokenKey, 0, len(configuredKeys))
	seenIDs := make(map[string]struct{}, len(configuredKeys))
	for index, configuredKey := range configuredKeys {
		key, err := validateVerificationTokenKey(configuredKey)
		if err != nil {
			if index == 0 {
				return nil, fmt.Errorf("validate current verification token key: %w", err)
			}
			return nil, fmt.Errorf("validate previous verification token key: %w", err)
		}
		if _, exists := seenIDs[key.id]; exists {
			return nil, fmt.Errorf("verification token key ID %q is duplicated", key.id)
		}
		seenIDs[key.id] = struct{}{}
		keys = append(keys, key)
	}

	return &VerificationTokenManager{
		current: keys[0],
		keys:    keys,
		random:  random,
		now:     now,
	}, nil
}

func validateVerificationTokenKey(config VerificationTokenKey) (verificationTokenKey, error) {
	keyID := strings.TrimSpace(config.ID)
	if !verificationTokenKeyIDPattern.MatchString(keyID) {
		return verificationTokenKey{}, errors.New("key ID must contain 1-64 letters, numbers, dots, underscores, or hyphens")
	}

	secret := strings.TrimSpace(config.Secret)
	if len([]byte(secret)) < 32 {
		return verificationTokenKey{}, errors.New("HMAC key must contain at least 32 bytes")
	}

	return verificationTokenKey{id: keyID, secret: []byte(secret)}, nil
}

func (m *VerificationTokenManager) issue(
	email string,
	tokenType string,
	expiresAt time.Time,
) (string, NewVerificationToken, error) {
	normalizedEmail, err := validate.Email(email)
	if err != nil {
		return "", NewVerificationToken{}, err
	}
	normalizedType, err := normalizeVerificationTokenType(tokenType)
	if err != nil {
		return "", NewVerificationToken{}, err
	}

	issuedAt := m.now().UTC()
	expiresAt = expiresAt.UTC()
	if !expiresAt.After(issuedAt) {
		return "", NewVerificationToken{}, errors.New("verification token expiry must be in the future")
	}
	if expiresAt.After(issuedAt.Add(verificationTokenMaximumTTL)) {
		return "", NewVerificationToken{}, fmt.Errorf("verification token expiry must not exceed %s", verificationTokenMaximumTTL)
	}

	value, err := rand.Int(m.random, big.NewInt(verificationCodeUpperBound))
	if err != nil {
		return "", NewVerificationToken{}, fmt.Errorf("generate verification code: %w", err)
	}
	code := fmt.Sprintf("%06d", value.Int64())

	return code, NewVerificationToken{
		Email:          normalizedEmail,
		TokenType:      normalizedType,
		TokenDigest:    verificationTokenDigest(m.current, normalizedEmail, normalizedType, code),
		TokenKeyID:     m.current.id,
		TokenVersion:   verificationTokenVersion,
		ExpiresAt:      expiresAt,
		IssuedAt:       issuedAt,
		RateLimitSince: issuedAt.Add(-verificationTokenIssueWindow),
		MaximumIssues:  verificationTokenIssueLimit,
	}, nil
}

func (m *VerificationTokenManager) consumption(
	email string,
	code string,
	tokenTypes ...string,
) (ConsumeVerificationTokenInput, error) {
	normalizedEmail, err := validate.Email(email)
	if err != nil {
		return ConsumeVerificationTokenInput{}, err
	}
	if !isVerificationCode(code) {
		return ConsumeVerificationTokenInput{}, ErrInvalidToken
	}
	if len(tokenTypes) == 0 {
		return ConsumeVerificationTokenInput{}, errors.New("at least one verification token purpose is required")
	}

	normalizedTypes := make([]string, 0, len(tokenTypes))
	for _, tokenType := range tokenTypes {
		normalizedType, err := normalizeVerificationTokenType(tokenType)
		if err != nil {
			return ConsumeVerificationTokenInput{}, err
		}
		if !slices.Contains(normalizedTypes, normalizedType) {
			normalizedTypes = append(normalizedTypes, normalizedType)
		}
	}

	candidateCount := len(m.keys) * len(normalizedTypes)
	digests := make([][]byte, 0, candidateCount)
	keyIDs := make([]string, 0, candidateCount)
	versions := make([]int16, 0, candidateCount)
	for _, key := range m.keys {
		for _, tokenType := range normalizedTypes {
			digests = append(digests, verificationTokenDigest(key, normalizedEmail, tokenType, code))
			keyIDs = append(keyIDs, key.id)
			versions = append(versions, verificationTokenVersion)
		}
	}

	return ConsumeVerificationTokenInput{
		Email:         normalizedEmail,
		TokenTypes:    normalizedTypes,
		TokenDigests:  digests,
		TokenKeyIDs:   keyIDs,
		TokenVersions: versions,
		LegacyToken:   code,
		ConsumedAt:    m.now().UTC(),
	}, nil
}

// RateLimitKey derives an opaque, fixed-size cache key. Raw email addresses,
// codes, and network identifiers never appear in Redis keys or logs.
func (m *VerificationTokenManager) RateLimitKey(scope, identityType, identity string) (string, error) {
	scope = strings.TrimSpace(scope)
	identityType = strings.TrimSpace(identityType)
	identity = strings.TrimSpace(identity)
	if scope == "" || identityType == "" || identity == "" {
		return "", errors.New("verification rate limit scope and identity are required")
	}

	mac := hmac.New(sha256.New, m.current.secret)
	_, _ = mac.Write([]byte("fortyone:verification-rate-limit:v1\x00"))
	_, _ = mac.Write([]byte(scope))
	_, _ = mac.Write([]byte("\x00"))
	_, _ = mac.Write([]byte(identityType))
	_, _ = mac.Write([]byte("\x00"))
	_, _ = mac.Write([]byte(identity))

	return fmt.Sprintf(
		"auth-verification:v1:%s:%s:%s",
		m.current.id,
		scope,
		hex.EncodeToString(mac.Sum(nil)),
	), nil
}

func verificationTokenDigest(key verificationTokenKey, email, tokenType, code string) []byte {
	mac := hmac.New(sha256.New, key.secret)
	_, _ = mac.Write([]byte("fortyone:verification-token:v1\x00"))
	_, _ = mac.Write([]byte(email))
	_, _ = mac.Write([]byte("\x00"))
	_, _ = mac.Write([]byte(tokenType))
	_, _ = mac.Write([]byte("\x00"))
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil)
}

func normalizeVerificationTokenType(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case TokenTypeLogin, TokenTypeRegistration:
		return value, nil
	default:
		return "", fmt.Errorf("unsupported verification token purpose %q", value)
	}
}

func isVerificationCode(value string) bool {
	if len(value) != 6 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
