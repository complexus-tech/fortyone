//go:build integration

package slack

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSlackCredentialRewrapCoversMixedGenerationsAndRollback(t *testing.T) {
	database := testkit.NewPostgres(t)
	db := database.Pool

	ctx := t.Context()
	repository := slackrepository.New(db)
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x31}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x42}, 32)}
	oldVault := newSlackRotationVault(t, oldRef, oldKey)
	rotatedVault := newSlackRotationVault(t, newRef, oldKey, newKey)
	oldCodec := newSlackRotationCodec(t, oldVault)
	rotatedCodec := newSlackRotationCodec(t, rotatedVault)

	oldUserID, oldWorkspaceID := seedSlackCredentialOwner(t, db)
	oldBinding := slackCredentialBinding{
		WorkspaceID:       oldWorkspaceID,
		SlackTeamID:       "T-ROTATION-OLD",
		InstallGeneration: uuid.New(),
	}
	oldEnvelope := sealSlackRotationCredential(t, oldCodec, oldBinding, "xoxb-install-old")
	oldInstallation, err := repository.UpsertSlackWorkspace(ctx, oldWorkspaceID, oldUserID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       oldBinding.SlackTeamID,
		SlackTeamName:     "Rotation old",
		SlackTeamDomain:   "rotation-old",
		BotAccessToken:    oldEnvelope,
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: oldBinding.InstallGeneration,
	})
	if err != nil {
		t.Fatalf("persist old Slack installation credential: %v", err)
	}

	currentUserID, currentWorkspaceID := seedSlackCredentialOwner(t, db)
	currentBinding := slackCredentialBinding{
		WorkspaceID:       currentWorkspaceID,
		SlackTeamID:       "T-ROTATION-CURRENT",
		InstallGeneration: uuid.New(),
	}
	currentEnvelope := sealSlackRotationCredential(t, rotatedCodec, currentBinding, "xoxb-install-current")
	if _, err := repository.UpsertSlackWorkspace(ctx, currentWorkspaceID, currentUserID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       currentBinding.SlackTeamID,
		SlackTeamName:     "Rotation current",
		SlackTeamDomain:   "rotation-current",
		BotAccessToken:    currentEnvelope,
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: currentBinding.InstallGeneration,
	}); err != nil {
		t.Fatalf("persist current Slack installation credential: %v", err)
	}

	_, uninstallWorkspaceID := seedSlackCredentialOwner(t, db)
	uninstallBinding := slackCredentialBinding{
		WorkspaceID:       uninstallWorkspaceID,
		SlackTeamID:       "T-ROTATION-UNINSTALL",
		InstallGeneration: uuid.New(),
	}
	uninstallEnvelope := sealSlackRotationCredential(t, oldCodec, uninstallBinding, "xoxb-uninstall-old")
	uninstall, err := repository.EnqueueSlackUninstall(ctx, slackrepository.SlackUninstallInput{
		SlackWorkspaceID:     uuid.New(),
		WorkspaceID:          uninstallWorkspaceID,
		InstallGeneration:    uninstallBinding.InstallGeneration,
		SlackTeamID:          uninstallBinding.SlackTeamID,
		CredentialPayload:    uninstallEnvelope,
		CredentialKeyVersion: credentialvault.CurrentVersion,
	})
	if err != nil {
		t.Fatalf("persist old Slack uninstall credential: %v", err)
	}

	processor := &EventProcessor{repo: repository, codec: rotatedCodec}
	report, err := processor.RewrapCredentials(ctx)
	if err != nil {
		t.Fatalf("RewrapCredentials() error = %v", err)
	}
	if report.ActiveKey != newRef || report.Scanned != 3 || report.Rewrapped != 2 || report.Current != 1 || report.Stale != 0 {
		t.Fatalf("RewrapCredentials() = %#v", report)
	}

	oldStored, err := repository.GetSlackWorkspace(ctx, oldWorkspaceID)
	if err != nil {
		t.Fatalf("read rewrapped Slack installation: %v", err)
	}
	if oldStored.ID != oldInstallation.ID || oldStored.InstallGeneration != oldBinding.InstallGeneration {
		t.Fatal("rewrap changed Slack installation identity or generation")
	}
	assertSlackRotationEnvelope(t, oldStored.BotAccessToken, newRef)
	opened, _, err := rotatedCodec.open(oldBinding, oldStored.BotAccessToken)
	if err != nil {
		t.Fatalf("open rewrapped Slack installation: %v", err)
	}
	if opened.AccessToken != "xoxb-install-old" {
		t.Fatal("rewrapped Slack installation credential changed")
	}
	uninstallStored := readSlackUninstallCredential(t, db, uninstall.ID)
	if uninstallStored.Generation != uninstallBinding.InstallGeneration {
		t.Fatal("rewrap changed Slack uninstall generation")
	}
	assertSlackRotationEnvelope(t, uninstallStored.Payload, newRef)
	opened, _, err = rotatedCodec.open(uninstallBinding, uninstallStored.Payload)
	if err != nil {
		t.Fatalf("open rewrapped Slack uninstall: %v", err)
	}
	if opened.AccessToken != "xoxb-uninstall-old" {
		t.Fatal("rewrapped Slack uninstall credential changed")
	}

	second, err := processor.RewrapCredentials(ctx)
	if err != nil || second.Scanned != 3 || second.Current != 3 || second.Rewrapped != 0 || second.Stale != 0 {
		t.Fatalf("second RewrapCredentials() = (%#v, %v)", second, err)
	}

	// A deployment rollback can make the retained previous key active again.
	// The same bounded operation safely moves every DEK wrapping back without
	// changing provider generations or credential ciphertext.
	rollbackVault := newSlackRotationVault(t, oldRef, oldKey, newKey)
	rollbackCodec := newSlackRotationCodec(t, rollbackVault)
	rollbackProcessor := &EventProcessor{repo: repository, codec: rollbackCodec}
	rollback, err := rollbackProcessor.RewrapCredentials(ctx)
	if err != nil {
		t.Fatalf("rollback RewrapCredentials() error = %v", err)
	}
	if rollback.ActiveKey != oldRef || rollback.Scanned != 3 || rollback.Rewrapped != 3 || rollback.Current != 0 || rollback.Stale != 0 {
		t.Fatalf("rollback RewrapCredentials() = %#v", rollback)
	}
	oldStored, err = repository.GetSlackWorkspace(ctx, oldWorkspaceID)
	if err != nil {
		t.Fatalf("read rolled-back Slack installation: %v", err)
	}
	assertSlackRotationEnvelope(t, oldStored.BotAccessToken, oldRef)
	uninstallStored = readSlackUninstallCredential(t, db, uninstall.ID)
	assertSlackRotationEnvelope(t, uninstallStored.Payload, oldRef)
	rollbackSecond, err := rollbackProcessor.RewrapCredentials(ctx)
	if err != nil || rollbackSecond.Scanned != 3 || rollbackSecond.Current != 3 || rollbackSecond.Rewrapped != 0 {
		t.Fatalf("second rollback RewrapCredentials() = (%#v, %v)", rollbackSecond, err)
	}
}

