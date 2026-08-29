package developeroauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/google/uuid"
)

const (
	lookupPrefixBytes       = 6
	secretBytes             = 32
	encodedPrefixLen        = lookupPrefixBytes * 2
	encodedSecretLen        = 43
	authorizationCodeHeader = "f41_oac_v1_"
	refreshTokenHeader      = "f41_ort_v1_" // #nosec G101 -- Public token format prefix, not credential material.
	clientSecretHeader      = "f41_ocs_v1_" // #nosec G101 -- Public token format prefix, not credential material.
)

type TokenManager struct {
	active developeroauthdomain.DigestKeyRef
	keys   map[developeroauthdomain.DigestKeyRef][]byte
	random io.Reader
}

func NewTokenManager(config TokenKeyringConfig) (*TokenManager, error) {
	return newTokenManager(config, rand.Reader)
}

func newTokenManager(config TokenKeyringConfig, random io.Reader) (*TokenManager, error) {
	if random == nil {
		return nil, errors.New("OAuth token random source is required")
	}
	if err := validateDigestKeyRef(config.Active); err != nil {
		return nil, err
	}
	keys := make(map[developeroauthdomain.DigestKeyRef][]byte, len(config.Keys))
	for _, key := range config.Keys {
		if err := validateDigestKeyRef(key.Ref); err != nil {
			return nil, err
		}
		if len(key.Material) != digestKeyBytes {
			return nil, fmt.Errorf("OAuth digest key %s must contain %d bytes", key.Ref.ID, digestKeyBytes)
		}
		if _, duplicate := keys[key.Ref]; duplicate {
			return nil, fmt.Errorf("duplicate OAuth digest key %s", key.Ref.ID)
		}
		keys[key.Ref] = append([]byte(nil), key.Material...)
	}
	if _, found := keys[config.Active]; !found {
		return nil, errors.New("active OAuth digest key is unavailable")
	}
	return &TokenManager{active: config.Active, keys: keys, random: random}, nil
}

func (manager *TokenManager) Issue(kind developeroauthdomain.SecretKind, id uuid.UUID) (developeroauthdomain.IssuedSecret, error) {
	if manager == nil || id == uuid.Nil {
		return developeroauthdomain.IssuedSecret{}, errors.New("OAuth token manager and token ID are required")
	}
	prefixBytes := make([]byte, lookupPrefixBytes)
	secretMaterial := make([]byte, secretBytes)
	if _, err := io.ReadFull(manager.random, prefixBytes); err != nil {
		return developeroauthdomain.IssuedSecret{}, fmt.Errorf("generate OAuth lookup prefix: %w", err)
	}
	if _, err := io.ReadFull(manager.random, secretMaterial); err != nil {
		return developeroauthdomain.IssuedSecret{}, fmt.Errorf("generate OAuth secret: %w", err)
	}
	defer zeroByteSlices([][]byte{prefixBytes, secretMaterial})
	prefix := hex.EncodeToString(prefixBytes)
	secret := base64.RawURLEncoding.EncodeToString(secretMaterial)
	header, err := secretHeader(kind)
	if err != nil {
		return developeroauthdomain.IssuedSecret{}, err
	}
	digest, err := manager.digest(manager.active, kind, id, prefix, secret)
	if err != nil {
		return developeroauthdomain.IssuedSecret{}, err
	}
	return developeroauthdomain.IssuedSecret{
		Plaintext: developeroauthdomain.NewPlaintextSecret(header + prefix + "_" + secret),
		Material: developeroauthdomain.SecretMaterial{
			ID: id, Kind: kind, LookupPrefix: prefix, Digest: digest, DigestKey: manager.active,
		},
	}, nil
}

func (manager *TokenManager) ParseLookupPrefix(raw string, expected developeroauthdomain.SecretKind) (string, error) {
	parsed, err := parseSecret(raw)
	if err != nil || parsed.kind != expected {
		return "", invalidSecretError(expected)
	}
	return parsed.prefix, nil
}

