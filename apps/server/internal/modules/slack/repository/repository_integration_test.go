//go:build integration

package slackrepository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestSlackRepositoryPG18SecurityAtomicityAndStableOrdering(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	pool := postgres.Pool
	ctx := t.Context()
	assertSlackPostgres18(t, ctx, pool)
	repository := New(pool)

	primary := seedSlackTenant(t, ctx, pool, "primary", "admin")
	member := seedSlackActor(t, ctx, pool, primary.workspaceID, "member", "member", true)
	foreign := seedSlackTenant(t, ctx, pool, "foreign", "admin")
	teamID := seedSlackTeam(t, ctx, pool, primary.workspaceID, "PLAT", false)
	generation := uuid.New()
	installation := mustUpsertSlackInstallation(t, ctx, repository, primary, "T-PRIMARY", generation)

	t.Run("live actor and tenant checks are repeated at persistence", func(t *testing.T) {
		query := slackdomain.WorkspaceActorQuery{WorkspaceID: primary.workspaceID, ActorID: member}
		visible, err := repository.GetSlackWorkspaceForMember(ctx, query)
		require.NoError(t, err)
		require.Equal(t, installation.ID, visible.ID)

		_, err = repository.GetSlackWorkspaceForMember(ctx, slackdomain.WorkspaceActorQuery{
			WorkspaceID: primary.workspaceID,
			ActorID:     foreign.actorID,
		})
		require.ErrorIs(t, err, slackdomain.ErrForbidden)

		_, err = pool.Exec(ctx, `UPDATE users SET is_active = FALSE WHERE user_id = $1`, member)
		require.NoError(t, err)
		_, err = repository.GetSlackWorkspaceForMember(ctx, query)
		require.ErrorIs(t, err, slackdomain.ErrForbidden)
		_, err = pool.Exec(ctx, `UPDATE users SET is_active = TRUE WHERE user_id = $1`, member)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `UPDATE workspaces SET deleted_at = NOW() WHERE workspace_id = $1`, primary.workspaceID)
		require.NoError(t, err)
		_, err = repository.GetSlackWorkspaceForMember(ctx, query)
		require.ErrorIs(t, err, slackdomain.ErrForbidden)
		_, err = pool.Exec(ctx, `UPDATE workspaces SET deleted_at = NULL WHERE workspace_id = $1`, primary.workspaceID)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, primary.workspaceID, member)
		require.NoError(t, err)
		_, err = repository.GetSlackWorkspaceForMember(ctx, query)
		require.ErrorIs(t, err, slackdomain.ErrForbidden)
	})

	t.Run("channel and audience writes require admin and exact generation", func(t *testing.T) {
		memberID := seedSlackActor(t, ctx, pool, primary.workspaceID, "audience-member", "member", true)
		baseCommand := slackdomain.SyncChannelsCommand{
			WorkspaceID:            primary.workspaceID,
			ActorID:                memberID,
			InstallationID:         installation.ID,
			InstallationGeneration: generation,
			Channels: []slackdomain.ChannelUpsert{{
				SlackChannelID: "C-PLATFORM",
				Name:           "platform",
				IsMember:       true,
			}},
			Now: time.Now().UTC(),
		}
		require.ErrorIs(t, repository.UpsertChannels(ctx, baseCommand), slackdomain.ErrForbidden)

		baseCommand.ActorID = primary.actorID
		require.NoError(t, repository.UpsertChannels(ctx, baseCommand))
		audience := slackdomain.ReplaceChannelAudienceCommand{
			WorkspaceID:            primary.workspaceID,
			ActorID:                memberID,
			InstallationID:         installation.ID,
			InstallationGeneration: generation,
			SlackChannelID:         "C-PLATFORM",
			Configured:             true,
			TeamIDs:                []uuid.UUID{teamID},
			Now:                    time.Now().UTC(),
		}
		require.ErrorIs(t, repository.ReplaceAssistantChannelTeamAccess(ctx, audience), slackdomain.ErrForbidden)
		audience.ActorID = primary.actorID
		require.NoError(t, repository.ReplaceAssistantChannelTeamAccess(ctx, audience))

		currentGeneration := uuid.New()
		current := mustUpsertSlackInstallation(t, ctx, repository, primary, "T-PRIMARY", currentGeneration)
		require.Equal(t, installation.ID, current.ID)
		require.ErrorIs(t, repository.UpsertChannels(ctx, baseCommand), slackdomain.ErrConflict)
		require.ErrorIs(t, repository.ReplaceAssistantChannelTeamAccess(ctx, audience), slackdomain.ErrConflict)

		channels, err := repository.ListChannelsForMember(ctx, slackdomain.WorkspaceActorQuery{
			WorkspaceID: primary.workspaceID,
			ActorID:     primary.actorID,
		})
		require.NoError(t, err)
		require.Len(t, channels, 1)
		require.Equal(t, "C-PLATFORM", channels[0].SlackChannelID)

		installation = current
		generation = currentGeneration
	})

	t.Run("disconnect is actor-checked and commits durable outbox with deletion", func(t *testing.T) {
		memberID := seedSlackActor(t, ctx, pool, primary.workspaceID, "disconnect-member", "member", true)
		command := slackdomain.DisconnectInstallationCommand{
			WorkspaceID: primary.workspaceID,
			ActorID:     memberID,
			Now:         time.Now().UTC(),
		}
		_, err := repository.DisconnectSlackWorkspace(ctx, command)
		require.ErrorIs(t, err, slackdomain.ErrForbidden)
		stillPresent, err := repository.GetSlackWorkspace(ctx, primary.workspaceID)
		require.NoError(t, err)
		require.Equal(t, installation.ID, stillPresent.ID)

		command.ActorID = primary.actorID
		uninstall, err := repository.DisconnectSlackWorkspace(ctx, command)
		require.NoError(t, err)
		require.Equal(t, installation.ID, uninstall.SlackWorkspaceID)
		require.Equal(t, generation, uninstall.InstallGeneration)
		_, err = repository.GetSlackWorkspace(ctx, primary.workspaceID)
		require.ErrorIs(t, err, slackdomain.ErrNotFound)

		var outboxCount int
		err = pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM slack_uninstall_outbox
			WHERE id = $1
			  AND status = 'pending'
			  AND credential_payload IS NOT NULL
		`, uninstall.ID).Scan(&outboxCount)
		require.NoError(t, err)
		require.Equal(t, 1, outboxCount)
	})

	t.Run("request logs are tenant scoped and deterministically tie broken", func(t *testing.T) {
		fixed := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
		ids := []uuid.UUID{
			uuid.MustParse("10000000-0000-4000-8000-000000000001"),
			uuid.MustParse("10000000-0000-4000-8000-000000000002"),
			uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		}
		for _, id := range ids {
			_, err := pool.Exec(ctx, `
				INSERT INTO slack_request_logs (
					id, request_type, endpoint, workspace_id, response_code, outcome, created_at
				) VALUES ($1, 'events', '/slack/events', $2, 200, 'accepted', $3)
			`, id, primary.workspaceID, fixed)
			require.NoError(t, err)
		}
		_, err := pool.Exec(ctx, `
			INSERT INTO slack_request_logs (
				request_type, endpoint, workspace_id, response_code, outcome, created_at
			) VALUES ('events', '/slack/events', $1, 200, 'accepted', $2)
		`, foreign.workspaceID, fixed.Add(time.Hour))
		require.NoError(t, err)

		logs, err := repository.ListRequestLogsForAdmin(ctx, slackdomain.ListRequestLogsQuery{
			WorkspaceID: primary.workspaceID,
			ActorID:     primary.actorID,
			Limit:       10,
		})
		require.NoError(t, err)
		require.Len(t, logs, len(ids))
		require.Equal(t, ids[2], logs[0].ID)
		require.Equal(t, ids[1], logs[1].ID)
		require.Equal(t, ids[0], logs[2].ID)

		_, err = repository.ListRequestLogsForAdmin(ctx, slackdomain.ListRequestLogsQuery{
			WorkspaceID: primary.workspaceID,
			ActorID:     foreign.actorID,
			Limit:       10,
		})
		require.ErrorIs(t, err, slackdomain.ErrForbidden)
		assertSlackRequestLogIndexPlan(t, ctx, pool, primary.workspaceID, primary.actorID)
	})

	t.Run("legacy webhook reseal uses stable scan and exact compare and swap", func(t *testing.T) {
		webhookTenant := seedSlackTenant(t, ctx, pool, "webhook", "admin")
		webhookGeneration := uuid.New()
		webhookInstallation := mustUpsertSlackInstallation(t, ctx, repository, webhookTenant, "T-WEBHOOK", webhookGeneration)
		legacyPayload := "legacy-secretbox-payload"
		receiptID := uuid.MustParse("20000000-0000-4000-8000-000000000001")
		_, err := pool.Exec(ctx, `
			INSERT INTO messaging_inbound_events (
				id, provider, workspace_id, installation_id, installation_generation,
				external_workspace_id, external_event_id, event_type, payload_encrypted, status
			) VALUES ($1, 'slack', $2, $3, $4, 'T-WEBHOOK', 'Ev-webhook', 'event_callback', $5, 'pending')
		`, receiptID, webhookTenant.workspaceID, webhookInstallation.ID, webhookGeneration, legacyPayload)
		require.NoError(t, err)

		records, err := repository.ListLegacySlackWebhookPayloads(ctx, uuid.Nil, 10)
		require.NoError(t, err)
		require.Len(t, records, 1)
		require.Equal(t, receiptID, records[0].ID)

		_, err = pool.Exec(ctx, `UPDATE messaging_inbound_events SET payload_encrypted = 'concurrent' WHERE id = $1`, receiptID)
		require.NoError(t, err)
		err = repository.UpgradeLegacySlackWebhookPayload(ctx, records[0], legacyPayload, "slack-webhook.v2.replacement")
		require.ErrorIs(t, err, slackdomain.ErrNotFound)

		var persisted string
		require.NoError(t, pool.QueryRow(ctx, `SELECT payload_encrypted FROM messaging_inbound_events WHERE id = $1`, receiptID).Scan(&persisted))
		require.Equal(t, "concurrent", persisted)
	})
}

func TestSlackRepositoryConcurrentInstallationsHaveSingleWinner(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	pool := postgres.Pool
	ctx := t.Context()
	repository := New(pool)
	tenant := seedSlackTenant(t, ctx, pool, "race", "admin")
	generations := []uuid.UUID{uuid.New(), uuid.New()}
	credentials := []string{
		sealSlackRepositoryCredential(t, tenant.workspaceID, generations[0]),
		sealSlackRepositoryCredential(t, tenant.workspaceID, generations[1]),
	}

	start := make(chan struct{})
	errorsByAttempt := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		index := index
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := repository.UpsertSlackWorkspace(ctx, tenant.workspaceID, tenant.actorID, OAuthInstallPayload{
				SlackTeamID:       fmt.Sprintf("T-RACE-%d", index),
				SlackTeamName:     "Race",
				SlackTeamDomain:   fmt.Sprintf("race-%d", index),
				BotAccessToken:    credentials[index],
				CredentialVersion: credentialvault.CurrentVersion,
				InstallGeneration: generations[index],
			})
			errorsByAttempt <- err
		}()
	}
	close(start)
	wait.Wait()
	close(errorsByAttempt)

	successes := 0
	conflicts := 0
	for err := range errorsByAttempt {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, slackdomain.ErrConflict):
			conflicts++
		default:
			t.Fatalf("concurrent Slack installation returned unexpected error: %v", err)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
}

type slackIntegrationTenant struct {
	workspaceID uuid.UUID
	actorID     uuid.UUID
}

func seedSlackTenant(
	t testing.TB,
	ctx context.Context,
	pool *pgxpool.Pool,
	label, role string,
) slackIntegrationTenant {
	t.Helper()
	actorID := uuid.New()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, TRUE, FALSE)
	`, actorID, "slack-"+label+"-"+suffix, "slack-"+label+"-"+suffix+"@example.test", "Slack "+label)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, "Slack "+label, "slack-"+label+"-"+suffix, actorID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, actorID, role)
	require.NoError(t, err)
	return slackIntegrationTenant{workspaceID: workspaceID, actorID: actorID}
}