func TestSlackCredentialRewrapFencesConcurrentRefreshAndRevoke(t *testing.T) {
	database := testkit.NewPostgres(t)
	db := database.Pool

	ctx := t.Context()
	repository := slackrepository.New(db)
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x51}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x62}, 32)}
	oldVault := newSlackRotationVault(t, oldRef, oldKey)
	rotatedVault := newSlackRotationVault(t, newRef, oldKey, newKey)
	oldCodec := newSlackRotationCodec(t, oldVault)
	rotatedCodec := newSlackRotationCodec(t, rotatedVault)
	userID, workspaceID := seedSlackCredentialOwner(t, db)
	oldBinding := slackCredentialBinding{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T-ROTATION-RACE",
		InstallGeneration: uuid.New(),
	}
	oldEnvelope := sealSlackRotationCredential(t, oldCodec, oldBinding, "xoxb-race-old")
	if _, err := repository.UpsertSlackWorkspace(ctx, workspaceID, userID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       oldBinding.SlackTeamID,
		SlackTeamName:     "Rotation race",
		SlackTeamDomain:   "rotation-race",
		BotAccessToken:    oldEnvelope,
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: oldBinding.InstallGeneration,
	}); err != nil {
		t.Fatalf("persist race Slack credential: %v", err)
	}
	candidates, err := repository.ListSlackCredentialsForRewrap(ctx, nil, credentialvault.MaxMaintenanceBatchSize)
	if err != nil || len(candidates) != 1 {
		t.Fatalf("ListSlackCredentialsForRewrap() = (%d, %v), want one", len(candidates), err)
	}
	candidate := candidates[0]
	staleResult, err := rotatedVault.Rewrap(oldBinding.vaultContext(), candidate.Credential)
	if err != nil || !staleResult.Changed {
		t.Fatalf("prepare stale Slack rewrap: changed=%t err=%v", staleResult.Changed, err)
	}

	refreshedBinding := oldBinding
	refreshedBinding.InstallGeneration = uuid.New()
	refreshedEnvelope := sealSlackRotationCredential(t, rotatedCodec, refreshedBinding, "xoxb-race-new")
	start := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(2)
	var refreshErr, rewrapErr error
	var replaced bool
	go func() {
		defer wait.Done()
		<-start
		_, refreshErr = repository.UpsertSlackWorkspace(ctx, workspaceID, userID, slackrepository.OAuthInstallPayload{
			SlackTeamID:       refreshedBinding.SlackTeamID,
			SlackTeamName:     "Rotation race refreshed",
			SlackTeamDomain:   "rotation-race",
			BotAccessToken:    refreshedEnvelope,
			CredentialVersion: credentialvault.CurrentVersion,
			InstallGeneration: refreshedBinding.InstallGeneration,
		})
	}()
	go func() {
		defer wait.Done()
		<-start
		replaced, rewrapErr = repository.RewrapSlackCredential(ctx, candidate, staleResult.Envelope)
	}()
	close(start)
	wait.Wait()
	if refreshErr != nil || rewrapErr != nil {
		t.Fatalf("concurrent Slack refresh/rewrap errors = (%v, %v)", refreshErr, rewrapErr)
	}
	_ = replaced // Either order is valid; the refreshed generation must win.
	final, err := repository.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		t.Fatalf("read concurrently refreshed Slack credential: %v", err)
	}
	if final.InstallGeneration != refreshedBinding.InstallGeneration || final.BotAccessToken != refreshedEnvelope {
		t.Fatal("credential rewrap overwrote a concurrent Slack refresh")
	}
	opened, _, err := rotatedCodec.open(refreshedBinding, final.BotAccessToken)
	if err != nil {
		t.Fatalf("open concurrently refreshed Slack credential: %v", err)
	}
	if opened.AccessToken != "xoxb-race-new" {
		t.Fatal("concurrently refreshed Slack credential changed")
	}

	if err := repository.DeactivateSlackWorkspaceByTeamID(ctx, final.SlackTeamID, final.InstallGeneration); err != nil {
		t.Fatalf("revoke Slack installation: %v", err)
	}
	revokedCandidate := slackrepository.SlackCredentialRewrapRecord{
		SlackWorkspaceID:  final.ID,
		WorkspaceID:       final.WorkspaceID,
		SlackTeamID:       final.SlackTeamID,
		InstallGeneration: final.InstallGeneration,
		Credential:        final.BotAccessToken,
		CredentialVersion: final.CredentialVersion,
	}
	replaced, err = repository.RewrapSlackCredential(ctx, revokedCandidate, final.BotAccessToken)
	if err != nil || replaced {
		t.Fatalf("rewrap after Slack revoke = (%v, %v), want false", replaced, err)
	}
	if _, err := repository.GetSlackWorkspace(ctx, workspaceID); !errors.Is(err, slackdomain.ErrNotFound) {
		t.Fatalf("revoked Slack credential lookup error = %v, want ErrNotFound", err)
	}
}

