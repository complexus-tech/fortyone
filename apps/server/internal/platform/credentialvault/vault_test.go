package credentialvault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestVaultRoundTripAndRedaction(t *testing.T) {
	t.Parallel()
	vault := testVault(t, KeyRef{ID: "primary", Version: 2}, []Key{
		{Ref: KeyRef{ID: "primary", Version: 2}, Material: bytes.Repeat([]byte{0x42}, keySize)},
	})
	binding := testContext()

	sealed, err := vault.Seal(binding, []byte("github-secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if !strings.HasPrefix(sealed, EnvelopePrefix) || strings.Contains(sealed, "github-secret") {
		t.Fatalf("Seal() returned an invalid envelope")
	}
	opened, err := vault.Open(binding, sealed)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if got := string(opened.Reveal()); got != "github-secret" {
		t.Fatal("Open() returned different plaintext")
	}
	if opened.String() != "[REDACTED]" {
		t.Fatalf("String() = %q, want redaction", opened.String())
	}
	encoded, err := json.Marshal(opened)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if strings.Contains(string(encoded), "github-secret") {
		t.Fatal("MarshalJSON() exposed plaintext")
	}
	for _, formatted := range []string{
		fmt.Sprintf("%v", opened),
		fmt.Sprintf("%+v", opened),
		fmt.Sprintf("%#v", opened),
	} {
		if strings.Contains(formatted, "github-secret") || strings.Contains(formatted, "103 105 116 104 117 98") {
			t.Fatalf("formatted Secret exposed plaintext: %q", formatted)
		}
	}
}

func TestVaultRejectsTamperingAndWrongContext(t *testing.T) {
	t.Parallel()
	vault := testVault(t, KeyRef{ID: "primary", Version: 1}, []Key{
		{Ref: KeyRef{ID: "primary", Version: 1}, Material: bytes.Repeat([]byte{0x21}, keySize)},
	})
	binding := testContext()
	sealed, err := vault.Seal(binding, []byte("slack-secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	for name, test := range map[string]struct {
		mutate func(string) string
		want   error
	}{
		"ciphertext": {mutate: func(value string) string {
			document := decodeTestEnvelope(t, value)
			document.Ciphertext = flipEncodedByte(t, document.Ciphertext)
			return encodeTestEnvelope(t, document)
		}, want: ErrAuthentication},
		"wrapped data key": {mutate: func(value string) string {
			document := decodeTestEnvelope(t, value)
			document.WrappedDEK = flipEncodedByte(t, document.WrappedDEK)
			return encodeTestEnvelope(t, document)
		}, want: ErrAuthentication},
		"key version": {mutate: func(value string) string {
			document := decodeTestEnvelope(t, value)
			document.KEKVersion = 2
			return encodeTestEnvelope(t, document)
		}, want: ErrUnknownKey},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := vault.Open(binding, test.mutate(sealed))
			if !errors.Is(err, test.want) {
				t.Fatalf("Open() error = %v, want %v", err, test.want)
			}
		})
	}

	wrongContexts := map[string]Context{
		"provider":        {Provider: "github", TenantID: binding.TenantID, SubjectID: binding.SubjectID, CredentialType: binding.CredentialType, Generation: binding.Generation},
		"tenant":          {Provider: binding.Provider, TenantID: "other-workspace", SubjectID: binding.SubjectID, CredentialType: binding.CredentialType, Generation: binding.Generation},
		"subject":         {Provider: binding.Provider, TenantID: binding.TenantID, SubjectID: "other-installation", CredentialType: binding.CredentialType, Generation: binding.Generation},
		"credential type": {Provider: binding.Provider, TenantID: binding.TenantID, SubjectID: binding.SubjectID, CredentialType: "refresh-token", Generation: binding.Generation},
		"generation":      {Provider: binding.Provider, TenantID: binding.TenantID, SubjectID: binding.SubjectID, CredentialType: binding.CredentialType, Generation: "generation-2"},
	}
	for name, wrong := range wrongContexts {
		t.Run("wrong "+name, func(t *testing.T) {
			_, err := vault.Open(wrong, sealed)
			if !errors.Is(err, ErrAuthentication) {
				t.Fatalf("Open() error = %v, want ErrAuthentication", err)
			}
		})
	}
}

