//go:build integration

package workspacesrepository

import (
	"bytes"
	"testing"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspacePurgeDefersUnsafeSlackCredentialsAndSnapshotsCurrentVaultEnvelope(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	repository := New(postgres.Pool)
	ctx := t.Context()

	userID := uuid.New()
	workspaceID := uuid.New()
	installationID := uuid.New()
	generation := uuid.New()
	suffix := uuid.NewString()
	processedAt := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	deletedAt := processedAt.Add(-72 * time.Hour)
	const slackTeamID = "T-VAULT-PURGE"

	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "vault-purge-"+suffix, "vault-purge-"+suffix+"@example.com")
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by, deleted_at)
		VALUES ($1, 'Vault purge contract', $2, $3, $4)
	`, workspaceID, "vault-purge-"+suffix, userID, deletedAt)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO slack_workspaces (
			id, workspace_id, slack_team_id, slack_team_name, slack_team_domain,
			bot_access_token, credential_payload, credential_key_version,
			installation_generation, installed_by_user_id
		) VALUES ($1, $2, $3, 'Vault purge', 'vault-purge', '', 'plaintext-with-vault-metadata', $4, $5, $6)
	`, installationID, workspaceID, slackTeamID, credentialvault.CurrentVersion, generation, userID)
	require.NoError(t, err)

	batch := workspacedomain.DeletedWorkspacePurgeBatch{
		DeletedBefore:               processedAt.Add(-48 * time.Hour),
		ProcessedAt:                 processedAt,
		BatchSize:                   500,
		IntegrationLifecycleLockKey: 42,
	}
	unsafeResult, err := repository.PurgeSoftDeletedWorkspacesBatch(ctx, batch)
	require.NoError(t, err)
	require.Equal(t, 1, unsafeResult.CandidateCount)
	require.Equal(t, int64(0), unsafeResult.Deleted)
	require.Equal(t, int64(1), unsafeResult.Blocked)
	require.True(t, unsafeResult.Cursor.Valid)
	require.WithinDuration(t, deletedAt, unsafeResult.Cursor.DeletedAt, 0)
	require.Equal(t, workspaceID, unsafeResult.Cursor.WorkspaceID)

	var workspaceCount int
	require.NoError(t, postgres.Pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM workspaces WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&workspaceCount))
	require.Equal(t, 1, workspaceCount)
	var outboxCount int
	require.NoError(t, postgres.Pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM slack_uninstall_outbox WHERE slack_workspace_id = $1`,
		installationID,
	).Scan(&outboxCount))
	require.Zero(t, outboxCount)

	ref := credentialvault.KeyRef{ID: "purge-test", Version: 1}
	vault, err := credentialvault.New(credentialvault.Config{
		Active: ref,
		Keys:   []credentialvault.Key{{Ref: ref, Material: bytes.Repeat([]byte{0x6d}, 32)}},
	})
	require.NoError(t, err)
	envelope, err := vault.Seal(credentialvault.Context{
		Provider:       "slack",
		TenantID:       workspaceID.String(),
		SubjectID:      slackTeamID,
		CredentialType: "bot-oauth",
		Generation:     generation.String(),
	}, []byte(`{"accessToken":"xoxb-purge"}`))
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		UPDATE slack_workspaces
		SET credential_payload = $2,
			credential_key_version = $3,
			bot_access_token = ''
		WHERE id = $1
	`, installationID, envelope, credentialvault.CurrentVersion)
	require.NoError(t, err)

	vaultResult, err := repository.PurgeSoftDeletedWorkspacesBatch(ctx, batch)
	require.NoError(t, err)
	require.Equal(t, 1, vaultResult.CandidateCount)
	require.Equal(t, int64(1), vaultResult.Deleted)
	require.Zero(t, vaultResult.Blocked)
	require.True(t, vaultResult.Cursor.Valid)
	require.WithinDuration(t, deletedAt, vaultResult.Cursor.DeletedAt, 0)
	require.Equal(t, workspaceID, vaultResult.Cursor.WorkspaceID)

	var outboxPayload string
	var outboxVersion int
	var nextAttemptAt, createdAt, updatedAt time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT credential_payload, credential_key_version, next_attempt_at, created_at, updated_at
		FROM slack_uninstall_outbox
		WHERE slack_workspace_id = $1
	`, installationID).Scan(
		&outboxPayload,
		&outboxVersion,
		&nextAttemptAt,
		&createdAt,
		&updatedAt,
	))
	require.Equal(t, envelope, outboxPayload)
	require.Equal(t, credentialvault.CurrentVersion, outboxVersion)
	require.WithinDuration(t, processedAt, nextAttemptAt, 0)
	require.WithinDuration(t, processedAt, createdAt, 0)
	require.WithinDuration(t, processedAt, updatedAt, 0)
}
