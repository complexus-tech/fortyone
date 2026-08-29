package slackrepository

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func TestCredentialPersistenceRejectsPlaintextBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	repository := &Repo{}
	if _, err := repository.UpsertSlackWorkspace(context.Background(), uuid.New(), uuid.New(), OAuthInstallPayload{
		SlackTeamID:       "T123",
		BotAccessToken:    "xoxb-plaintext",
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: uuid.New(),
	}); err == nil {
		t.Fatal("UpsertSlackWorkspace(plaintext) error = nil")
	}
	if _, err := repository.EnqueueSlackUninstall(context.Background(), SlackUninstallInput{
		WorkspaceID:          uuid.New(),
		SlackTeamID:          "T123",
		CredentialPayload:    "xoxb-plaintext",
		CredentialKeyVersion: credentialvault.CurrentVersion,
	}); err == nil {
		t.Fatal("EnqueueSlackUninstall(plaintext) error = nil")
	}
	if _, err := repository.RewrapSlackCredential(context.Background(), SlackCredentialRewrapRecord{
		SlackWorkspaceID:  uuid.New(),
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
		Credential:        "vault.v2.expected",
		CredentialVersion: credentialvault.CurrentVersion,
	}, "xoxb-plaintext"); err == nil {
		t.Fatal("RewrapSlackCredential(plaintext) error = nil")
	}
	if _, err := repository.RewrapSlackUninstallCredential(context.Background(), SlackUninstallCredentialRewrapRecord{
		UninstallID:       uuid.New(),
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
		Credential:        "vault.v2.expected",
		CredentialVersion: credentialvault.CurrentVersion,
	}, "xoxb-plaintext"); err == nil {
		t.Fatal("RewrapSlackUninstallCredential(plaintext) error = nil")
	}
}