func TestSlackCredentialRewrapRejectsUnknownAndTamperedCiphertextWithoutPersistence(t *testing.T) {
	database := testkit.NewPostgres(t)
	db := database.Pool

	ctx := t.Context()
	repository := slackrepository.New(db)
	oldRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 1}
	newRef := credentialvault.KeyRef{ID: "provider-credentials", Version: 2}
	oldKey := credentialvault.Key{Ref: oldRef, Material: bytes.Repeat([]byte{0x71}, 32)}
	newKey := credentialvault.Key{Ref: newRef, Material: bytes.Repeat([]byte{0x72}, 32)}
	oldVault := newSlackRotationVault(t, oldRef, oldKey)
	newOnlyVault := newSlackRotationVault(t, newRef, newKey)
	rotatedVault := newSlackRotationVault(t, newRef, oldKey, newKey)
	oldCodec := newSlackRotationCodec(t, oldVault)
	newOnlyCodec := newSlackRotationCodec(t, newOnlyVault)
	rotatedCodec := newSlackRotationCodec(t, rotatedVault)

	unknownUserID, unknownWorkspaceID := seedSlackCredentialOwner(t, db)
	unknownBinding := slackCredentialBinding{
		WorkspaceID:       unknownWorkspaceID,
		SlackTeamID:       "T-ROTATION-UNKNOWN",
		InstallGeneration: uuid.New(),
	}
	unknownEnvelope := sealSlackRotationCredential(t, oldCodec, unknownBinding, "xoxb-unknown-key")
	if _, err := repository.UpsertSlackWorkspace(ctx, unknownWorkspaceID, unknownUserID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       unknownBinding.SlackTeamID,
		SlackTeamName:     "Rotation unknown",
		SlackTeamDomain:   "rotation-unknown",
		BotAccessToken:    unknownEnvelope,
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: unknownBinding.InstallGeneration,
	}); err != nil {
		t.Fatalf("persist unknown-key Slack credential: %v", err)
	}
	unknownProcessor := &EventProcessor{repo: repository, codec: newOnlyCodec}
	_, err := unknownProcessor.RewrapCredentials(ctx)
	if !errors.Is(err, credentialvault.ErrUnknownKey) {
		t.Fatalf("RewrapCredentials(unknown key) error = %v, want ErrUnknownKey", err)
	}
	if strings.Contains(err.Error(), "xoxb-unknown-key") || strings.Contains(err.Error(), unknownEnvelope) {
		t.Fatal("unknown-key rewrap error exposed credential material")
	}
	unknownStored, err := repository.GetSlackWorkspace(ctx, unknownWorkspaceID)
	if err != nil || unknownStored.BotAccessToken != unknownEnvelope {
		t.Fatal("unknown-key rewrap changed the persisted credential")
	}
	if err := repository.DeactivateSlackWorkspaceByTeamID(ctx, unknownBinding.SlackTeamID, unknownBinding.InstallGeneration); err != nil {
		t.Fatalf("remove unknown-key fixture: %v", err)
	}

	tamperedUserID, tamperedWorkspaceID := seedSlackCredentialOwner(t, db)
	tamperedBinding := slackCredentialBinding{
		WorkspaceID:       tamperedWorkspaceID,
		SlackTeamID:       "T-ROTATION-TAMPERED",
		InstallGeneration: uuid.New(),
	}
	tamperedEnvelope := sealSlackRotationCredential(t, oldCodec, tamperedBinding, "xoxb-tampered")
	replacement := "A"
	if strings.HasSuffix(tamperedEnvelope, replacement) {
		replacement = "B"
	}
	tamperedEnvelope = tamperedEnvelope[:len(tamperedEnvelope)-1] + replacement
	if err := persistSlackRotationCredential(t, repository, tamperedUserID, tamperedBinding, tamperedEnvelope); err != nil {
		t.Fatalf("persist tampered Slack credential: %v", err)
	}
	tamperedProcessor := &EventProcessor{repo: repository, codec: rotatedCodec}
	_, err = tamperedProcessor.RewrapCredentials(ctx)
	if err == nil || (!errors.Is(err, credentialvault.ErrAuthentication) && !errors.Is(err, credentialvault.ErrMalformedEnvelope)) {
		t.Fatalf("RewrapCredentials(tampered) error = %v", err)
	}
	if strings.Contains(err.Error(), "xoxb-tampered") || strings.Contains(err.Error(), tamperedEnvelope) {
		t.Fatal("tampered rewrap error exposed credential material")
	}
	tamperedStored, readErr := repository.GetSlackWorkspace(ctx, tamperedWorkspaceID)
	if readErr != nil || tamperedStored.BotAccessToken != tamperedEnvelope {
		t.Fatal("tampered rewrap changed the persisted credential")
	}
}

