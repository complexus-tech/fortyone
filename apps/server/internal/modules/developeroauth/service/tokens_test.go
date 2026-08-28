package developeroauth

import (
	"bytes"
	"encoding/json"
	"testing"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOpaqueOAuthSecretsAreKeyedVersionedAndRedacted(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{0x41}, digestKeyBytes)
	manager, err := newTokenManager(TokenKeyringConfig{
		Active: developeroauthdomain.DigestKeyRef{ID: "current"},
		Keys:   []DigestKey{{Ref: developeroauthdomain.DigestKeyRef{ID: "current"}, Material: key}},
	}, bytes.NewReader(bytes.Repeat([]byte{0x5a}, 128)))
	require.NoError(t, err)

	issued, err := manager.Issue(developeroauthdomain.SecretAuthorizationCode, uuid.New())
	require.NoError(t, err)
	require.NotContains(t, issued.Plaintext.Reveal(), issued.Material.DigestKey.ID)
	require.Equal(t, "[REDACTED]", issued.Plaintext.String())
	encoded, err := json.Marshal(issued.Plaintext)
	require.NoError(t, err)
	require.JSONEq(t, `"[REDACTED]"`, string(encoded))

	prefix, err := manager.ParseLookupPrefix(issued.Plaintext.Reveal(), developeroauthdomain.SecretAuthorizationCode)
	require.NoError(t, err)
	require.Equal(t, issued.Material.LookupPrefix, prefix)
	require.NoError(t, manager.Verify(issued.Plaintext.Reveal(), issued.Material))
	require.ErrorIs(t, manager.Verify(issued.Plaintext.Reveal()+"changed", issued.Material), developeroauthdomain.ErrAuthorizationCode)
	require.ErrorIs(t, manager.Verify(issued.Plaintext.Reveal(), developeroauthdomain.SecretMaterial{
		ID: issued.Material.ID, Kind: developeroauthdomain.SecretRefreshToken,
		LookupPrefix: issued.Material.LookupPrefix, Digest: issued.Material.Digest, DigestKey: issued.Material.DigestKey,
	}), developeroauthdomain.ErrRefreshToken)
}

func TestOAuthDigestKeyRotationRetainsReadOldWriteNew(t *testing.T) {
	t.Parallel()
	oldRef := developeroauthdomain.DigestKeyRef{ID: "old"}
	newRef := developeroauthdomain.DigestKeyRef{ID: "new"}
	oldKey := bytes.Repeat([]byte{0x11}, digestKeyBytes)
	newKey := bytes.Repeat([]byte{0x22}, digestKeyBytes)
	oldManager, err := newTokenManager(TokenKeyringConfig{
		Active: oldRef,
		Keys:   []DigestKey{{Ref: oldRef, Material: oldKey}},
	}, bytes.NewReader(bytes.Repeat([]byte{0x33}, 128)))
	require.NoError(t, err)
	issuedBeforeRotation, err := oldManager.Issue(developeroauthdomain.SecretRefreshToken, uuid.New())
	require.NoError(t, err)

	rotatedManager, err := newTokenManager(TokenKeyringConfig{
		Active: newRef,
		Keys: []DigestKey{
			{Ref: oldRef, Material: oldKey},
			{Ref: newRef, Material: newKey},
		},
	}, bytes.NewReader(bytes.Repeat([]byte{0x44}, 128)))
	require.NoError(t, err)
	require.NoError(t, rotatedManager.Verify(issuedBeforeRotation.Plaintext.Reveal(), issuedBeforeRotation.Material))

	issuedAfterRotation, err := rotatedManager.Issue(developeroauthdomain.SecretRefreshToken, uuid.New())
	require.NoError(t, err)
	require.Equal(t, newRef, issuedAfterRotation.Material.DigestKey)
}

func TestOAuthClientSecretsUseAnIndependentVersionedFormat(t *testing.T) {
	t.Parallel()

	manager, err := newTokenManager(TokenKeyringConfig{
		Active: developeroauthdomain.DigestKeyRef{ID: "current"},
		Keys: []DigestKey{{
			Ref:      developeroauthdomain.DigestKeyRef{ID: "current"},
			Material: bytes.Repeat([]byte{0x71}, digestKeyBytes),
		}},
	}, bytes.NewReader(bytes.Repeat([]byte{0x2a}, 128)))
	require.NoError(t, err)

	issued, err := manager.Issue(developeroauthdomain.SecretClientSecret, uuid.New())
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix([]byte(issued.Plaintext.Reveal()), []byte(clientSecretHeader)))
	require.NotContains(t, issued.Plaintext.Reveal(), issued.Material.DigestKey.ID)
	require.NoError(t, manager.Verify(issued.Plaintext.Reveal(), issued.Material))

	prefix, err := manager.ParseLookupPrefix(issued.Plaintext.Reveal(), developeroauthdomain.SecretClientSecret)
	require.NoError(t, err)
	require.Equal(t, issued.Material.LookupPrefix, prefix)
	require.ErrorIs(t, manager.Verify(issued.Plaintext.Reveal()+"x", issued.Material), developeroauthdomain.ErrClientSecret)
	require.ErrorIs(t, manager.Verify(issued.Plaintext.Reveal(), developeroauthdomain.SecretMaterial{
		ID: issued.Material.ID, Kind: developeroauthdomain.SecretAuthorizationCode,
		LookupPrefix: issued.Material.LookupPrefix, Digest: issued.Material.Digest, DigestKey: issued.Material.DigestKey,
	}), developeroauthdomain.ErrAuthorizationCode)
}