func TestVaultUsesKeyIDAndVersionForRotation(t *testing.T) {
	t.Parallel()
	oldRef := KeyRef{ID: "provider-credentials", Version: 1}
	newRef := KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x11}, keySize)}
	newKey := Key{Ref: newRef, Material: bytes.Repeat([]byte{0x22}, keySize)}
	binding := testContext()

	oldVault := testVault(t, oldRef, []Key{oldKey})
	oldEnvelope, err := oldVault.Seal(binding, []byte("credential"))
	if err != nil {
		t.Fatalf("old Seal() error = %v", err)
	}
	rotatedVault := testVault(t, newRef, []Key{oldKey, newKey})
	if _, err := rotatedVault.Open(binding, oldEnvelope); err != nil {
		t.Fatalf("rotated Open(old) error = %v", err)
	}
	newEnvelope, err := rotatedVault.Seal(binding, []byte("credential"))
	if err != nil {
		t.Fatalf("rotated Seal() error = %v", err)
	}
	if got := decodeTestEnvelope(t, newEnvelope); got.KEKID != newRef.ID || got.KEKVersion != newRef.Version {
		t.Fatalf("new envelope key = %s@%d, want %s@%d", got.KEKID, got.KEKVersion, newRef.ID, newRef.Version)
	}
	withoutOld := testVault(t, newRef, []Key{newKey})
	if _, err := withoutOld.Open(binding, oldEnvelope); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("Open(old without key) error = %v, want ErrUnknownKey", err)
	}
}

func TestVaultRewrapPreservesCredentialCiphertextAndGeneration(t *testing.T) {
	t.Parallel()
	oldRef := KeyRef{ID: "provider-credentials", Version: 7}
	newRef := KeyRef{ID: "provider-credentials", Version: 8}
	oldKey := Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x17}, keySize)}
	newKey := Key{Ref: newRef, Material: bytes.Repeat([]byte{0x28}, keySize)}
	binding := testContext()

	oldVault := testVault(t, oldRef, []Key{oldKey})
	oldEnvelope, err := oldVault.Seal(binding, []byte("credential-to-rewrap"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	before := decodeTestEnvelope(t, oldEnvelope)
	metadata, err := Inspect(oldEnvelope)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if metadata.Key != oldRef || metadata.Version != CurrentVersion || metadata.AADVersion != aadVersion {
		t.Fatalf("Inspect() = %#v", metadata)
	}

	rotatedVault := testVault(t, newRef, []Key{oldKey, newKey})
	result, err := rotatedVault.Rewrap(binding, oldEnvelope)
	if err != nil {
		t.Fatalf("Rewrap() error = %v", err)
	}
	if !result.Changed || result.From != oldRef || result.To != newRef || result.Envelope == oldEnvelope {
		t.Fatal("Rewrap() returned invalid rotation metadata")
	}
	after := decodeTestEnvelope(t, result.Envelope)
	if after.Nonce != before.Nonce || after.Ciphertext != before.Ciphertext {
		t.Fatal("Rewrap() changed the credential ciphertext instead of only rewrapping its DEK")
	}
	if after.WrappedDEK == before.WrappedDEK {
		t.Fatal("Rewrap() retained the previous DEK wrapping")
	}
	if after.KEKID != newRef.ID || after.KEKVersion != newRef.Version {
		t.Fatalf("rewrapped key = %s@%d", after.KEKID, after.KEKVersion)
	}

	opened, err := rotatedVault.Open(binding, result.Envelope)
	if err != nil {
		t.Fatalf("Open(rewrapped) error = %v", err)
	}
	defer opened.Destroy()
	if got := string(opened.Reveal()); got != "credential-to-rewrap" {
		t.Fatal("Open(rewrapped) returned different plaintext")
	}
	wrongGeneration := binding
	wrongGeneration.Generation = "generation-2"
	if _, err := rotatedVault.Open(wrongGeneration, result.Envelope); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Open(rewrapped with wrong generation) error = %v", err)
	}

	idempotent, err := rotatedVault.Rewrap(binding, result.Envelope)
	if err != nil {
		t.Fatalf("Rewrap(current) error = %v", err)
	}
	if idempotent.Changed || idempotent.Envelope != result.Envelope || idempotent.From != newRef || idempotent.To != newRef {
		t.Fatal("Rewrap(current) did not return an authenticated exact no-op")
	}
}