func (manager *TokenManager) Verify(raw string, record developeroauthdomain.SecretMaterial) error {
	parsed, err := parseSecret(raw)
	if err != nil || parsed.kind != record.Kind || parsed.prefix != record.LookupPrefix {
		return invalidSecretError(record.Kind)
	}
	digest, err := manager.digest(record.DigestKey, record.Kind, record.ID, parsed.prefix, parsed.secret)
	if err != nil {
		return errors.Join(invalidSecretError(record.Kind), err)
	}
	if len(digest) != len(record.Digest) || subtle.ConstantTimeCompare(digest, record.Digest) != 1 {
		return invalidSecretError(record.Kind)
	}
	return nil
}

type parsedSecret struct {
	kind   developeroauthdomain.SecretKind
	prefix string
	secret string
}

func parseSecret(raw string) (parsedSecret, error) {
	var kind developeroauthdomain.SecretKind
	var header string
	switch {
	case strings.HasPrefix(raw, authorizationCodeHeader):
		kind, header = developeroauthdomain.SecretAuthorizationCode, authorizationCodeHeader
	case strings.HasPrefix(raw, refreshTokenHeader):
		kind, header = developeroauthdomain.SecretRefreshToken, refreshTokenHeader
	case strings.HasPrefix(raw, clientSecretHeader):
		kind, header = developeroauthdomain.SecretClientSecret, clientSecretHeader
	default:
		return parsedSecret{}, errors.New("unrecognized OAuth secret")
	}
	remainder := strings.TrimPrefix(raw, header)
	if len(remainder) != encodedPrefixLen+1+encodedSecretLen || remainder[encodedPrefixLen] != '_' {
		return parsedSecret{}, errors.New("malformed OAuth secret")
	}
	prefix := remainder[:encodedPrefixLen]
	secret := remainder[encodedPrefixLen+1:]
	if _, err := hex.DecodeString(prefix); err != nil {
		return parsedSecret{}, errors.New("malformed OAuth secret prefix")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil || len(decoded) != secretBytes {
		return parsedSecret{}, errors.New("malformed OAuth secret material")
	}
	zeroByteSlices([][]byte{decoded})
	return parsedSecret{kind: kind, prefix: prefix, secret: secret}, nil
}

func secretHeader(kind developeroauthdomain.SecretKind) (string, error) {
	switch kind {
	case developeroauthdomain.SecretAuthorizationCode:
		return authorizationCodeHeader, nil
	case developeroauthdomain.SecretRefreshToken:
		return refreshTokenHeader, nil
	case developeroauthdomain.SecretClientSecret:
		return clientSecretHeader, nil
	default:
		return "", errors.New("invalid OAuth secret kind")
	}
}

func invalidSecretError(kind developeroauthdomain.SecretKind) error {
	switch kind {
	case developeroauthdomain.SecretAuthorizationCode:
		return developeroauthdomain.ErrAuthorizationCode
	case developeroauthdomain.SecretClientSecret:
		return developeroauthdomain.ErrClientSecret
	default:
		return developeroauthdomain.ErrRefreshToken
	}
}

func (manager *TokenManager) digest(
	ref developeroauthdomain.DigestKeyRef,
	kind developeroauthdomain.SecretKind,
	id uuid.UUID,
	prefix string,
	secret string,
) ([]byte, error) {
	key, found := manager.keys[ref]
	if !found {
		return nil, developeroauthdomain.ErrTokenKeyUnavailable
	}
	mac := hmac.New(sha256.New, key)
	_, _ = io.WriteString(mac, "fortyone/developer-oauth/v1\x00")
	_, _ = io.WriteString(mac, string(kind))
	_, _ = io.WriteString(mac, "\x00"+id.String()+"\x00"+prefix+"\x00"+secret)
	return mac.Sum(nil), nil
}
