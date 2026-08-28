package invitations

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	invitationTokenPrefix            = "wi1"
	invitationTokenVersion     int16 = 1
	invitationTokenRandomBytes       = 32
)

var invitationTokenKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

// InvitationTokenKey identifies one invitation-token HMAC key generation.
// Current is used for issuance; previous keys remain available only while
// outstanding invitations created under those generations can still expire.
type InvitationTokenKey struct {
	ID     string
	Secret string
}

// InvitationTokenConfig configures a versioned invitation-token keyring.
type InvitationTokenConfig struct {
	Current  InvitationTokenKey
	Previous []InvitationTokenKey
}

type invitationTokenKey struct {
	id     string
	secret []byte
}

// InvitationTokenManager owns token format validation, CSPRNG issuance, and
// HMAC lookup derivation. Raw bearer values never cross into repository input.
type InvitationTokenManager struct {
	current invitationTokenKey
	keys    map[string]invitationTokenKey
	random  io.Reader
}

// NewInvitationTokenManager validates and constructs a token keyring.
func NewInvitationTokenManager(config InvitationTokenConfig) (*InvitationTokenManager, error) {
	return newInvitationTokenManager(config, rand.Reader)
}

func newInvitationTokenManager(config InvitationTokenConfig, random io.Reader) (*InvitationTokenManager, error) {
	if random == nil {
		return nil, errors.New("invitation token random source is required")
	}

	configuredKeys := append([]InvitationTokenKey{config.Current}, config.Previous...)
	keys := make(map[string]invitationTokenKey, len(configuredKeys))
	for index, configuredKey := range configuredKeys {
		key, err := validateInvitationTokenKey(configuredKey)
		if err != nil {
			if index == 0 {
				return nil, fmt.Errorf("validate current invitation token key: %w", err)
			}
			return nil, fmt.Errorf("validate previous invitation token key: %w", err)
		}
		if _, exists := keys[key.id]; exists {
			return nil, fmt.Errorf("invitation token key ID %q is duplicated", key.id)
		}
		keys[key.id] = key
	}

	current := keys[strings.TrimSpace(config.Current.ID)]
	return &InvitationTokenManager{current: current, keys: keys, random: random}, nil
}

func validateInvitationTokenKey(config InvitationTokenKey) (invitationTokenKey, error) {
	keyID := strings.TrimSpace(config.ID)
	if !invitationTokenKeyIDPattern.MatchString(keyID) {
		return invitationTokenKey{}, errors.New("key ID must contain 1-64 letters, numbers, underscores, or hyphens")
	}

	secret := strings.TrimSpace(config.Secret)
	if len([]byte(secret)) < 32 {
		return invitationTokenKey{}, errors.New("HMAC key must contain at least 32 bytes")
	}

	return invitationTokenKey{id: keyID, secret: []byte(secret)}, nil
}

// Issue returns a one-time raw bearer for the email boundary and its
// independently safe persistence representation.
func (m *InvitationTokenManager) Issue() (string, StoredInvitationToken, error) {
	if m == nil || len(m.current.secret) == 0 {
		return "", StoredInvitationToken{}, errors.New("invitation token manager is not configured")
	}

	nonce := make([]byte, invitationTokenRandomBytes)
	if _, err := io.ReadFull(m.random, nonce); err != nil {
		return "", StoredInvitationToken{}, fmt.Errorf("generate invitation token: %w", err)
	}

	unsignedToken := strings.Join([]string{
		invitationTokenPrefix,
		m.current.id,
		base64.RawURLEncoding.EncodeToString(nonce),
	}, ".")
	signature := invitationTokenSignature(m.current, unsignedToken)
	rawToken := unsignedToken + "." + base64.RawURLEncoding.EncodeToString(signature)

	return rawToken, StoredInvitationToken{
		Digest:  invitationTokenDigest(m.current, rawToken),
		Nonce:   nonce,
		KeyID:   m.current.id,
		Version: invitationTokenVersion,
	}, nil
}

