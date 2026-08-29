package developeroauth

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
)

const (
	DevelopmentDigestKeyID       = "development"
	DevelopmentEncodedDigestKeys = `{"development":"b2F1dGgtZGV2ZWxvcG1lbnQtZGlnZXN0LWtleS0wMDE="}`
	digestKeyBytes               = 32
)

var digestKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
var developmentDigestKeyMaterial = []byte("oauth-development-digest-key-001")

type DigestKey struct {
	Ref      developeroauthdomain.DigestKeyRef
	Material []byte
}

type TokenKeyringConfig struct {
	Active developeroauthdomain.DigestKeyRef
	Keys   []DigestKey
}

func ParseEncodedTokenKeyring(activeID string, encodedKeys string) (TokenKeyringConfig, error) {
	active := developeroauthdomain.DigestKeyRef{ID: strings.TrimSpace(activeID)}
	if err := validateDigestKeyRef(active); err != nil {
		return TokenKeyringConfig{}, fmt.Errorf("active OAuth digest key: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(encodedKeys)))
	var encoded map[string]string
	if err := decoder.Decode(&encoded); err != nil {
		return TokenKeyringConfig{}, fmt.Errorf("OAuth digest keyring must be a JSON object: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return TokenKeyringConfig{}, errors.New("OAuth digest keyring must contain exactly one JSON object")
		}
		return TokenKeyringConfig{}, fmt.Errorf("OAuth digest keyring has trailing input: %w", err)
	}
	if len(encoded) == 0 {
		return TokenKeyringConfig{}, errors.New("OAuth digest keyring is empty")
	}
	keys := make([]DigestKey, 0, len(encoded))
	activeFound := false
	for keyID, encodedMaterial := range encoded {
		ref := developeroauthdomain.DigestKeyRef{ID: strings.TrimSpace(keyID)}
		if err := validateDigestKeyRef(ref); err != nil {
			return TokenKeyringConfig{}, fmt.Errorf("OAuth digest key %q: %w", keyID, err)
		}
		material, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedMaterial))
		if err != nil || len(material) != digestKeyBytes {
			return TokenKeyringConfig{}, fmt.Errorf("OAuth digest key %s must be base64-encoded %d-byte material", keyID, digestKeyBytes)
		}
		for _, existing := range keys {
			if subtle.ConstantTimeCompare(existing.Material, material) == 1 {
				return TokenKeyringConfig{}, fmt.Errorf("OAuth digest keys %s and %s must use independent material", existing.Ref.ID, ref.ID)
			}
		}
		keys = append(keys, DigestKey{Ref: ref, Material: append([]byte(nil), material...)})
		activeFound = activeFound || ref == active
	}
	if !activeFound {
		return TokenKeyringConfig{}, fmt.Errorf("active OAuth digest key %s is missing", active.ID)
	}
	return TokenKeyringConfig{Active: active, Keys: keys}, nil
}

func ContainsDevelopmentDigestKey(config TokenKeyringConfig) bool {
	for _, key := range config.Keys {
		if key.Ref.ID == DevelopmentDigestKeyID ||
			subtle.ConstantTimeCompare(key.Material, developmentDigestKeyMaterial) == 1 ||
			allZeroBytes(key.Material) {
			return true
		}
	}
	return false
}

func DigestKeyringReusesSecret(config TokenKeyringConfig, secret string) bool {
	candidates := secretCandidates(secret)
	defer zeroByteSlices(candidates)
	for _, key := range config.Keys {
		for _, candidate := range candidates {
			if len(candidate) == len(key.Material) && subtle.ConstantTimeCompare(candidate, key.Material) == 1 {
				return true
			}
		}
	}
	return false
}

func validateDigestKeyRef(ref developeroauthdomain.DigestKeyRef) error {
	if !digestKeyIDPattern.MatchString(ref.ID) {
		return errors.New("key ID must contain 1-64 letters, digits, dots, underscores, or hyphens")
	}
	return nil
}

func secretCandidates(secret string) [][]byte {
	trimmed := strings.TrimSpace(secret)
	candidates := [][]byte{[]byte(secret)}
	if trimmed != secret {
		candidates = append(candidates, []byte(trimmed))
	}
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, err := encoding.DecodeString(trimmed); err == nil {
			candidates = append(candidates, decoded)
		}
	}
	return candidates
}

func zeroByteSlices(values [][]byte) {
	for _, value := range values {
		for index := range value {
			value[index] = 0
		}
	}
}

func allZeroBytes(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}
