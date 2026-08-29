package slack

import (
	"bytes"
	"context"
	"errors"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type credentialRotationRepository struct {
	SlackEventRepository
	installations []slackrepository.SlackCredentialRewrapRecord
	uninstalls    []slackrepository.SlackUninstallCredentialRewrapRecord
	stale         bool
	writes        int
}

func (repository *credentialRotationRepository) ListSlackCredentialsForRewrap(
	_ context.Context,
	after *uuid.UUID,
	_ int,
) ([]slackrepository.SlackCredentialRewrapRecord, error) {
	if after != nil {
		return nil, nil
	}
	return append([]slackrepository.SlackCredentialRewrapRecord(nil), repository.installations...), nil
}

func (repository *credentialRotationRepository) RewrapSlackCredential(
	_ context.Context,
	record slackrepository.SlackCredentialRewrapRecord,
	rewrapped string,
) (bool, error) {
	if repository.stale || len(repository.installations) == 0 || repository.installations[0].Credential != record.Credential {
		return false, nil
	}
	repository.installations[0].Credential = rewrapped
	repository.writes++
	return true, nil
}

func (repository *credentialRotationRepository) ListSlackUninstallCredentialsForRewrap(
	_ context.Context,
	after *uuid.UUID,
	_ int,
) ([]slackrepository.SlackUninstallCredentialRewrapRecord, error) {
	if after != nil {
		return nil, nil
	}
	return append([]slackrepository.SlackUninstallCredentialRewrapRecord(nil), repository.uninstalls...), nil
}

func (repository *credentialRotationRepository) RewrapSlackUninstallCredential(
	_ context.Context,
	record slackrepository.SlackUninstallCredentialRewrapRecord,
	rewrapped string,
) (bool, error) {
	if repository.stale || len(repository.uninstalls) == 0 || repository.uninstalls[0].Credential != record.Credential {
		return false, nil
	}
	repository.uninstalls[0].Credential = rewrapped
	repository.writes++
	return true, nil
}

func TestRewrapCredentialsRotatesInstallAndUninstallIdempotently(t *testing.T) {
	t.Parallel()
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x31}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x42}, 32)}
	oldVault, err := credentialvault.New(credentialvault.Config{Active: oldRef, Keys: []credentialvault.Key{oldKey}})
	if err != nil {
		t.Fatalf("create old vault: %v", err)
	}
	rotatedVault, err := credentialvault.New(credentialvault.Config{Active: newRef, Keys: []credentialvault.Key{oldKey, newKey}})
	if err != nil {
		t.Fatalf("create rotated vault: %v", err)
	}
	oldCodec, err := newCredentialCodec(oldVault)
	if err != nil {
		t.Fatalf("create old codec: %v", err)
	}
	rotatedCodec, err := newCredentialCodec(rotatedVault)
	if err != nil {
		t.Fatalf("create rotated codec: %v", err)
	}

	workspaceID := uuid.New()
	generation := uuid.New()
	binding := slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: "T-ROTATE", InstallGeneration: generation}
	installationEnvelope, version, err := oldCodec.seal(binding, slackCredential{AccessToken: "xoxb-install"})
	if err != nil {
		t.Fatalf("seal installation credential: %v", err)
	}
	uninstallEnvelope, _, err := oldCodec.seal(binding, slackCredential{AccessToken: "xoxb-uninstall"})
	if err != nil {
		t.Fatalf("seal uninstall credential: %v", err)
	}
	repository := &credentialRotationRepository{
		installations: []slackrepository.SlackCredentialRewrapRecord{{
			SlackWorkspaceID:  uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       binding.SlackTeamID,
			InstallGeneration: generation,
			Credential:        installationEnvelope,
			CredentialVersion: version,
		}},
		uninstalls: []slackrepository.SlackUninstallCredentialRewrapRecord{{
			UninstallID:       uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       binding.SlackTeamID,
			InstallGeneration: generation,
			Credential:        uninstallEnvelope,
			CredentialVersion: version,
		}},
	}
	processor := &EventProcessor{repo: repository, codec: rotatedCodec}

	report, err := processor.RewrapCredentials(context.Background())
	if err != nil {
		t.Fatalf("RewrapCredentials() error = %v", err)
	}
	if report.ActiveKey != newRef || report.Scanned != 2 || report.Rewrapped != 2 || report.Current != 0 || report.Stale != 0 {
		t.Fatalf("RewrapCredentials() report = %#v", report)
	}
	if repository.writes != 2 {
		t.Fatalf("credential writes = %d, want 2", repository.writes)
	}
	for _, envelope := range []string{repository.installations[0].Credential, repository.uninstalls[0].Credential} {
		metadata, err := credentialvault.Inspect(envelope)
		if err != nil || metadata.Key != newRef {
			t.Fatalf("rewrapped metadata = (%#v, %v)", metadata, err)
		}
	}

	report, err = processor.RewrapCredentials(context.Background())
	if err != nil {
		t.Fatalf("second RewrapCredentials() error = %v", err)
	}
	if report.Scanned != 2 || report.Current != 2 || report.Rewrapped != 0 || report.Stale != 0 || repository.writes != 2 {
		t.Fatalf("second RewrapCredentials() = %#v, writes=%d", report, repository.writes)
	}
}

