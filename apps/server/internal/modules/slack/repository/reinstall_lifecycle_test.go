package slackrepository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type requestThreadRebindExecerStub struct {
	query string
	args  []any
	err   error
}

func (s *requestThreadRebindExecerStub) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	s.query = query
	s.args = append([]any(nil), args...)
	return requestThreadRebindResult(1), s.err
}

type requestThreadRebindResult int64

func (r requestThreadRebindResult) LastInsertId() (int64, error) { return 0, nil }
func (r requestThreadRebindResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestRebindSlackRequestThreadsScopesTheGenerationRotation(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	previousGeneration := uuid.New()
	currentGeneration := uuid.New()
	execer := &requestThreadRebindExecerStub{}

	err := rebindSlackRequestThreadsTx(
		context.Background(),
		execer,
		workspaceID,
		" T123 ",
		previousGeneration,
		currentGeneration,
	)

	require.NoError(t, err)
	require.Equal(t, []any{workspaceID, "T123", previousGeneration, currentGeneration}, execer.args)
	query := strings.Join(strings.Fields(execer.query), " ")
	for _, clause := range []string{
		"SET installation_generation = $4",
		"WHERE workspace_id = $1",
		"AND provider = 'slack'",
		"AND external_workspace_id = $2",
		"AND installation_generation = $3",
	} {
		require.Contains(t, query, clause)
	}
}

func TestRebindSlackRequestThreadsRejectsUnsafeLifecycleInputs(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	previousGeneration := uuid.New()
	currentGeneration := uuid.New()
	tests := []struct {
		name               string
		workspaceID        uuid.UUID
		slackTeamID        string
		previousGeneration uuid.UUID
		currentGeneration  uuid.UUID
	}{
		{name: "missing workspace", slackTeamID: "T123", previousGeneration: previousGeneration, currentGeneration: currentGeneration},
		{name: "missing Slack team", workspaceID: workspaceID, previousGeneration: previousGeneration, currentGeneration: currentGeneration},
		{name: "missing previous generation", workspaceID: workspaceID, slackTeamID: "T123", currentGeneration: currentGeneration},
		{name: "missing current generation", workspaceID: workspaceID, slackTeamID: "T123", previousGeneration: previousGeneration},
		{name: "generation did not rotate", workspaceID: workspaceID, slackTeamID: "T123", previousGeneration: previousGeneration, currentGeneration: previousGeneration},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			execer := &requestThreadRebindExecerStub{}

			err := rebindSlackRequestThreadsTx(
				context.Background(),
				execer,
				test.workspaceID,
				test.slackTeamID,
				test.previousGeneration,
				test.currentGeneration,
			)

			require.Error(t, err)
			require.Empty(t, execer.query)
		})
	}
}

func TestRebindSlackRequestThreadsWrapsPersistenceFailure(t *testing.T) {
	t.Parallel()

	execer := &requestThreadRebindExecerStub{err: errors.New("database unavailable")}
	err := rebindSlackRequestThreadsTx(
		context.Background(),
		execer,
		uuid.New(),
		"T123",
		uuid.New(),
		uuid.New(),
	)

	require.ErrorContains(t, err, "rebind Slack request threads to refreshed installation")
	require.ErrorContains(t, err, "database unavailable")
}