type slackRotationUninstallCredential struct {
	Payload    string
	Generation uuid.UUID
}

func readSlackUninstallCredential(t testing.TB, db *pgxpool.Pool, uninstallID uuid.UUID) slackRotationUninstallCredential {
	t.Helper()
	var stored slackRotationUninstallCredential
	if err := db.QueryRow(t.Context(), `
		SELECT credential_payload, installation_generation
		FROM slack_uninstall_outbox
		WHERE id = $1
	`, uninstallID).Scan(&stored.Payload, &stored.Generation); err != nil {
		t.Fatalf("read Slack uninstall credential: %v", err)
	}
	return stored
}

func newSlackRotationVault(t testing.TB, active credentialvault.KeyRef, keys ...credentialvault.Key) *credentialvault.Vault {
	t.Helper()
	vault, err := credentialvault.New(credentialvault.Config{Active: active, Keys: keys})
	if err != nil {
		t.Fatalf("create Slack rotation vault: %v", err)
	}
	return vault
}

func newSlackRotationCodec(t testing.TB, vault *credentialvault.Vault) *credentialCodec {
	t.Helper()
	codec, err := newCredentialCodec(vault)
	if err != nil {
		t.Fatalf("create Slack rotation credential codec: %v", err)
	}
	return codec
}

