// Package credentialvault encrypts retained provider credentials with a
// versioned envelope and context-bound authenticated encryption.
package credentialvault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	// CurrentVersion follows the pre-existing v1 integration secret envelope.
	// Version 2 is the first shared, context-bound envelope-encryption format.
	CurrentVersion = 2
	// EnvelopePrefix identifies the current serialized vault format.
	EnvelopePrefix = "vault.v2."

	algorithm    = "AES-256-GCM+AES-256-GCM-KW"
	aadVersion   = 1
	dekSize      = 32
	keySize      = 32
	maxPlaintext = 64 << 10
	maxEnvelope  = 256 << 10
)

var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

var (
	ErrNotConfigured      = errors.New("credential vault is not configured")
	ErrInvalidContext     = errors.New("credential vault context is invalid")
	ErrEmptyPlaintext     = errors.New("credential vault plaintext is empty")
	ErrPlaintextTooLarge  = errors.New("credential vault plaintext is too large")
	ErrMalformedEnvelope  = errors.New("credential vault envelope is malformed")
	ErrUnsupportedVersion = errors.New("credential vault envelope version is unsupported")
	ErrUnknownKey         = errors.New("credential vault key is unavailable")
	ErrAuthentication     = errors.New("credential vault authentication failed")
)

// Context is authenticated with the credential and is never inferred from the
// envelope. Supplying a credential under another tenant, provider, subject,
// type, or generation therefore fails authentication.
type Context struct {
	Provider       string `json:"provider"`
	TenantID       string `json:"tenant_id"`
	SubjectID      string `json:"subject_id"`
	CredentialType string `json:"credential_type"`
	Generation     string `json:"generation"`
}

// KeyRef identifies one key-encryption key generation. A key ID can retain
// several versions during a rotation window.
type KeyRef struct {
	ID      string
	Version uint32
}

// Metadata is the non-secret serialized envelope metadata that operators
// can use to plan a key rotation. Inspecting metadata does not prove envelope
// authenticity; Open or Rewrap performs that check with the caller-supplied
// credential context.
type Metadata struct {
	Version    int
	Algorithm  string
	AADVersion int
	Key        KeyRef
}

// RewrapResult describes one logical KEK rotation. Rewrap preserves the
// credential ciphertext and its provider generation; only the wrapped DEK and
// key reference change. An envelope already using the active key is returned
// unchanged after its authentication tags have been verified.
type RewrapResult struct {
	Envelope string
	From     KeyRef
	To       KeyRef
	Changed  bool
}

// Key is keyring configuration. Material must contain exactly 32 random bytes.
type Key struct {
	Ref      KeyRef
	Material []byte
}

// Config defines the active encryption key and every key accepted for
// decryption. Retired keys can be removed after all envelopes have been
// rewrapped or expired.
type Config struct {
	Active KeyRef
	Keys   []Key
}

// Secret is a redacted plaintext holder. Reveal returns a copy so callers
// cannot mutate the vault-owned bytes. String and JSON representations never
// disclose the credential.
type Secret struct {
	value []byte
}

func (s Secret) Reveal() []byte {
	return append([]byte(nil), s.value...)
}

// Destroy clears the vault-owned plaintext buffer. Callers must separately
// clear any copy returned by Reveal.
func (s *Secret) Destroy() {
	if s == nil {
		return
	}
	zero(s.value)
	s.value = nil
}

func (Secret) String() string {
	return "[REDACTED]"
}

func (Secret) GoString() string {
	return "credentialvault.Secret{[REDACTED]}"
}