// Lookup validates a presented token before deriving a fixed-size database
// lookup. Malformed versioned values never fall back to plaintext comparison.
func (m *InvitationTokenManager) Lookup(rawToken string) (InvitationTokenLookup, error) {
	if m == nil {
		return InvitationTokenLookup{}, ErrInvalidToken
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return InvitationTokenLookup{}, ErrInvalidToken
	}

	if strings.Contains(rawToken, ".") {
		parts := strings.Split(rawToken, ".")
		if len(parts) != 4 || parts[0] != invitationTokenPrefix {
			return InvitationTokenLookup{}, ErrInvalidToken
		}

		key, exists := m.keys[parts[1]]
		if !exists {
			return InvitationTokenLookup{}, ErrInvalidToken
		}
		nonce, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil || len(nonce) != invitationTokenRandomBytes ||
			base64.RawURLEncoding.EncodeToString(nonce) != parts[2] {
			return InvitationTokenLookup{}, ErrInvalidToken
		}
		signature, err := base64.RawURLEncoding.DecodeString(parts[3])
		if err != nil || len(signature) != sha256.Size ||
			base64.RawURLEncoding.EncodeToString(signature) != parts[3] {
			return InvitationTokenLookup{}, ErrInvalidToken
		}
		unsignedToken := strings.Join(parts[:3], ".")
		if !hmac.Equal(signature, invitationTokenSignature(key, unsignedToken)) {
			return InvitationTokenLookup{}, ErrInvalidToken
		}

		return InvitationTokenLookup{
			Digest:  invitationTokenDigest(key, rawToken),
			KeyID:   key.id,
			Version: invitationTokenVersion,
		}, nil
	}

	// Before migration 155, the service issued padded URL-safe base64 tokens
	// containing exactly 32 random bytes. Accept only that historical shape.
	legacyNonce, err := base64.URLEncoding.DecodeString(rawToken)
	if err != nil || len(rawToken) != base64.URLEncoding.EncodedLen(invitationTokenRandomBytes) || len(legacyNonce) != invitationTokenRandomBytes {
		return InvitationTokenLookup{}, ErrInvalidToken
	}
	return InvitationTokenLookup{LegacyToken: rawToken}, nil
}

// Restore reconstructs a signed bearer only at the asynchronous email
// delivery boundary. The persisted digest is verified before returning it so
// corrupt metadata cannot produce a different credential.
func (m *InvitationTokenManager) Restore(stored StoredInvitationToken) (string, error) {
	if m == nil || len(stored.Nonce) != invitationTokenRandomBytes || len(stored.Digest) != sha256.Size || stored.Version != invitationTokenVersion {
		return "", ErrInvalidToken
	}
	key, exists := m.keys[stored.KeyID]
	if !exists {
		return "", ErrInvalidToken
	}

	unsignedToken := strings.Join([]string{
		invitationTokenPrefix,
		key.id,
		base64.RawURLEncoding.EncodeToString(stored.Nonce),
	}, ".")
	rawToken := unsignedToken + "." + base64.RawURLEncoding.EncodeToString(invitationTokenSignature(key, unsignedToken))
	if !hmac.Equal(stored.Digest, invitationTokenDigest(key, rawToken)) {
		return "", ErrInvalidToken
	}
	return rawToken, nil
}

func invitationTokenSignature(key invitationTokenKey, unsignedToken string) []byte {
	mac := hmac.New(sha256.New, key.secret)
	_, _ = mac.Write([]byte("fortyone:workspace-invitation-signature:v1\x00"))
	_, _ = mac.Write([]byte(unsignedToken))
	return mac.Sum(nil)
}

func invitationTokenDigest(key invitationTokenKey, rawToken string) []byte {
	mac := hmac.New(sha256.New, key.secret)
	_, _ = mac.Write([]byte("fortyone:workspace-invitation:v1\x00"))
	_, _ = mac.Write([]byte(rawToken))
	return mac.Sum(nil)
}
