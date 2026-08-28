package developercredentials

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
)

const (
	DevelopmentDigestKeyID       = "development"
	DevelopmentDigestKeyVersion  = uint32(1)
	DevelopmentEncodedDigestKeys = `{"development@1":"ZGV2ZWxvcGVyLWNyZWRlbnRpYWwtZGV2LWtleS0wMDE="}`
	digestKeyBytes               = 32
)

var digestKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
var developmentDigestKeyMaterial = []byte("developer-credential-dev-key-001")

type DigestKey struct {
	Ref      developercredentialsdomain.DigestKeyRef
	Material []byte
}

type TokenKeyringConfig struct {
	Active developercredentialsdomain.DigestKeyRef
	Keys   []DigestKey
}

func ParseEncodedTokenKeyring(activeID string, activeVersion uint32, encodedKeys string) (TokenKeyringConfig, error) {
	active := developercredentialsdomain.DigestKeyRef{ID: strings.TrimSpace(activeID), Version: activeVersion}
	if err := validateDigestKeyRef(active); err != nil {
		return TokenKeyringConfig{}, fmt.Errorf("active API credential digest key: %w", err)
	}

	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(encodedKeys)))
	encoded, err := decodeTokenKeyMap(decoder)
	if err != nil {
		return TokenKeyringConfig{}, fmt.Errorf("API credential digest keyring must be a JSON object: %w", err)
	}
	if err := requireTokenKeyringEOF(decoder); err != nil {
		return TokenKeyringConfig{}, err
	}
	if len(encoded) == 0 {
		return TokenKeyringConfig{}, errors.New("API credential digest keyring is empty")
	}

	keys := make([]DigestKey, 0, len(encoded))
	activeFound := false
	for encodedRef, encodedMaterial := range encoded {
		ref, err := parseDigestKeyRef(encodedRef)
		if err != nil {
			return TokenKeyringConfig{}, err
		}
		material, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedMaterial))
		if err != nil || len(material) != digestKeyBytes {
			return TokenKeyringConfig{}, fmt.Errorf("API credential digest key %s must be base64-encoded %d-byte material", encodedRef, digestKeyBytes)
		}
		keys = append(keys, DigestKey{Ref: ref, Material: material})
		activeFound = activeFound || ref == active
	}
	if !activeFound {
		return TokenKeyringConfig{}, fmt.Errorf("active API credential digest key %s@%d is missing", active.ID, active.Version)
	}
	return TokenKeyringConfig{Active: active, Keys: keys}, nil
}

func ContainsDevelopmentDigestKey(config TokenKeyringConfig) bool {
	for _, key := range config.Keys {
		if key.Ref.ID == DevelopmentDigestKeyID || allZeroBytes(key.Material) ||
			subtle.ConstantTimeCompare(key.Material, developmentDigestKeyMaterial) == 1 {
			return true
		}
	}
	return false
}

func DigestKeyringReusesSecret(config TokenKeyringConfig, secret string) bool {
	candidates := secretCandidates(secret)
	defer zeroCandidates(candidates)
	for _, key := range config.Keys {
		for _, candidate := range candidates {
			if len(candidate) == len(key.Material) && subtle.ConstantTimeCompare(key.Material, candidate) == 1 {
				return true
			}
		}
	}
	return false
}

func validateDigestKeyRef(ref developercredentialsdomain.DigestKeyRef) error {
	if !digestKeyIDPattern.MatchString(ref.ID) {
		return errors.New("key ID must contain 1-64 letters, digits, dots, underscores, or hyphens")
	}
	if ref.Version == 0 {
		return errors.New("key version must be positive")
	}
	return nil
}

func parseDigestKeyRef(value string) (developercredentialsdomain.DigestKeyRef, error) {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return developercredentialsdomain.DigestKeyRef{}, fmt.Errorf("API credential digest key reference %q must use <id>@<version>", value)
	}
	version, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || version == 0 {
		return developercredentialsdomain.DigestKeyRef{}, fmt.Errorf("API credential digest key reference %q has an invalid version", value)
	}
	ref := developercredentialsdomain.DigestKeyRef{ID: parts[0], Version: uint32(version)}
	if err := validateDigestKeyRef(ref); err != nil {
		return developercredentialsdomain.DigestKeyRef{}, fmt.Errorf("API credential digest key reference %q: %w", value, err)
	}
	return ref, nil
}

func decodeTokenKeyMap(decoder *json.Decoder) (map[string]string, error) {
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("expected an object")
	}
	values := make(map[string]string)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("expected a string key reference")
		}
		if _, duplicate := values[key]; duplicate {
			return nil, fmt.Errorf("duplicate key reference %q", key)
		}
		var material string
		if err := decoder.Decode(&material); err != nil {
			return nil, fmt.Errorf("key reference %q must map to a base64 string: %w", key, err)
		}
		values[key] = material
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := closing.(json.Delim); !ok || delimiter != '}' {
		return nil, errors.New("expected the end of an object")
	}
	return values, nil
}

func requireTokenKeyringEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); errors.Is(err, io.EOF) {
		return nil
	} else if err != nil {
		return fmt.Errorf("API credential digest keyring has trailing input: %w", err)
	}
	return errors.New("API credential digest keyring must contain exactly one JSON object")
}

func secretCandidates(secret string) [][]byte {
	trimmed := strings.TrimSpace(secret)
	candidates := [][]byte{[]byte(secret)}
	if trimmed != secret {
		candidates = append(candidates, []byte(trimmed))
	}
	for _, encoding := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if decoded, err := encoding.DecodeString(trimmed); err == nil {
			candidates = append(candidates, decoded)
		}
	}
	return candidates
}

func zeroCandidates(candidates [][]byte) {
	for _, candidate := range candidates {
		for index := range candidate {
			candidate[index] = 0
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
