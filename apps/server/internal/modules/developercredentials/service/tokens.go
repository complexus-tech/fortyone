package developercredentials

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
	"strconv"
	"strings"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	"github.com/google/uuid"
)

const (
	tokenVersion     = int16(1)
	lookupPrefixSize = 6
	secretSize       = 32
	encodedPrefixLen = lookupPrefixSize * 2
	encodedSecretLen = 43
	personalHeader   = "f41_pat_v1_"
	serviceHeader    = "f41_sak_v1_"
)

type TokenManager struct {
	active developercredentialsdomain.DigestKeyRef
	keys   map[developercredentialsdomain.DigestKeyRef][]byte
	random io.Reader
}

func NewTokenManager(config TokenKeyringConfig) (*TokenManager, error) {
	return newTokenManager(config, rand.Reader)
}

func newTokenManager(config TokenKeyringConfig, random io.Reader) (*TokenManager, error) {
	if random == nil {
		return nil, errors.New("API credential random source is required")
	}
	if err := validateDigestKeyRef(config.Active); err != nil {
		return nil, err
	}
	keys := make(map[developercredentialsdomain.DigestKeyRef][]byte, len(config.Keys))
	for _, key := range config.Keys {
		if err := validateDigestKeyRef(key.Ref); err != nil {
			return nil, err
		}
		if len(key.Material) != digestKeyBytes {
			return nil, fmt.Errorf("API credential digest key %s@%d must contain %d bytes", key.Ref.ID, key.Ref.Version, digestKeyBytes)
		}
		if _, duplicate := keys[key.Ref]; duplicate {
			return nil, fmt.Errorf("duplicate API credential digest key %s@%d", key.Ref.ID, key.Ref.Version)
		}
		for existingRef, existingMaterial := range keys {
			if subtle.ConstantTimeCompare(existingMaterial, key.Material) == 1 {
				return nil, fmt.Errorf(
					"API credential digest keys %s@%d and %s@%d must use independent material",
					existingRef.ID, existingRef.Version, key.Ref.ID, key.Ref.Version,
				)
			}
		}
		keys[key.Ref] = append([]byte(nil), key.Material...)
	}
	if _, found := keys[config.Active]; !found {
		return nil, errors.New("active API credential digest key is unavailable")
	}
	return &TokenManager{active: config.Active, keys: keys, random: random}, nil
}

type issuedToken struct {
	Plaintext developercredentialsdomain.PlaintextToken
	Material  developercredentialsdomain.CredentialMaterial
}

func (manager *TokenManager) issue(kind developercredentialsdomain.CredentialKind, credentialID uuid.UUID) (issuedToken, error) {
	if manager == nil || credentialID == uuid.Nil {
		return issuedToken{}, errors.New("credential token manager and ID are required")
	}
	if err := kind.Validate(); err != nil {
		return issuedToken{}, err
	}
	prefixBytes := make([]byte, lookupPrefixSize)
	secretBytes := make([]byte, secretSize)
	if _, err := io.ReadFull(manager.random, prefixBytes); err != nil {
		return issuedToken{}, fmt.Errorf("generate credential lookup prefix: %w", err)
	}
	if _, err := io.ReadFull(manager.random, secretBytes); err != nil {
		return issuedToken{}, fmt.Errorf("generate credential secret: %w", err)
	}
	defer zeroCandidates([][]byte{prefixBytes, secretBytes})

	prefix := hex.EncodeToString(prefixBytes)
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	header, err := tokenHeader(kind)
	if err != nil {
		return issuedToken{}, err
	}
	plaintext := header + prefix + "_" + secret
	digest, err := manager.digest(manager.active, kind, credentialID, prefix, secret)
	if err != nil {
		return issuedToken{}, err
	}
	return issuedToken{
		Plaintext: developercredentialsdomain.NewPlaintextToken(plaintext),
		Material: developercredentialsdomain.CredentialMaterial{
			ID:           credentialID,
			Kind:         kind,
			LookupPrefix: prefix,
			SecretDigest: digest,
			TokenVersion: tokenVersion,
			DigestKey:    manager.active,
		},
	}, nil
}

func (manager *TokenManager) verify(raw string, record developercredentialsdomain.VerificationRecord) error {
	parsed, err := parseToken(raw)
	if err != nil {
		return developercredentialsdomain.ErrAuthenticationFailed
	}
	if parsed.Kind != record.CredentialKind ||
		parsed.Version != record.TokenVersion ||
		parsed.Prefix != record.LookupPrefix {
		return developercredentialsdomain.ErrAuthenticationFailed
	}
	digest, err := manager.digest(record.DigestKey, parsed.Kind, record.CredentialID, parsed.Prefix, parsed.Secret)
	if err != nil {
		return errors.Join(developercredentialsdomain.ErrAuthenticationFailed, err)
	}
	if len(digest) != len(record.SecretDigest) || subtle.ConstantTimeCompare(digest, record.SecretDigest) != 1 {
		return developercredentialsdomain.ErrAuthenticationFailed
	}
	return nil
}

type parsedToken struct {
	Kind    developercredentialsdomain.CredentialKind
	Version int16
	Prefix  string
	Secret  string
}

func parseToken(raw string) (parsedToken, error) {
	var kind developercredentialsdomain.CredentialKind
	var header string
	switch {
	case strings.HasPrefix(raw, personalHeader):
		kind, header = developercredentialsdomain.CredentialPersonalAccessToken, personalHeader
	case strings.HasPrefix(raw, serviceHeader):
		kind, header = developercredentialsdomain.CredentialServiceAccountKey, serviceHeader
	default:
		return parsedToken{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	remainder := strings.TrimPrefix(raw, header)
	if len(remainder) != encodedPrefixLen+1+encodedSecretLen || remainder[encodedPrefixLen] != '_' {
		return parsedToken{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	prefix := remainder[:encodedPrefixLen]
	secret := remainder[encodedPrefixLen+1:]
	if _, err := hex.DecodeString(prefix); err != nil {
		return parsedToken{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	decodedSecret, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil || len(decodedSecret) != secretSize {
		return parsedToken{}, developercredentialsdomain.ErrAuthenticationFailed
	}
	zeroCandidates([][]byte{decodedSecret})
	return parsedToken{Kind: kind, Version: tokenVersion, Prefix: prefix, Secret: secret}, nil
}

func tokenHeader(kind developercredentialsdomain.CredentialKind) (string, error) {
	switch kind {
	case developercredentialsdomain.CredentialPersonalAccessToken:
		return personalHeader, nil
	case developercredentialsdomain.CredentialServiceAccountKey:
		return serviceHeader, nil
	default:
		return "", developercredentialsdomain.ErrInvalidCredentialKind
	}
}

func (manager *TokenManager) digest(
	ref developercredentialsdomain.DigestKeyRef,
	kind developercredentialsdomain.CredentialKind,
	credentialID uuid.UUID,
	prefix string,
	secret string,
) ([]byte, error) {
	key, found := manager.keys[ref]
	if !found {
		return nil, developercredentialsdomain.ErrTokenKeyUnavailable
	}
	mac := hmac.New(sha256.New, key)
	_, _ = io.WriteString(mac, "fortyone/api-credential\x00")
	_, _ = io.WriteString(mac, strconv.FormatInt(int64(tokenVersion), 10))
	_, _ = io.WriteString(mac, "\x00"+string(kind)+"\x00")
	_, _ = io.WriteString(mac, credentialID.String())
	_, _ = io.WriteString(mac, "\x00"+prefix+"\x00"+secret)
	return mac.Sum(nil), nil
}