func sealSlackRotationCredential(t testing.TB, codec *credentialCodec, binding slackCredentialBinding, token string) string {
	t.Helper()
	envelope, version, err := codec.seal(binding, slackCredential{AccessToken: token})
	if err != nil {
		t.Fatalf("seal Slack rotation credential: %v", err)
	}
	if version != credentialvault.CurrentVersion {
		t.Fatalf("sealed Slack rotation credential version = %d", version)
	}
	return envelope
}

func persistSlackRotationCredential(
	t testing.TB,
	repository *slackrepository.Repo,
	userID uuid.UUID,
	binding slackCredentialBinding,
	envelope string,
) error {
	t.Helper()
	_, err := repository.UpsertSlackWorkspace(t.Context(), binding.WorkspaceID, userID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       binding.SlackTeamID,
		SlackTeamName:     "Rotation failure",
		SlackTeamDomain:   "rotation-failure",
		BotAccessToken:    envelope,
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: binding.InstallGeneration,
	})
	return err
}

func assertSlackRotationEnvelope(t testing.TB, envelope string, want credentialvault.KeyRef) {
	t.Helper()
	metadata, err := credentialvault.Inspect(envelope)
	if err != nil {
		t.Fatalf("inspect Slack rotation envelope: %v", err)
	}
	if metadata.Key != want {
		t.Fatalf("Slack rotation envelope key = %#v, want %#v", metadata.Key, want)
	}
}
