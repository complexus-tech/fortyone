package invitations

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const invitationTokenTestSecret = "invitation-token-test-secret-that-is-at-least-32-bytes"

func TestInvitationTokenIssueLookupAndRestore(t *testing.T) {
	t.Parallel()

	manager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "2026-08-v1", Secret: invitationTokenTestSecret},
	}, bytes.NewReader(bytes.Repeat([]byte{0x42}, invitationTokenRandomBytes)))
	require.NoError(t, err)

	raw, stored, err := manager.Issue()
	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.NotContains(t, string(stored.Digest), raw)
	require.Len(t, stored.Digest, 32)
	require.Len(t, stored.Nonce, 32)
	require.Equal(t, "2026-08-v1", stored.KeyID)
	require.Equal(t, int16(1), stored.Version)

	lookup, err := manager.Lookup(raw)
	require.NoError(t, err)
	require.Equal(t, stored.Digest, lookup.Digest)
	require.Equal(t, stored.KeyID, lookup.KeyID)
	require.Equal(t, stored.Version, lookup.Version)
	require.Empty(t, lookup.LegacyToken)

	restored, err := manager.Restore(stored)
	require.NoError(t, err)
	require.Equal(t, raw, restored)
}

func TestInvitationTokenRotationRetainsOnlyConfiguredPreviousKeys(t *testing.T) {
	t.Parallel()

	oldManager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "old", Secret: invitationTokenTestSecret + "-old"},
	}, bytes.NewReader(bytes.Repeat([]byte{1}, invitationTokenRandomBytes)))
	require.NoError(t, err)
	oldRaw, oldStored, err := oldManager.Issue()
	require.NoError(t, err)

	rotated, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current:  InvitationTokenKey{ID: "new", Secret: invitationTokenTestSecret + "-new"},
		Previous: []InvitationTokenKey{{ID: "old", Secret: invitationTokenTestSecret + "-old"}},
	})
	require.NoError(t, err)
	lookup, err := rotated.Lookup(oldRaw)
	require.NoError(t, err)
	require.Equal(t, oldStored.Digest, lookup.Digest)
	restored, err := rotated.Restore(oldStored)
	require.NoError(t, err)
	require.Equal(t, oldRaw, restored)

	withoutOld, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "new", Secret: invitationTokenTestSecret + "-new"},
	})
	require.NoError(t, err)
	_, err = withoutOld.Lookup(oldRaw)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestInvitationTokenRejectsMalformedOrTamperedValues(t *testing.T) {
	t.Parallel()

	manager, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	})
	require.NoError(t, err)
	raw, stored, err := manager.Issue()
	require.NoError(t, err)
	parts := strings.Split(raw, ".")
	require.Len(t, parts, 4)

	signature, err := base64.RawURLEncoding.DecodeString(parts[3])
	require.NoError(t, err)
	signature[0] ^= 0xff
	tamperedSignature := strings.Join([]string{
		parts[0], parts[1], parts[2], base64.RawURLEncoding.EncodeToString(signature),
	}, ".")

	const rawURLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	lastCharacterIndex := strings.IndexByte(rawURLAlphabet, raw[len(raw)-1])
	require.GreaterOrEqual(t, lastCharacterIndex, 0)
	require.Zero(t, lastCharacterIndex%4, "canonical 32-byte base64url suffix")
	nonCanonicalAlias := raw[:len(raw)-1] + string(rawURLAlphabet[lastCharacterIndex+1])
	canonicalSignature, err := base64.RawURLEncoding.DecodeString(parts[3])
	require.NoError(t, err)
	aliasSignature, err := base64.RawURLEncoding.DecodeString(strings.Split(nonCanonicalAlias, ".")[3])
	require.NoError(t, err)
	require.Equal(t, canonicalSignature, aliasSignature, "test must exercise a non-canonical encoding of the same signature")

	for _, candidate := range []string{
		"",
		"wi1.v1.short.signature",
		"wi2.v1." + base64.RawURLEncoding.EncodeToString(make([]byte, 32)) + "." + base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
		tamperedSignature,
		nonCanonicalAlias,
	} {
		_, lookupErr := manager.Lookup(candidate)
		require.ErrorIs(t, lookupErr, ErrInvalidToken, candidate)
	}

	stored.Digest[0] ^= 0xff
	_, err = manager.Restore(stored)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestInvitationTokenLegacyCompatibilityIsShapeRestricted(t *testing.T) {
	t.Parallel()

	manager, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	})
	require.NoError(t, err)
	legacy := base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{7}, invitationTokenRandomBytes))

	lookup, err := manager.Lookup(legacy)
	require.NoError(t, err)
	require.Equal(t, legacy, lookup.LegacyToken)
	require.Empty(t, lookup.Digest)

	_, err = manager.Lookup(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, invitationTokenRandomBytes)))
	require.ErrorIs(t, err, ErrInvalidToken)

	legacyNonce := make([]byte, invitationTokenRandomBytes)
	legacyNonce[0], legacyNonce[1] = 0xc2, 0x20
	legacyStartingWithVersionLetters := base64.URLEncoding.EncodeToString(legacyNonce)
	require.True(t, bytes.HasPrefix([]byte(legacyStartingWithVersionLetters), []byte("wi")))
	lookup, err = manager.Lookup(legacyStartingWithVersionLetters)
	require.NoError(t, err)
	require.Equal(t, legacyStartingWithVersionLetters, lookup.LegacyToken)
}

func TestInvitationTokenManagerValidatesConfigurationAndEntropySource(t *testing.T) {
	t.Parallel()

	_, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: "short"},
	})
	require.ErrorContains(t, err, "at least 32 bytes")

	_, err = NewInvitationTokenManager(InvitationTokenConfig{
		Current:  InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
		Previous: []InvitationTokenKey{{ID: "v1", Secret: invitationTokenTestSecret + "-previous"}},
	})
	require.ErrorContains(t, err, "duplicated")

	_, err = NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "invalid.key", Secret: invitationTokenTestSecret},
	})
	require.ErrorContains(t, err, "key ID")

	manager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	}, errorReader{})
	require.NoError(t, err)
	_, _, err = manager.Issue()
	require.ErrorContains(t, err, "generate invitation token")
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }

func FuzzInvitationTokenLookupNeverPanics(f *testing.F) {
	manager, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	})
	if err != nil {
		f.Fatal(err)
	}
	f.Add("")
	f.Add("wi1.v1.invalid.invalid")
	f.Add(base64.URLEncoding.EncodeToString(make([]byte, 32)))

	f.Fuzz(func(t *testing.T, candidate string) {
		lookup, lookupErr := manager.Lookup(candidate)
		if lookupErr == nil {
			require.True(t, len(lookup.Digest) == 32 || lookup.LegacyToken != "")
		}
	})
}
