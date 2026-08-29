package credentialmigration

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"testing"
	"time"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type migrationStoreStub struct {
	records  []figmadomain.LegacyCredential
	upgraded map[uuid.UUID]string
}

func (store *migrationStoreStub) ListLegacyCredentials(
	context.Context,
	*uuid.UUID,
	int32,
) ([]figmadomain.LegacyCredential, error) {
	return append([]figmadomain.LegacyCredential(nil), store.records...), nil
}

func (store *migrationStoreStub) UpgradeLegacyCredential(
	_ context.Context,
	record figmadomain.LegacyCredential,
	nextPayload string,
) (bool, error) {
	store.upgraded[record.ID] = nextPayload
	return true, nil
}

func TestMigratorMovesLegacyCredentialIntoContextBoundVault(t *testing.T) {
	t.Parallel()

	const legacySecret = "legacy-provider-local-key"
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	token := figmadomain.Token{
		AccessToken: "figma-access-token", RefreshToken: "figma-refresh-token",
		TokenType: "Bearer", ExpiresAt: now.Add(time.Hour),
	}
	record := figmadomain.LegacyCredential{
		ID: uuid.New(), WorkspaceID: uuid.New(),
		InstallationGeneration: uuid.New(),
		Payload:                legacyEnvelope(t, legacySecret, token),
	}
	store := &migrationStoreStub{
		records:  []figmadomain.LegacyCredential{record},
		upgraded: make(map[uuid.UUID]string),
	}
	vault := migrationTestVault(t)
	migrator, err := New(store, vault, legacySecret)
	require.NoError(t, err)

	report, err := migrator.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, Report{Scanned: 1, Migrated: 1}, report)

	envelope := store.upgraded[record.ID]
	require.Contains(t, envelope, credentialvault.EnvelopePrefix)
	opened, err := vault.Open(
		figmaprovider.CredentialContext(
			record.WorkspaceID,
			record.ID,
			record.InstallationGeneration,
		),
		envelope,
	)
	require.NoError(t, err)
	defer opened.Destroy()
	var migrated figmadomain.Token
	require.NoError(t, json.Unmarshal(opened.Reveal(), &migrated))
	require.Equal(t, token, migrated)

	_, err = vault.Open(
		figmaprovider.CredentialContext(
			uuid.New(),
			record.ID,
			record.InstallationGeneration,
		),
		envelope,
	)
	require.ErrorIs(t, err, credentialvault.ErrAuthentication)
}

func TestOpenLegacyCredentialRejectsTrailingJSON(t *testing.T) {
	t.Parallel()

	const legacySecret = "legacy-provider-local-key"
	token := figmadomain.Token{
		AccessToken: "figma-access-token",
		ExpiresAt:   time.Date(2026, time.August, 28, 13, 0, 0, 0, time.UTC),
	}
	plaintext, err := json.Marshal(token)
	require.NoError(t, err)
	plaintext = append(plaintext, []byte(`{"unexpected":"document"}`)...)
	envelope := legacyRawEnvelope(t, legacySecret, plaintext)
	clear(plaintext)

	opened, err := openLegacyCredential(legacySecret, envelope)
	require.Error(t, err)
	require.Nil(t, opened)
}

func migrationTestVault(t testing.TB) *credentialvault.Vault {
	t.Helper()
	key := credentialvault.Key{
		Ref:      credentialvault.KeyRef{ID: "figma-migration-test", Version: 1},
		Material: []byte("0123456789abcdef0123456789abcdef"),
	}
	vault, err := credentialvault.New(credentialvault.Config{
		Active: key.Ref,
		Keys:   []credentialvault.Key{key},
	})
	require.NoError(t, err)
	return vault
}

func legacyEnvelope(t testing.TB, secret string, token figmadomain.Token) string {
	t.Helper()
	plaintext, err := json.Marshal(token)
	require.NoError(t, err)
	defer clear(plaintext)
	return legacyRawEnvelope(t, secret, plaintext)
}

func legacyRawEnvelope(t testing.TB, secret string, plaintext []byte) string {
	t.Helper()
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)
	sealed := append(nonce, gcm.Seal(nil, nonce, plaintext, nil)...)
	return base64.RawURLEncoding.EncodeToString(sealed)
}