// TestSlackReinstallRequestThreadPostgresContract exercises the complete
// installation refresh transaction against a migrated PostgreSQL database.
// It is skipped in ordinary unit-test runs.
func TestSlackReinstallRequestThreadPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("SLACK_REINSTALL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SLACK_REINSTALL_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	repo := New(nil, db)
	userID, workspaceID, teamID := seedSlackReinstallWorkspace(t, db, "primary")
	otherUserID, otherWorkspaceID, otherTeamID := seedSlackReinstallWorkspace(t, db, "other")
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM workspaces WHERE workspace_id IN ($1, $2)", workspaceID, otherWorkspaceID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id IN ($1, $2)", userID, otherUserID)
	})

	first, err := repo.UpsertSlackWorkspace(ctx, workspaceID, userID, OAuthInstallPayload{
		SlackTeamID:       "T-PRIMARY",
		SlackTeamName:     "Primary",
		SlackTeamDomain:   "primary",
		BotAccessToken:    "encrypted-token-v1",
		LegacyAccessToken: "xoxb-v1",
		CredentialVersion: 1,
	})
	require.NoError(t, err)
	seedSlackChannelAudience(t, db, workspaceID, first.ID, teamID, "C-PRIMARY")

	matchingThreadID := seedSlackRequestThread(t, db, workspaceID, teamID, "T-PRIMARY", first.InstallGeneration, "matching")
	staleGenerationThreadID := seedSlackRequestThread(t, db, workspaceID, teamID, "T-PRIMARY", uuid.New(), "stale-generation")
	otherTeamThreadID := seedSlackRequestThread(t, db, workspaceID, teamID, "T-OTHER", first.InstallGeneration, "other-team")
	otherWorkspaceThreadID := seedSlackRequestThread(t, db, otherWorkspaceID, otherTeamID, "T-PRIMARY", first.InstallGeneration, "other-workspace")

	refreshed, err := repo.UpsertSlackWorkspace(ctx, workspaceID, userID, OAuthInstallPayload{
		SlackTeamID:       "T-PRIMARY",
		SlackTeamName:     "Primary",
		SlackTeamDomain:   "primary",
		BotAccessToken:    "encrypted-token-v2",
		LegacyAccessToken: "xoxb-v2",
		CredentialVersion: 1,
	})
	require.NoError(t, err)
	require.NotEqual(t, first.InstallGeneration, refreshed.InstallGeneration)

	require.Equal(t, refreshed.InstallGeneration, requestThreadGeneration(t, db, matchingThreadID))
	require.NotEqual(t, refreshed.InstallGeneration, requestThreadGeneration(t, db, staleGenerationThreadID))
	require.Equal(t, first.InstallGeneration, requestThreadGeneration(t, db, otherTeamThreadID))
	require.Equal(t, first.InstallGeneration, requestThreadGeneration(t, db, otherWorkspaceThreadID))

	_, err = repo.DisconnectSlackWorkspace(ctx, workspaceID)
	require.NoError(t, err)
	require.Zero(t, slackChannelAudienceCount(t, db, workspaceID), "disconnect must rely on the installation foreign-key cascade")
	reconnected, err := repo.UpsertSlackWorkspace(ctx, workspaceID, userID, OAuthInstallPayload{
		SlackTeamID:       "T-PRIMARY",
		SlackTeamName:     "Primary",
		SlackTeamDomain:   "primary",
		BotAccessToken:    "encrypted-token-v3",
		LegacyAccessToken: "xoxb-v3",
		CredentialVersion: 1,
	})
	require.NoError(t, err)
	require.NotEqual(t, refreshed.InstallGeneration, reconnected.InstallGeneration)
	require.Equal(t, refreshed.InstallGeneration, requestThreadGeneration(t, db, matchingThreadID), "an explicit disconnect must remain a hard lifecycle fence")

	otherInstallation, err := repo.UpsertSlackWorkspace(ctx, otherWorkspaceID, otherUserID, OAuthInstallPayload{
		SlackTeamID:       "T-OTHER-WORKSPACE",
		SlackTeamName:     "Other",
		SlackTeamDomain:   "other",
		BotAccessToken:    "encrypted-other-token",
		LegacyAccessToken: "xoxb-other",
		CredentialVersion: 1,
	})
	require.NoError(t, err)
	seedSlackChannelAudience(t, db, otherWorkspaceID, otherInstallation.ID, otherTeamID, "C-OTHER")
	require.NoError(t, repo.DeactivateSlackWorkspaceByTeamID(ctx, otherInstallation.SlackTeamID, otherInstallation.InstallGeneration))
	require.Zero(t, slackChannelAudienceCount(t, db, otherWorkspaceID), "provider revocation must rely on the installation foreign-key cascade")
}

func seedSlackReinstallWorkspace(t *testing.T, db *sqlx.DB, prefix string) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	suffix := uuid.NewString()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, prefix+"-"+suffix, prefix+"-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, prefix+" Slack Reinstall", prefix+"-"+suffix, userID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, userID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, $3, $4, '#111827')
	`, teamID, workspaceID, prefix+" Team", "R"+strings.ToUpper(suffix[:6]))
	require.NoError(t, err)
	return userID, workspaceID, teamID
}

func seedSlackRequestThread(
	t *testing.T,
	db *sqlx.DB,
	workspaceID uuid.UUID,
	teamID uuid.UUID,
	externalWorkspaceID string,
	installationGeneration uuid.UUID,
	suffix string,
) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	requestID := uuid.New()
	threadID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO integration_requests (
			id, workspace_id, team_id, provider, source_type,
			source_external_id, title, created_by_user_id
		) VALUES ($1, $2, $3, 'slack', 'message', $4, $5, NULL)
	`, requestID, workspaceID, teamID, suffix+"-"+uuid.NewString(), suffix)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO integration_request_threads (
			id, workspace_id, integration_request_id, provider,
			external_workspace_id, installation_generation,
			external_channel_id, external_thread_id
		) VALUES ($1, $2, $3, 'slack', $4, $5, $6, $7)
	`, threadID, workspaceID, requestID, externalWorkspaceID, installationGeneration, "C-"+suffix, "TS-"+uuid.NewString())
	require.NoError(t, err)
	return threadID
}

func requestThreadGeneration(t *testing.T, db *sqlx.DB, threadID uuid.UUID) uuid.UUID {
	t.Helper()
	var generation uuid.UUID
	require.NoError(t, db.GetContext(context.Background(), &generation, `
		SELECT installation_generation
		FROM integration_request_threads
		WHERE id = $1
	`, threadID))
	return generation
}

func seedSlackChannelAudience(
	t *testing.T,
	db *sqlx.DB,
	workspaceID uuid.UUID,
	slackWorkspaceID uuid.UUID,
	teamID uuid.UUID,
	channelID string,
) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO slack_channel_team_access (
			workspace_id, slack_workspace_id, slack_channel_id, team_id
		) VALUES ($1, $2, $3, $4)
	`, workspaceID, slackWorkspaceID, channelID, teamID)
	require.NoError(t, err)
}

func slackChannelAudienceCount(t *testing.T, db *sqlx.DB, workspaceID uuid.UUID) int {
	t.Helper()
	var count int
	require.NoError(t, db.GetContext(context.Background(), &count, `
		SELECT COUNT(*)
		FROM slack_channel_team_access
		WHERE workspace_id = $1
	`, workspaceID))
	return count
}