func TestVaultErrorsDoNotExposeCredentialOrEnvelope(t *testing.T) {
	t.Parallel()
	vault := testVault(t, KeyRef{ID: "safe-errors", Version: 1}, []Key{{
		Ref:      KeyRef{ID: "safe-errors", Version: 1},
		Material: bytes.Repeat([]byte{0x38}, keySize),
	}})
	binding := testContext()
	const plaintext = "credential-must-not-enter-errors"
	sealed, err := vault.Seal(binding, []byte(plaintext))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	wrong := binding
	wrong.Generation = "wrong-generation"
	_, openErr := vault.Open(wrong, sealed)
	if !errors.Is(openErr, ErrAuthentication) {
		t.Fatalf("Open(wrong generation) error = %v", openErr)
	}
	tampered := decodeTestEnvelope(t, sealed)
	tampered.Ciphertext = flipEncodedByte(t, tampered.Ciphertext)
	_, rewrapErr := vault.Rewrap(binding, encodeTestEnvelope(t, tampered))
	if !errors.Is(rewrapErr, ErrAuthentication) {
		t.Fatalf("Rewrap(tampered) error = %v", rewrapErr)
	}
	for _, err := range []error{openErr, rewrapErr} {
		message := err.Error()
		if strings.Contains(message, plaintext) || strings.Contains(message, sealed) {
			t.Fatal("credential vault error exposed credential material")
		}
	}
}

func TestVaultRewrapFailsClosedForUnknownAndTamperedEnvelopes(t *testing.T) {
	t.Parallel()
	oldRef := KeyRef{ID: "old", Version: 1}
	newRef := KeyRef{ID: "new", Version: 1}
	oldKey := Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x31}, keySize)}
	newKey := Key{Ref: newRef, Material: bytes.Repeat([]byte{0x42}, keySize)}
	binding := testContext()
	oldVault := testVault(t, oldRef, []Key{oldKey})
	oldEnvelope, err := oldVault.Seal(binding, []byte("credential"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	withoutOld := testVault(t, newRef, []Key{newKey})
	if _, err := withoutOld.Rewrap(binding, oldEnvelope); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("Rewrap(unknown old key) error = %v", err)
	}

	rotatedVault := testVault(t, newRef, []Key{oldKey, newKey})
	tampered := decodeTestEnvelope(t, oldEnvelope)
	tampered.Ciphertext = flipEncodedByte(t, tampered.Ciphertext)
	if _, err := rotatedVault.Rewrap(binding, encodeTestEnvelope(t, tampered)); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Rewrap(tampered ciphertext) error = %v", err)
	}

	currentEnvelope, err := rotatedVault.Seal(binding, []byte("current-credential"))
	if err != nil {
		t.Fatalf("Seal(current) error = %v", err)
	}
	tampered = decodeTestEnvelope(t, currentEnvelope)
	tampered.Ciphertext = flipEncodedByte(t, tampered.Ciphertext)
	if _, err := rotatedVault.Rewrap(binding, encodeTestEnvelope(t, tampered)); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("Rewrap(tampered current ciphertext) error = %v", err)
	}
}

func TestVaultConcurrentSealOpenAndRewrap(t *testing.T) {
	t.Parallel()
	oldRef := KeyRef{ID: "concurrent", Version: 1}
	newRef := KeyRef{ID: "concurrent", Version: 2}
	oldKey := Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x51}, keySize)}
	newKey := Key{Ref: newRef, Material: bytes.Repeat([]byte{0x62}, keySize)}
	oldVault, err := New(Config{Active: oldRef, Keys: []Key{oldKey}})
	if err != nil {
		t.Fatalf("New(old) error = %v", err)
	}
	rotatedVault, err := New(Config{Active: newRef, Keys: []Key{oldKey, newKey}})
	if err != nil {
		t.Fatalf("New(rotated) error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const workers = 32
	var wait sync.WaitGroup
	errorsByWorker := make(chan error, workers)
	for index := 0; index < workers; index++ {
		index := index
		wait.Add(1)
		go func() {
			defer wait.Done()
			select {
			case <-ctx.Done():
				errorsByWorker <- ctx.Err()
				return
			default:
			}
			binding := testContext()
			binding.SubjectID = fmt.Sprintf("installation-%d", index)
			plaintext := []byte(fmt.Sprintf("credential-%d", index))
			sealed, err := oldVault.Seal(binding, plaintext)
			if err != nil {
				errorsByWorker <- err
				return
			}
			rewrapped, err := rotatedVault.Rewrap(binding, sealed)
			if err != nil {
				errorsByWorker <- err
				return
			}
			opened, err := rotatedVault.Open(binding, rewrapped.Envelope)
			if err != nil {
				errorsByWorker <- err
				return
			}
			defer opened.Destroy()
			copy := opened.Reveal()
			defer zero(copy)
			if !bytes.Equal(copy, plaintext) {
				errorsByWorker <- errors.New("opened credential did not round trip")
			}
		}()
	}
	wait.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		if err != nil {
			t.Fatalf("concurrent vault operation: %v", err)
		}
	}
}

