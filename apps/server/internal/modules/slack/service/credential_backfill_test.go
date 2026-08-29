package slack

import (
	"context"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type credentialBackfillRepositoryStub struct {
	*eventRepositoryStub
	legacyUninstalls  []slackrepository.LegacySlackUninstallCredentialRecord
	uninstallUpgrades int
	uninstallEnvelope string
}

func (r *credentialBackfillRepositoryStub) ListLegacySlackUninstallCredentials(_ context.Context, _ int) ([]slackrepository.LegacySlackUninstallCredentialRecord, error) {
	return append([]slackrepository.LegacySlackUninstallCredentialRecord(nil), r.legacyUninstalls...), nil
}

func (r *credentialBackfillRepositoryStub) UpgradeSlackUninstallCredential(
	_ context.Context,
	_ slackrepository.LegacySlackUninstallCredentialRecord,
	encrypted string,
	_ int,
) error {
	r.uninstallUpgrades++
	r.uninstallEnvelope = encrypted
	if len(r.legacyUninstalls) > 0 {
		r.legacyUninstalls = r.legacyUninstalls[1:]
	}
	return nil
}

func TestEventProcessorBackfillsLegacyCredentials(t *testing.T) {
	baseRepo := newEventRepositoryStub()
	repo := &credentialBackfillRepositoryStub{eventRepositoryStub: baseRepo}
	repo.legacyCredentials = []slackrepository.LegacySlackCredentialRecord{
		{SlackWorkspaceID: uuid.New(), WorkspaceID: testWorkspaceID, SlackTeamID: "T1", InstallGeneration: testInstallGeneration, Credential: "xoxb-legacy-one"},
		{SlackWorkspaceID: uuid.New(), WorkspaceID: testWorkspaceID, SlackTeamID: "T1", InstallGeneration: testInstallGeneration, Credential: "xoxb-legacy-two"},
	}
	processor := newTestEventProcessor(t, baseRepo, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.repo = repo
	cutover := mustTestLegacyCutover(t)

	upgraded, err := processor.BackfillLegacyCredentials(context.Background(), cutover)
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if upgraded != 2 || repo.credentialUpgrades != 2 {
		t.Fatalf("credential upgrades = %d/%d, want 2/2", upgraded, repo.credentialUpgrades)
	}
	if len(repo.legacyCredentials) != 0 {
		t.Fatalf("legacy credentials remaining = %d, want 0", len(repo.legacyCredentials))
	}
	if !credentialvault.IsEnvelope(repo.installation.BotAccessToken) {
		t.Fatal("upgraded credential is not a versioned ciphertext")
	}
}

func TestEventProcessorScrubsLegacyColumnsAfterVersionedCredentialRollout(t *testing.T) {
	baseRepo := newEventRepositoryStub()
	repo := &credentialBackfillRepositoryStub{eventRepositoryStub: baseRepo}
	repo.versionedLegacyCredentials = 2
	processor := newTestEventProcessor(t, baseRepo, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.repo = repo
	cutover := mustTestLegacyCutover(t)

	updated, err := processor.BackfillLegacyCredentials(context.Background(), cutover)
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if updated != 2 || repo.versionedLegacyCredentials != 0 {
		t.Fatalf("scrubbed/remaining = %d/%d, want 2/0", updated, repo.versionedLegacyCredentials)
	}
}

func TestEventProcessorBackfillsLegacyUninstallCredentials(t *testing.T) {
	baseRepo := newEventRepositoryStub()
	repo := &credentialBackfillRepositoryStub{eventRepositoryStub: baseRepo}
	processor := newTestEventProcessor(t, baseRepo, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	processor.repo = repo
	cutover := mustTestLegacyCutover(t)
	legacyEnvelope, err := cutover.box.Seal([]byte(`{"accessToken":"xoxb-legacy-uninstall"}`))
	if err != nil {
		t.Fatalf("seal legacy uninstall fixture: %v", err)
	}
	repo.legacyUninstalls = []slackrepository.LegacySlackUninstallCredentialRecord{{
		UninstallID:       uuid.New(),
		WorkspaceID:       testWorkspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: testInstallGeneration,
		Credential:        legacyEnvelope,
		CredentialVersion: 1,
	}}

	updated, err := processor.BackfillLegacyCredentials(context.Background(), cutover)
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if updated != 1 || repo.uninstallUpgrades != 1 || !credentialvault.IsEnvelope(repo.uninstallEnvelope) {
		t.Fatalf("uninstall credential upgrades = %d/%d envelope=%t", updated, repo.uninstallUpgrades, credentialvault.IsEnvelope(repo.uninstallEnvelope))
	}
}