func (Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

type envelope struct {
	Version    int    `json:"version"`
	Algorithm  string `json:"algorithm"`
	AADVersion int    `json:"aad_version"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	WrapNonce  string `json:"wrap_nonce"`
	WrappedDEK string `json:"wrapped_dek"`
	KEKID      string `json:"kek_id"`
	KEKVersion uint32 `json:"kek_version"`
}

type Vault struct {
	active KeyRef
	keys   map[KeyRef][keySize]byte
	random io.Reader
}

func New(cfg Config) (*Vault, error) {
	return newWithReader(cfg, rand.Reader)
}

func newWithReader(cfg Config, random io.Reader) (*Vault, error) {
	if random == nil {
		return nil, fmt.Errorf("%w: random source is required", ErrNotConfigured)
	}
	if err := validateKeyRef(cfg.Active); err != nil {
		return nil, fmt.Errorf("%w: active key: %v", ErrNotConfigured, err)
	}
	keys := make(map[KeyRef][keySize]byte, len(cfg.Keys))
	for _, configured := range cfg.Keys {
		if err := validateKeyRef(configured.Ref); err != nil {
			return nil, fmt.Errorf("%w: key reference: %v", ErrNotConfigured, err)
		}
		if len(configured.Material) != keySize {
			return nil, fmt.Errorf("%w: key %s@%d must contain exactly %d bytes", ErrNotConfigured, configured.Ref.ID, configured.Ref.Version, keySize)
		}
		if _, exists := keys[configured.Ref]; exists {
			return nil, fmt.Errorf("%w: duplicate key %s@%d", ErrNotConfigured, configured.Ref.ID, configured.Ref.Version)
		}
		var material [keySize]byte
		copy(material[:], configured.Material)
		keys[configured.Ref] = material
	}
	if _, ok := keys[cfg.Active]; !ok {
		return nil, fmt.Errorf("%w: active key %s@%d is absent from the keyring", ErrNotConfigured, cfg.Active.ID, cfg.Active.Version)
	}
	return &Vault{active: cfg.Active, keys: keys, random: random}, nil
}

func (v *Vault) Seal(binding Context, plaintext []byte) (string, error) {
	if v == nil || v.random == nil || len(v.keys) == 0 {
		return "", ErrNotConfigured
	}
	if len(plaintext) == 0 {
		return "", ErrEmptyPlaintext
	}
	if len(plaintext) > maxPlaintext {
		return "", ErrPlaintextTooLarge
	}
	aad, err := binding.additionalData()
	if err != nil {
		return "", err
	}
	kek, ok := v.keys[v.active]
	if !ok {
		return "", ErrUnknownKey
	}

	dek := make([]byte, dekSize)
	if _, err := io.ReadFull(v.random, dek); err != nil {
		return "", fmt.Errorf("credential vault: generate data key: %w", err)
	}
	defer zero(dek)

	payloadAEAD, err := newAEAD(dek)
	if err != nil {
		return "", fmt.Errorf("credential vault: create payload cipher: %w", err)
	}
	payloadNonce, err := randomNonce(v.random, payloadAEAD.NonceSize())
	if err != nil {
		return "", fmt.Errorf("credential vault: generate payload nonce: %w", err)
	}
	ciphertext := payloadAEAD.Seal(nil, payloadNonce, plaintext, aad)

	wrapAEAD, err := newAEAD(kek[:])
	if err != nil {
		return "", fmt.Errorf("credential vault: create key-wrap cipher: %w", err)
	}
	wrapNonce, err := randomNonce(v.random, wrapAEAD.NonceSize())
	if err != nil {
		return "", fmt.Errorf("credential vault: generate key-wrap nonce: %w", err)
	}
	wrappedDEK := wrapAEAD.Seal(nil, wrapNonce, dek, wrapAdditionalData(v.active, aad))

	encoded, err := json.Marshal(envelope{
		Version:    CurrentVersion,
		Algorithm:  algorithm,
		AADVersion: aadVersion,
		Nonce:      encode(payloadNonce),
		Ciphertext: encode(ciphertext),
		WrapNonce:  encode(wrapNonce),
		WrappedDEK: encode(wrappedDEK),
		KEKID:      v.active.ID,
		KEKVersion: v.active.Version,
	})
	if err != nil {
		return "", fmt.Errorf("credential vault: encode envelope: %w", err)
	}
	return EnvelopePrefix + encode(encoded), nil
}

func (v *Vault) Open(binding Context, encoded string) (Secret, error) {
	if v == nil || len(v.keys) == 0 {
		return Secret{}, ErrNotConfigured
	}
	aad, err := binding.additionalData()
	if err != nil {
		return Secret{}, err
	}
	stored, err := decodeEnvelope(encoded)
	if err != nil {
		return Secret{}, err
	}
	dek, err := v.unwrapDataKey(stored, aad)
	if err != nil {
		return Secret{}, err
	}
	defer zero(dek)
	plaintext, err := openPayload(stored, aad, dek)
	if err != nil {
		return Secret{}, err
	}
	return Secret{value: plaintext}, nil
}

// ActiveKeyRef returns the key generation used for new envelopes and rewraps.
// The returned reference contains no key material.
func (v *Vault) ActiveKeyRef() (KeyRef, error) {
	if v == nil || len(v.keys) == 0 {
		return KeyRef{}, ErrNotConfigured
	}
	return v.active, nil
}

// Inspect returns non-secret envelope metadata without decrypting the
// credential. Callers must use Open or Rewrap before trusting the envelope.
func Inspect(encoded string) (Metadata, error) {
	stored, err := decodeEnvelope(encoded)
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{
		Version:    stored.Version,
		Algorithm:  stored.Algorithm,
		AADVersion: stored.AADVersion,
		Key:        KeyRef{ID: stored.KEKID, Version: stored.KEKVersion},
	}, nil
}

// Rewrap verifies the complete envelope and moves only its DEK wrapping from
// the referenced KEK to the active KEK. It never returns credential plaintext,
// never changes the credential generation in AAD, and is logically idempotent.
// A caller persists Changed results with a compare-and-swap over the original
// envelope and provider generation.
func (v *Vault) Rewrap(binding Context, encoded string) (RewrapResult, error) {
	if v == nil || v.random == nil || len(v.keys) == 0 {
		return RewrapResult{}, ErrNotConfigured
	}
	aad, err := binding.additionalData()
	if err != nil {
		return RewrapResult{}, err
	}
	stored, err := decodeEnvelope(encoded)
	if err != nil {
		return RewrapResult{}, err
	}
	from := KeyRef{ID: stored.KEKID, Version: stored.KEKVersion}
	dek, err := v.unwrapDataKey(stored, aad)
	if err != nil {
		return RewrapResult{}, err
	}
	defer zero(dek)

	// Validate the payload tag before promoting an envelope. The plaintext is
	// never exposed and is cleared immediately after authentication.
	plaintext, err := openPayload(stored, aad, dek)
	if err != nil {
		return RewrapResult{}, err
	}
	zero(plaintext)
	if from == v.active {
		return RewrapResult{Envelope: encoded, From: from, To: v.active}, nil
	}

	kek, ok := v.keys[v.active]
	if !ok {
		return RewrapResult{}, ErrUnknownKey
	}
	wrapAEAD, err := newAEAD(kek[:])
	if err != nil {
		return RewrapResult{}, fmt.Errorf("credential vault: create key-wrap cipher: %w", err)
	}
	wrapNonce, err := randomNonce(v.random, wrapAEAD.NonceSize())
	if err != nil {
		return RewrapResult{}, fmt.Errorf("credential vault: generate key-wrap nonce: %w", err)
	}
	stored.WrapNonce = encode(wrapNonce)
	stored.WrappedDEK = encode(wrapAEAD.Seal(nil, wrapNonce, dek, wrapAdditionalData(v.active, aad)))
	stored.KEKID = v.active.ID
	stored.KEKVersion = v.active.Version
	rewrapped, err := encodeEnvelope(stored)
	if err != nil {
		return RewrapResult{}, err
	}
	return RewrapResult{
		Envelope: rewrapped,
		From:     from,
		To:       v.active,
		Changed:  true,
	}, nil
}

func IsEnvelope(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), EnvelopePrefix)
}

func decodeEnvelope(encoded string) (envelope, error) {
	if len(encoded) > maxEnvelope {
		return envelope{}, ErrMalformedEnvelope
	}
	encoded = strings.TrimSpace(encoded)
	if !strings.HasPrefix(encoded, EnvelopePrefix) {
		if strings.HasPrefix(encoded, "vault.v") {
			return envelope{}, ErrUnsupportedVersion
		}
		return envelope{}, ErrMalformedEnvelope
	}
	raw, err := decode(strings.TrimPrefix(encoded, EnvelopePrefix))
	if err != nil {
		return envelope{}, fmt.Errorf("%w: invalid envelope encoding", ErrMalformedEnvelope)
	}
	var stored envelope
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&stored); err != nil {
		return envelope{}, fmt.Errorf("%w: invalid envelope document", ErrMalformedEnvelope)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return envelope{}, fmt.Errorf("%w: invalid envelope document", ErrMalformedEnvelope)
	}
	if stored.Version != CurrentVersion || stored.Algorithm != algorithm || stored.AADVersion != aadVersion {
		return envelope{}, ErrUnsupportedVersion
	}
	if err := validateKeyRef(KeyRef{ID: stored.KEKID, Version: stored.KEKVersion}); err != nil {
		return envelope{}, ErrMalformedEnvelope
	}
	return stored, nil
}

func encodeEnvelope(stored envelope) (string, error) {
	encoded, err := json.Marshal(stored)
	if err != nil {
		return "", fmt.Errorf("credential vault: encode envelope: %w", err)
	}
	return EnvelopePrefix + encode(encoded), nil
}

func (v *Vault) unwrapDataKey(stored envelope, aad []byte) ([]byte, error) {
	ref := KeyRef{ID: stored.KEKID, Version: stored.KEKVersion}
	kek, ok := v.keys[ref]
	if !ok {
		return nil, ErrUnknownKey
	}
	wrapNonce, err := decode(stored.WrapNonce)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid key-wrap nonce", ErrMalformedEnvelope)
	}
	wrappedDEK, err := decode(stored.WrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid wrapped key", ErrMalformedEnvelope)
	}
	wrapAEAD, err := newAEAD(kek[:])
	if err != nil {
		return nil, fmt.Errorf("credential vault: create key-wrap cipher: %w", err)
	}
	if len(wrapNonce) != wrapAEAD.NonceSize() {
		return nil, fmt.Errorf("%w: invalid key-wrap nonce length", ErrMalformedEnvelope)
	}
	dek, err := wrapAEAD.Open(nil, wrapNonce, wrappedDEK, wrapAdditionalData(ref, aad))
	if err != nil {
		return nil, ErrAuthentication
	}
	if len(dek) != dekSize {
		zero(dek)
		return nil, fmt.Errorf("%w: invalid data key length", ErrMalformedEnvelope)
	}
	return dek, nil
}

func openPayload(stored envelope, aad, dek []byte) ([]byte, error) {
	payloadNonce, err := decode(stored.Nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid payload nonce", ErrMalformedEnvelope)
	}
	ciphertext, err := decode(stored.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ciphertext", ErrMalformedEnvelope)
	}
	payloadAEAD, err := newAEAD(dek)
	if err != nil {
		return nil, fmt.Errorf("credential vault: create payload cipher: %w", err)
	}
	if len(payloadNonce) != payloadAEAD.NonceSize() {
		return nil, fmt.Errorf("%w: invalid payload nonce length", ErrMalformedEnvelope)
	}
	plaintext, err := payloadAEAD.Open(nil, payloadNonce, ciphertext, aad)
	if err != nil {
		return nil, ErrAuthentication
	}
	if len(plaintext) == 0 || len(plaintext) > maxPlaintext {
		zero(plaintext)
		return nil, ErrMalformedEnvelope
	}
	return plaintext, nil
}

func (c Context) additionalData() ([]byte, error) {
	c.Provider = strings.ToLower(strings.TrimSpace(c.Provider))
	c.TenantID = strings.TrimSpace(c.TenantID)
	c.SubjectID = strings.TrimSpace(c.SubjectID)
	c.CredentialType = strings.ToLower(strings.TrimSpace(c.CredentialType))
	c.Generation = strings.TrimSpace(c.Generation)
	if c.Provider == "" || c.TenantID == "" || c.SubjectID == "" || c.CredentialType == "" || c.Generation == "" {
		return nil, ErrInvalidContext
	}
	encoded, err := json.Marshal(struct {
		Domain  string  `json:"domain"`
		Version int     `json:"version"`
		Context Context `json:"context"`
	}{
		Domain:  "fortyone:provider-credential",
		Version: aadVersion,
		Context: c,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: encode additional data", ErrInvalidContext)
	}
	return encoded, nil
}

func validateKeyRef(ref KeyRef) error {
	if ref.ID != strings.TrimSpace(ref.ID) || !keyIDPattern.MatchString(ref.ID) {
		return errors.New("key id must contain 1-64 safe characters")
	}
	if ref.Version == 0 {
		return errors.New("key version must be positive")
	}
	return nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func randomNonce(random io.Reader, size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(random, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func wrapAdditionalData(ref KeyRef, aad []byte) []byte {
	prefix := fmt.Sprintf("fortyone:credential-vault:dek:v1:%s:%d\n", ref.ID, ref.Version)
	return append([]byte(prefix), aad...)
}

func encode(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func decode(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}

func zero(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