func seedSlackActor(
	t testing.TB,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	label, role string,
	active bool,
) uuid.UUID {
	t.Helper()
	actorID := uuid.New()
	suffix := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, FALSE)
	`, actorID, "slack-"+label+"-"+suffix, "slack-"+label+"-"+suffix+"@example.test", "Slack "+label, active)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, actorID, role)
	require.NoError(t, err)
	return actorID
}

func seedSlackTeam(t testing.TB, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, code string, private bool) uuid.UUID {
	t.Helper()
	teamID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, workspace_id, code, name, color, is_private)
		VALUES ($1, $2, $3, $4, '#123456', $5)
	`, teamID, workspaceID, code, "Slack "+code, private)
	require.NoError(t, err)
	return teamID
}

func mustUpsertSlackInstallation(
	t testing.TB,
	ctx context.Context,
	repository *Repo,
	tenant slackIntegrationTenant,
	slackTeamID string,
	generation uuid.UUID,
) SlackWorkspaceRecord {
	t.Helper()
	installation, err := repository.UpsertSlackWorkspace(ctx, tenant.workspaceID, tenant.actorID, OAuthInstallPayload{
		SlackTeamID:       slackTeamID,
		SlackTeamName:     "Slack test",
		SlackTeamDomain:   strings.ToLower(slackTeamID),
		BotAccessToken:    sealSlackRepositoryCredential(t, tenant.workspaceID, generation),
		CredentialVersion: credentialvault.CurrentVersion,
		InstallGeneration: generation,
	})
	require.NoError(t, err)
	return installation
}

