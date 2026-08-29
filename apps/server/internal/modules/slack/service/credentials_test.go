package slack

import (
	"bytes"
	"context"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCredentialCodecRoundTripAndContextBinding(t *testing.T) {
	t.Parallel()

	codec, err := newCredentialCodec(newTestCredentialVault(t))
	if err != nil {
		t.Fatalf("newCredentialCodec() error = %v", err)
	}
	binding := slackCredentialBinding{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
	}
	sealed, version, err := codec.seal(binding, slackCredential{AccessToken: "xoxb-secret"})
	if err != nil {
		t.Fatalf("seal() error = %v", err)
	}
	if version != credentialvault.CurrentVersion || sealed == "xoxb-secret" {
		t.Fatalf("seal() returned an invalid vault envelope")
	}
	credential, openedVersion, err := codec.open(binding, sealed)
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	if credential.AccessToken != "xoxb-secret" || openedVersion != version {
		t.Fatal("open() returned different credential data or version")
	}

	wrongBinding := binding
	wrongBinding.WorkspaceID = uuid.New()
	if _, _, err := codec.open(wrongBinding, sealed); err == nil {
		t.Fatal("open(wrong binding) error = nil, want authentication failure")
	}
}

func TestCredentialCodecRejectsLegacyOutsideBackfill(t *testing.T) {
	t.Parallel()

	codec, err := newCredentialCodec(newTestCredentialVault(t))
	if err != nil {
		t.Fatalf("newCredentialCodec() error = %v", err)
	}
	binding := slackCredentialBinding{WorkspaceID: uuid.New(), SlackTeamID: "T123", InstallGeneration: uuid.New()}
	if _, _, err := codec.open(binding, "xoxb-legacy"); err == nil {
		t.Fatal("open(legacy) error = nil, want fail-closed rejection")
	}
	cutover, err := NewLegacyCutover("test-secret")
	if err != nil {
		t.Fatalf("NewLegacyCutover() error = %v", err)
	}
	legacy, legacyVersion, err := cutover.openCredential("xoxb-legacy")
	if err != nil {
		t.Fatalf("openLegacy() error = %v", err)
	}
	if legacy.AccessToken != "xoxb-legacy" || legacyVersion != 0 {
		t.Fatal("openLegacy() returned different credential data or version")
	}
}

func TestSlackServiceVaultCredentialsDoNotRequireLegacyOrWebhookSecret(t *testing.T) {
	t.Parallel()
	service := New(nil, nil, nil, nil, Config{CredentialVault: newTestCredentialVault(t)})
	require.NotNil(t, service.credentials)

	binding := slackCredentialBinding{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T-KEY-SEPARATION",
		InstallGeneration: uuid.New(),
	}
	encrypted, version, err := service.credentials.seal(binding, slackCredential{AccessToken: "xoxb-isolated"})
	require.NoError(t, err)
	require.Equal(t, credentialvault.CurrentVersion, version)
	opened, openedVersion, err := service.credentials.open(binding, encrypted)
	require.NoError(t, err)
	require.Equal(t, "xoxb-isolated", opened.AccessToken)
	require.Equal(t, version, openedVersion)
}

func TestDecodeSlackCredentialDoesNotEchoInvalidPlaintext(t *testing.T) {
	t.Parallel()
	const sensitiveValue = "xoxb-must-not-enter-errors"
	_, err := decodeSlackCredential([]byte(`{"accessToken":"valid","expiresAt":"` + sensitiveValue + `"}`))
	if err == nil {
		t.Fatal("decodeSlackCredential(invalid expiry) error = nil")
	}
	if strings.Contains(err.Error(), sensitiveValue) {
		t.Fatal("decodeSlackCredential error exposed credential plaintext")
	}
}

func TestBotTokenRejectsLegacyCredentialOutsideBackfill(t *testing.T) {
	workspaceID := uuid.New()
	slackWorkspaceID := uuid.New()
	repo := &credentialUpgradeRepo{mockRepo: &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		ID:                slackWorkspaceID,
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    "xoxb-legacy",
		CredentialVersion: 0,
	}}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	repo.slackWorkspace.BotAccessToken = "xoxb-legacy"
	repo.slackWorkspace.CredentialVersion = 0

	token, err := service.botToken(context.Background(), repo.slackWorkspace)
	require.ErrorContains(t, err, "requires vault migration")
	require.Empty(t, token)
	require.Zero(t, repo.upgradeCalls)
}

func newTestCredentialVault(t testing.TB) *credentialvault.Vault {
	t.Helper()
	vault, err := buildTestCredentialVault()
	if err != nil {
		t.Fatalf("create test credential vault: %v", err)
	}
	return vault
}

func buildTestCredentialVault() (*credentialvault.Vault, error) {
	ref := credentialvault.KeyRef{ID: "test", Version: 1}
	return credentialvault.New(credentialvault.Config{
		Active: ref,
		Keys: []credentialvault.Key{{
			Ref:      ref,
			Material: bytes.Repeat([]byte{0x4f}, 32),
		}},
	})
}