func TestSecretDestroyClearsVaultOwnedPlaintext(t *testing.T) {
	t.Parallel()
	secret := Secret{value: []byte("sensitive")}
	owned := secret.value
	secret.Destroy()
	if secret.value != nil {
		t.Fatal("Destroy() retained the vault-owned buffer")
	}
	if !bytes.Equal(owned, make([]byte, len(owned))) {
		t.Fatal("Destroy() did not clear the vault-owned buffer")
	}
}

func TestVaultRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	validKey := Key{Ref: KeyRef{ID: "primary", Version: 1}, Material: bytes.Repeat([]byte{1}, keySize)}
	if _, err := New(Config{Active: validKey.Ref, Keys: []Key{{Ref: validKey.Ref, Material: []byte("short")}}}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("New(short key) error = %v, want ErrNotConfigured", err)
	}
	vault := testVault(t, validKey.Ref, []Key{validKey})
	if _, err := vault.Seal(Context{}, []byte("secret")); !errors.Is(err, ErrInvalidContext) {
		t.Fatalf("Seal(invalid context) error = %v, want ErrInvalidContext", err)
	}
	if _, err := vault.Seal(testContext(), nil); !errors.Is(err, ErrEmptyPlaintext) {
		t.Fatalf("Seal(empty) error = %v, want ErrEmptyPlaintext", err)
	}
	if _, err := vault.Seal(testContext(), make([]byte, maxPlaintext+1)); !errors.Is(err, ErrPlaintextTooLarge) {
		t.Fatalf("Seal(oversized) error = %v, want ErrPlaintextTooLarge", err)
	}
	if _, err := vault.Open(testContext(), "xoxb-plaintext"); !errors.Is(err, ErrMalformedEnvelope) {
		t.Fatalf("Open(plaintext) error = %v, want ErrMalformedEnvelope", err)
	}
	if _, err := vault.Open(testContext(), "vault.v3.invalid"); !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("Open(unknown version) error = %v, want ErrUnsupportedVersion", err)
	}
	if _, err := vault.Open(testContext(), strings.Repeat("x", maxEnvelope+1)); !errors.Is(err, ErrMalformedEnvelope) {
		t.Fatalf("Open(oversized) error = %v, want ErrMalformedEnvelope", err)
	}
	sealed, err := vault.Seal(testContext(), []byte("secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	document, err := decode(strings.TrimPrefix(sealed, EnvelopePrefix))
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	withTrailingJSON := EnvelopePrefix + encode(append(document, []byte(` {"extra":true}`)...))
	if _, err := vault.Open(testContext(), withTrailingJSON); !errors.Is(err, ErrMalformedEnvelope) {
		t.Fatalf("Open(trailing JSON) error = %v, want ErrMalformedEnvelope", err)
	}
}

func TestParseEncodedKeyring(t *testing.T) {
	t.Parallel()
	material := bytes.Repeat([]byte{0x7a}, keySize)
	encoded := `{"blue@3":"` + base64.StdEncoding.EncodeToString(material) + `"}`
	cfg, err := ParseEncodedKeyring("blue", 3, encoded)
	if err != nil {
		t.Fatalf("ParseEncodedKeyring() error = %v", err)
	}
	if cfg.Active != (KeyRef{ID: "blue", Version: 3}) || len(cfg.Keys) != 1 || !bytes.Equal(cfg.Keys[0].Material, material) {
		t.Fatalf("ParseEncodedKeyring() = %#v", cfg)
	}
	if _, err := ParseEncodedKeyring("blue", 3, encoded+` {"other@1":"unused"}`); err == nil {
		t.Fatal("ParseEncodedKeyring(trailing JSON) error = nil")
	}
	duplicate := `{"blue@3":"` + base64.StdEncoding.EncodeToString(material) + `","blue@3":"` + base64.StdEncoding.EncodeToString(material) + `"}`
	if _, err := ParseEncodedKeyring("blue", 3, duplicate); err == nil {
		t.Fatal("ParseEncodedKeyring(duplicate key) error = nil")
	}
}

func TestContainsDevelopmentKeyUsesSemanticKeyringValidation(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		activeID string
		encoded  string
	}{
		"formatted default": {activeID: DevelopmentKeyID, encoded: `{ "development@1" : "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" }`},
		"renamed zero key":  {activeID: "production", encoded: `{"production@1":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`},
	} {
		t.Run(name, func(t *testing.T) {
			contains, err := ContainsDevelopmentKey(test.activeID, 1, test.encoded)
			if err != nil {
				t.Fatalf("ContainsDevelopmentKey() error = %v", err)
			}
			if !contains {
				t.Fatal("ContainsDevelopmentKey() = false")
			}
		})
	}
	productionMaterial := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	contains, err := ContainsDevelopmentKey("production", 1, `{"production@1":"`+productionMaterial+`"}`)
	if err != nil {
		t.Fatalf("ContainsDevelopmentKey(production) error = %v", err)
	}
	if contains {
		t.Fatal("ContainsDevelopmentKey(production) = true")
	}
}