func TestRewrapCredentialsTreatsConcurrentRefreshAsStale(t *testing.T) {
	t.Parallel()
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x51}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x62}, 32)}
	oldVault, err := credentialvault.New(credentialvault.Config{Active: oldRef, Keys: []credentialvault.Key{oldKey}})
	if err != nil {
		t.Fatalf("create old vault: %v", err)
	}
	rotatedVault, err := credentialvault.New(credentialvault.Config{Active: newRef, Keys: []credentialvault.Key{oldKey, newKey}})
	if err != nil {
		t.Fatalf("create rotated vault: %v", err)
	}
	binding := slackCredentialBinding{WorkspaceID: uuid.New(), SlackTeamID: "T-STALE", InstallGeneration: uuid.New()}
	oldEnvelope, err := oldVault.Seal(binding.vaultContext(), []byte(`{"accessToken":"xoxb-old"}`))
	if err != nil {
		t.Fatalf("seal old credential: %v", err)
	}
	repository := &credentialRotationRepository{
		stale: true,
		installations: []slackrepository.SlackCredentialRewrapRecord{{
			SlackWorkspaceID:  uuid.New(),
			WorkspaceID:       binding.WorkspaceID,
			SlackTeamID:       binding.SlackTeamID,
			InstallGeneration: binding.InstallGeneration,
			Credential:        oldEnvelope,
			CredentialVersion: credentialvault.CurrentVersion,
		}},
	}
	codec, err := newCredentialCodec(rotatedVault)
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}
	processor := &EventProcessor{repo: repository, codec: codec}

	report, err := processor.RewrapCredentials(context.Background())
	if err != nil {
		t.Fatalf("RewrapCredentials() error = %v", err)
	}
	if report.Scanned != 1 || report.Stale != 1 || report.Rewrapped != 0 || repository.writes != 0 {
		t.Fatalf("RewrapCredentials() = %#v, writes=%d", report, repository.writes)
	}
	if repository.installations[0].Credential != oldEnvelope {
		t.Fatal("stale rewrap changed the concurrent credential")
	}
}

func TestRewrapCredentialsRejectsTamperingWithoutWriting(t *testing.T) {
	t.Parallel()
	vault := newTestCredentialVault(t)
	binding := slackCredentialBinding{WorkspaceID: uuid.New(), SlackTeamID: "T-TAMPER", InstallGeneration: uuid.New()}
	envelope, err := vault.Seal(binding.vaultContext(), []byte(`{"accessToken":"xoxb-token"}`))
	if err != nil {
		t.Fatalf("seal credential: %v", err)
	}
	tampered := envelope[:len(envelope)-1] + "A"
	if tampered == envelope {
		tampered = envelope[:len(envelope)-1] + "B"
	}
	repository := &credentialRotationRepository{installations: []slackrepository.SlackCredentialRewrapRecord{{
		SlackWorkspaceID:  uuid.New(),
		WorkspaceID:       binding.WorkspaceID,
		SlackTeamID:       binding.SlackTeamID,
		InstallGeneration: binding.InstallGeneration,
		Credential:        tampered,
		CredentialVersion: credentialvault.CurrentVersion,
	}}}
	codec, err := newCredentialCodec(vault)
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}
	processor := &EventProcessor{repo: repository, codec: codec}

	_, err = processor.RewrapCredentials(context.Background())
	if err == nil || (!errors.Is(err, credentialvault.ErrAuthentication) && !errors.Is(err, credentialvault.ErrMalformedEnvelope)) {
		t.Fatalf("RewrapCredentials(tampered) error = %v", err)
	}
	if repository.writes != 0 {
		t.Fatalf("tampered credential writes = %d", repository.writes)
	}
}