func sealSlackRepositoryCredential(t testing.TB, workspaceID, generation uuid.UUID) string {
	t.Helper()
	ref := credentialvault.KeyRef{ID: "slack-integration", Version: 1}
	vault, err := credentialvault.New(credentialvault.Config{
		Active: ref,
		Keys: []credentialvault.Key{{
			Ref:      ref,
			Material: bytes.Repeat([]byte{0x53}, 32),
		}},
	})
	require.NoError(t, err)
	envelope, err := vault.Seal(credentialvault.Context{
		Provider:       "slack",
		TenantID:       workspaceID.String(),
		SubjectID:      "T-INTEGRATION",
		CredentialType: "installation",
		Generation:     generation.String(),
	}, []byte("xoxb-integration-secret"))
	require.NoError(t, err)
	return envelope
}

func assertSlackPostgres18(t testing.TB, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var version int
	require.NoError(t, pool.QueryRow(ctx, `SELECT CAST(current_setting('server_version_num') AS integer)`).Scan(&version))
	require.GreaterOrEqual(t, version, 180000)
}

func assertSlackRequestLogIndexPlan(
	t testing.TB,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, actorID uuid.UUID,
) {
	t.Helper()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `SET LOCAL enable_seqscan = off`)
	require.NoError(t, err)
	rows, err := tx.Query(ctx, `
		EXPLAIN (COSTS OFF)
		SELECT request.id
		FROM slack_request_logs AS request
		JOIN workspaces AS workspace
		  ON workspace.workspace_id = request.workspace_id
		 AND workspace.deleted_at IS NULL
		JOIN workspace_members AS membership
		  ON membership.workspace_id = request.workspace_id
		 AND membership.user_id = $2
		 AND membership.role = 'admin'
		JOIN users AS actor
		  ON actor.user_id = membership.user_id
		 AND actor.is_active = TRUE
		WHERE request.workspace_id = $1
		ORDER BY request.created_at DESC, request.id DESC
		LIMIT 50
	`, workspaceID, actorID)
	require.NoError(t, err)
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	require.NoError(t, rows.Err())
	require.Contains(t, plan.String(), "idx_slack_request_logs_workspace_id")
}