func TestReusesSecretMaterialRecognizesRawAndEncodedSecrets(t *testing.T) {
	t.Parallel()
	material := []byte("0123456789abcdef0123456789abcdef")
	encodedKeyring := `{"production@1":"` + base64.StdEncoding.EncodeToString(material) + `"}`
	for name, secret := range map[string]string{
		"raw":     string(material),
		"base64":  base64.StdEncoding.EncodeToString(material),
		"trimmed": " " + string(material) + " ",
	} {
		t.Run(name, func(t *testing.T) {
			reused, err := ReusesSecretMaterial("production", 1, encodedKeyring, secret)
			if err != nil {
				t.Fatalf("ReusesSecretMaterial() error = %v", err)
			}
			if !reused {
				t.Fatal("ReusesSecretMaterial() = false")
			}
		})
	}
	reused, err := ReusesSecretMaterial("production", 1, encodedKeyring, "different-independent-secret-12345")
	if err != nil {
		t.Fatalf("ReusesSecretMaterial(independent) error = %v", err)
	}
	if reused {
		t.Fatal("ReusesSecretMaterial(independent) = true")
	}
}

func TestVaultUsesFreshRandomnessPerEnvelope(t *testing.T) {
	t.Parallel()
	ref := KeyRef{ID: "randomness", Version: 1}
	vault, err := New(Config{Active: ref, Keys: []Key{{Ref: ref, Material: bytes.Repeat([]byte{0x31}, keySize)}}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	first, err := vault.Seal(testContext(), []byte("same-credential"))
	if err != nil {
		t.Fatalf("first Seal() error = %v", err)
	}
	second, err := vault.Seal(testContext(), []byte("same-credential"))
	if err != nil {
		t.Fatalf("second Seal() error = %v", err)
	}
	if first == second {
		t.Fatal("Seal() reused data-key or nonce material")
	}
}

func testVault(t *testing.T, active KeyRef, keys []Key) *Vault {
	t.Helper()
	random := bytes.NewReader(bytes.Repeat([]byte{0x5a}, 256))
	vault, err := newWithReader(Config{Active: active, Keys: keys}, random)
	if err != nil {
		t.Fatalf("newWithReader() error = %v", err)
	}
	return vault
}

func testContext() Context {
	return Context{
		Provider:       "slack",
		TenantID:       "workspace-1",
		SubjectID:      "installation-1",
		CredentialType: "bot-oauth",
		Generation:     "generation-1",
	}
}

func decodeTestEnvelope(t *testing.T, value string) envelope {
	t.Helper()
	raw, err := decode(strings.TrimPrefix(value, EnvelopePrefix))
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var document envelope
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return document
}

func encodeTestEnvelope(t *testing.T, document envelope) string {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return EnvelopePrefix + encode(raw)
}

func flipEncodedByte(t *testing.T, value string) string {
	t.Helper()
	raw, err := decode(value)
	if err != nil {
		t.Fatalf("decode field: %v", err)
	}
	raw[len(raw)-1] ^= 1
	return encode(raw)
}
