//go:build integration

package slack

import (
	"errors"
	"strings"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSlackCredentialBackfillEncryptsScrubsAndBindsLegacyTokens(t *testing.T) {
	database := testkit.NewPostgres(t)
	db := database.Pool

	ctx := t.Context()
	userID, workspaceID := seedSlackCredentialOwner(t, db)
	installationID := uuid.New()
	generation := uuid.New()
	const (
		slackTeamID = "T-VAULT-MIGRATION"
		legacyToken = "xoxb-legacy-token-must-not-survive"
	)
	if _, err := db.Exec(ctx, `
		INSERT INTO slack_workspaces (
			id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_access_token, credential_payload, credential_key_version,
			installation_generation, installed_by_user_id
		) VALUES ($1, $2, $3, 'Vault migration', 'vault-migration', $4, NULL, 0, $5, $6)
	`, installationID, workspaceID, slackTeamID, legacyToken, generation, userID); err != nil {
		t.Fatalf("insert legacy Slack credential: %v", err)
	}

	vault := newTestCredentialVault(t)
	codec, err := newCredentialCodec(vault)
	if err != nil {
		t.Fatalf("construct Slack credential codec: %v", err)
	}
	cutover, err := NewLegacyCutover("slack-legacy-migration-secret")
	if err != nil {
		t.Fatalf("construct Slack legacy cutover: %v", err)
	}
	uninstallID := uuid.New()
	legacyUninstallEnvelope, err := cutover.box.Seal([]byte(`{"accessToken":"xoxb-legacy-uninstall"}`))
	if err != nil {
		t.Fatalf("seal legacy Slack uninstall credential: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO slack_uninstall_outbox (
			id, slack_workspace_id, workspace_id, installation_generation,
			slack_team_id, credential_payload, credential_key_version,
			status, next_attempt_at
		) VALUES ($1, $2, $3, $4, $5, $6, 1, 'pending', NOW())
	`, uninstallID, installationID, workspaceID, generation, slackTeamID, legacyUninstallEnvelope); err != nil {
		t.Fatalf("insert legacy Slack uninstall credential: %v", err)
	}
	repository := slackrepository.New(db)
	processor := &EventProcessor{repo: repository, codec: codec}

	updated, err := processor.BackfillLegacyCredentials(ctx, cutover)
	if err != nil {
		t.Fatalf("BackfillLegacyCredentials() error = %v", err)
	}
	if updated != 2 {
		t.Fatalf("BackfillLegacyCredentials() = %d, want 2", updated)
	}

	var stored, legacyColumn string
	var version int
	var storedGeneration uuid.UUID
	if err := db.QueryRow(ctx, `
		SELECT credential_payload, bot_access_token, credential_key_version, installation_generation
		FROM slack_workspaces WHERE id = $1
	`, installationID).Scan(&stored, &legacyColumn, &version, &storedGeneration); err != nil {
		t.Fatalf("read migrated Slack credential: %v", err)
	}
	if !credentialvault.IsEnvelope(stored) || strings.Contains(stored, legacyToken) || legacyColumn != "" {
		t.Fatal("Slack backfill did not atomically replace and scrub the plaintext credential")
	}
	if version != credentialvault.CurrentVersion || storedGeneration != generation {
		t.Fatalf("migrated Slack metadata = version %d, generation %s", version, storedGeneration)
	}
	binding := slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: slackTeamID, InstallGeneration: generation}
	opened, openedVersion, err := codec.open(binding, stored)
	if err != nil {
		t.Fatalf("open migrated Slack credential: %v", err)
	}
	if opened.AccessToken != legacyToken || openedVersion != credentialvault.CurrentVersion {
		t.Fatalf("migrated Slack credential has wrong version %d or changed data", openedVersion)
	}
	wrongBinding := binding
	wrongBinding.SlackTeamID = "T-OTHER"
	if _, _, err := codec.open(wrongBinding, stored); err == nil {
		t.Fatal("open migrated Slack credential under another team error = nil")
	}
	var uninstallPayload string
	var uninstallVersion int
	if err := db.QueryRow(ctx, `
		SELECT credential_payload, credential_key_version
		FROM slack_uninstall_outbox WHERE id = $1
	`, uninstallID).Scan(&uninstallPayload, &uninstallVersion); err != nil {
		t.Fatalf("read migrated Slack uninstall credential: %v", err)
	}
	if !credentialvault.IsEnvelope(uninstallPayload) || uninstallVersion != credentialvault.CurrentVersion {
		t.Fatal("Slack uninstall credential was not migrated to the shared vault")
	}
	uninstallCredential, _, err := codec.open(binding, uninstallPayload)
	if err != nil || uninstallCredential.AccessToken != "xoxb-legacy-uninstall" {
		t.Fatalf("open migrated Slack uninstall credential returned different data: err=%v", err)
	}
	if updated, err := processor.BackfillLegacyCredentials(ctx, cutover); err != nil || updated != 0 {
		t.Fatalf("second BackfillLegacyCredentials() = (%d, %v), want no-op", updated, err)
	}

	if _, err := db.Exec(ctx, `UPDATE slack_workspaces SET bot_access_token = 'residual-plaintext' WHERE id = $1`, installationID); err != nil {
		t.Fatalf("seed versioned residual Slack plaintext: %v", err)
	}
	if updated, err := processor.BackfillLegacyCredentials(ctx, cutover); err != nil || updated != 1 {
		t.Fatalf("credential scrub BackfillLegacyCredentials() = (%d, %v), want 1", updated, err)
	}
	if err := db.QueryRow(ctx, `SELECT bot_access_token FROM slack_workspaces WHERE id = $1`, installationID).Scan(&legacyColumn); err != nil {
		t.Fatalf("read scrubbed Slack legacy column: %v", err)
	}
	if legacyColumn != "" {
		t.Fatal("versioned Slack credential retained legacy plaintext")
	}

	if _, err := repository.UpsertSlackWorkspace(ctx, workspaceID, userID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       slackTeamID,
		SlackTeamName:     "Vault migration",
		SlackTeamDomain:   "vault-migration",
		BotAccessToken:    "xoxb-new-plaintext",
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: uuid.New(),
	}); err == nil {
		t.Fatal("UpsertSlackWorkspace(plaintext) error = nil")
	}
	newGeneration := uuid.New()
	newBinding := slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: slackTeamID, InstallGeneration: newGeneration}
	newEnvelope, newVersion, err := codec.seal(newBinding, slackCredential{AccessToken: "xoxb-rotated"})
	if err != nil {
		t.Fatalf("seal rotated Slack credential: %v", err)
	}
	if _, err := repository.UpsertSlackWorkspace(ctx, workspaceID, userID, slackrepository.OAuthInstallPayload{
		SlackTeamID:       slackTeamID,
		SlackTeamName:     "Vault migration",
		SlackTeamDomain:   "vault-migration",
		BotAccessToken:    newEnvelope,
		CredentialVersion: newVersion,
		InstallGeneration: newGeneration,
	}); err != nil {
		t.Fatalf("UpsertSlackWorkspace(vault envelope) error = %v", err)
	}
	if err := db.QueryRow(ctx, `
		SELECT credential_payload, bot_access_token, credential_key_version, installation_generation
		FROM slack_workspaces WHERE id = $1
	`, installationID).Scan(&stored, &legacyColumn, &version, &storedGeneration); err != nil {
		t.Fatalf("read rotated Slack credential: %v", err)
	}
	if stored != newEnvelope || legacyColumn != "" || version != newVersion || storedGeneration != newGeneration {
		t.Fatal("new Slack write did not persist only the context-bound vault envelope")
	}
}

func TestSlackCredentialUpgradeCompareAndSwapDoesNotOverwriteConcurrentRotation(t *testing.T) {
	database := testkit.NewPostgres(t)
	db := database.Pool

	ctx := t.Context()
	userID, workspaceID := seedSlackCredentialOwner(t, db)
	installationID := uuid.New()
	generation := uuid.New()
	if _, err := db.Exec(ctx, `
		INSERT INTO slack_workspaces (
			id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_access_token, credential_payload, credential_key_version,
			installation_generation, installed_by_user_id
		) VALUES ($1, $2, 'T-CAS', 'CAS', 'cas', 'old-plaintext', NULL, 0, $3, $4)
	`, installationID, workspaceID, generation, userID); err != nil {
		t.Fatalf("insert legacy Slack credential: %v", err)
	}
	repository := slackrepository.New(db)
	records, err := repository.ListLegacySlackCredentials(ctx, 10)
	if err != nil || len(records) != 1 {
		t.Fatalf("ListLegacySlackCredentials() = (%d, %v), want one", len(records), err)
	}
	if _, err := db.Exec(ctx, `UPDATE slack_workspaces SET bot_access_token = 'concurrent-plaintext' WHERE id = $1`, installationID); err != nil {
		t.Fatalf("simulate concurrent Slack credential rotation: %v", err)
	}
	codec, err := newCredentialCodec(newTestCredentialVault(t))
	if err != nil {
		t.Fatalf("construct Slack credential codec: %v", err)
	}
	encrypted, version, err := codec.seal(
		slackCredentialBinding{WorkspaceID: workspaceID, SlackTeamID: "T-CAS", InstallGeneration: generation},
		slackCredential{AccessToken: "old-plaintext"},
	)
	if err != nil {
		t.Fatalf("seal Slack credential: %v", err)
	}
	if err := repository.UpgradeSlackCredential(ctx, records[0], encrypted, version); !errors.Is(err, slackdomain.ErrNotFound) {
		t.Fatalf("UpgradeSlackCredential() error = %v, want ErrNotFound", err)
	}
	var stored string
	if err := db.QueryRow(ctx, `SELECT bot_access_token FROM slack_workspaces WHERE id = $1`, installationID).Scan(&stored); err != nil {
		t.Fatalf("read concurrent Slack credential: %v", err)
	}
	if stored != "concurrent-plaintext" {
		t.Fatal("CAS overwrote concurrent Slack credential")
	}
}

func seedSlackCredentialOwner(t testing.TB, db *pgxpool.Pool) (uuid.UUID, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	if _, err := db.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Slack credential migration')
	`, userID, "slack-vault-"+suffix, "slack-vault-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert Slack credential owner: %v", err)
	}
	if _, err := db.Exec(t.Context(), `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Slack credential migration', $2, $3)
	`, workspaceID, "slack-vault-"+suffix, userID); err != nil {
		t.Fatalf("insert Slack credential workspace: %v", err)
	}
	if _, err := db.Exec(t.Context(), `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert Slack credential workspace administrator: %v", err)
	}
	return userID, workspaceID
}
