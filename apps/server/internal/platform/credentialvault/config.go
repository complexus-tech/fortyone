package credentialvault

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	DevelopmentKeyID      = "development"
	DevelopmentKeyVersion = uint32(1)
	// DevelopmentEncodedKeys is intentionally public and must never be accepted
	// by production configuration validation.
	DevelopmentEncodedKeys = `{"development@1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`
)

// ParseEncodedKeyring parses a JSON object whose keys are "<id>@<version>"
// references and whose values are base64-encoded 32-byte KEKs.
func ParseEncodedKeyring(activeID string, activeVersion uint32, encodedKeys string) (Config, error) {
	active := KeyRef{ID: strings.TrimSpace(activeID), Version: activeVersion}
	if err := validateKeyRef(active); err != nil {
		return Config{}, fmt.Errorf("credential vault active key: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(encodedKeys)))
	encoded, err := decodeEncodedKeyring(decoder)
	if err != nil {
		return Config{}, fmt.Errorf("credential vault keyring must be a JSON object: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Config{}, fmt.Errorf("credential vault keyring must contain exactly one JSON object: %w", err)
	}
	if len(encoded) == 0 {
		return Config{}, fmt.Errorf("credential vault keyring is empty")
	}
	keys := make([]Key, 0, len(encoded))
	for encodedRef, encodedMaterial := range encoded {
		ref, err := parseKeyRef(encodedRef)
		if err != nil {
			return Config{}, err
		}
		material, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedMaterial))
		if err != nil {
			return Config{}, fmt.Errorf("credential vault key %s@%d is not valid base64", ref.ID, ref.Version)
		}
		keys = append(keys, Key{Ref: ref, Material: material})
	}
	return Config{Active: active, Keys: keys}, nil
}

func decodeEncodedKeyring(decoder *json.Decoder) (map[string]string, error) {
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("expected an object")
	}
	encoded := make(map[string]string)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("expected a string key reference")
		}
		if _, duplicate := encoded[key]; duplicate {
			return nil, fmt.Errorf("duplicate key reference %q", key)
		}
		var material string
		if err := decoder.Decode(&material); err != nil {
			return nil, fmt.Errorf("key reference %q must map to a base64 string: %w", key, err)
		}
		encoded[key] = material
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := closing.(json.Delim); !ok || delimiter != '}' {
		return nil, errors.New("expected the end of an object")
	}
	return encoded, nil
}

func NewFromEncodedKeyring(activeID string, activeVersion uint32, encodedKeys string) (*Vault, error) {
	cfg, err := ParseEncodedKeyring(activeID, activeVersion, encodedKeys)
	if err != nil {
		return nil, err
	}
	defer func() {
		for index := range cfg.Keys {
			zero(cfg.Keys[index].Material)
		}
	}()
	return New(cfg)
}

// ContainsDevelopmentKey reports whether a keyring contains the public local
// development key or reserves the development key ID. Production startup uses
// this semantic check instead of comparing raw JSON, whose whitespace and
// member order are not security properties.
func ContainsDevelopmentKey(activeID string, activeVersion uint32, encodedKeys string) (bool, error) {
	cfg, err := ParseEncodedKeyring(activeID, activeVersion, encodedKeys)
	if err != nil {
		return false, err
	}
	defer func() {
		for index := range cfg.Keys {
			zero(cfg.Keys[index].Material)
		}
	}()
	for _, key := range cfg.Keys {
		if key.Ref.ID == DevelopmentKeyID || allZero(key.Material) {
			return true, nil
		}
	}
	return false, nil
}

// ReusesSecretMaterial reports whether a vault KEK is byte-for-byte equal to
// another application secret, supplied either as raw bytes or standard
// base64. It never includes either value in an error.
func ReusesSecretMaterial(activeID string, activeVersion uint32, encodedKeys, otherSecret string) (bool, error) {
	cfg, err := ParseEncodedKeyring(activeID, activeVersion, encodedKeys)
	if err != nil {
		return false, err
	}
	defer func() {
		for index := range cfg.Keys {
			zero(cfg.Keys[index].Material)
		}
	}()

	candidates := make([][]byte, 0, 3)
	raw := []byte(otherSecret)
	candidates = append(candidates, raw)
	trimmed := []byte(strings.TrimSpace(otherSecret))
	if string(trimmed) != otherSecret {
		candidates = append(candidates, trimmed)
	}
	encodedCandidate := strings.TrimSpace(otherSecret)
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, decodeErr := encoding.DecodeString(encodedCandidate); decodeErr == nil {
			candidates = append(candidates, decoded)
		}
	}
	defer func() {
		for _, candidate := range candidates {
			zero(candidate)
		}
	}()

	for _, key := range cfg.Keys {
		for _, candidate := range candidates {
			if len(candidate) == keySize && subtle.ConstantTimeCompare(key.Material, candidate) == 1 {
				return true, nil
			}
		}
	}
	return false, nil
}

func allZero(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}

func parseKeyRef(value string) (KeyRef, error) {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return KeyRef{}, fmt.Errorf("credential vault key reference %q must use <id>@<version>", value)
	}
	version, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || version == 0 {
		return KeyRef{}, fmt.Errorf("credential vault key reference %q has an invalid version", value)
	}
	ref := KeyRef{ID: parts[0], Version: uint32(version)}
	if err := validateKeyRef(ref); err != nil {
		return KeyRef{}, fmt.Errorf("credential vault key reference %q: %w", value, err)
	}
	return ref, nil
}
