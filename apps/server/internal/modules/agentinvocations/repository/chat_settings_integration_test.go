//go:build integration

package agentinvocationsrepository

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestDisabledChatInvocationsRemainPendingWithoutStarvingOtherSurfaces(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	queued := seedSurfaceInvocation(t, postgres.Pool, input, "chat_message")
	expired := seedSurfaceInvocation(t, postgres.Pool, input, "chat_message")
	expired, err := repo.Claim(ctx, expired.ID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE public.agent_invocations
		SET lease_expires_at=clock_timestamp()-INTERVAL '1 minute' WHERE id=$1`, expired.ID)
	require.NoError(t, err)

	// Both paused runs sort before the two unrelated surfaces. Filtering after
	// LIMIT would therefore starve comment/document recovery on every sweep.
	_, err = postgres.Pool.Exec(ctx, `UPDATE public.agent_invocations
		SET available_at=clock_timestamp()-INTERVAL '2 minutes' WHERE id=$1`, queued.ID)
	require.NoError(t, err)
	story := seedSurfaceInvocation(t, postgres.Pool, input, "story_comment")
	document := seedSurfaceInvocation(t, postgres.Pool, input, "document_comment")
	setInvocationChatEnabled(t, postgres.Pool, input.WorkspaceID, false)

	pending, err := repo.Pending(ctx, 2)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{story.ID, document.ID}, pending)
	for _, run := range []domain.Run{queued, expired} {
		before := invocationSnapshot(t, postgres.Pool, run.ID)
		_, err := repo.Claim(ctx, run.ID)
		require.ErrorIs(t, err, domain.ErrLeaseLost)
		require.Equal(t, before, invocationSnapshot(t, postgres.Pool, run.ID), "disabled claim must not mutate a lease or consume retries")
	}
	for _, run := range []domain.Run{story, document} {
		claimed, err := repo.Claim(ctx, run.ID)
		require.NoError(t, err)
		require.Equal(t, domain.StatusRunning, claimed.Status)
	}

	setInvocationChatEnabled(t, postgres.Pool, input.WorkspaceID, true)
	pending, err = repo.Pending(ctx, 2)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{queued.ID, expired.ID}, pending)
	resumed, err := repo.Claim(ctx, queued.ID)
	require.NoError(t, err)
	require.Equal(t, int32(1), resumed.AttemptCount)
	recovered, err := repo.Claim(ctx, expired.ID)
	require.NoError(t, err)
	require.Equal(t, expired.AttemptCount+1, recovered.AttemptCount)
	require.Greater(t, recovered.LeaseGeneration, expired.LeaseGeneration)
	require.NotEqual(t, expired.LeaseToken, recovered.LeaseToken)
}

func TestChatInvocationClaimRechecksSettingAfterRecoverySelection(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := seedSurfaceInvocation(t, postgres.Pool, input, "chat_message")
	pending, err := repo.Pending(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{run.ID}, pending)
	setInvocationChatEnabled(t, postgres.Pool, input.WorkspaceID, false)
	before := invocationSnapshot(t, postgres.Pool, run.ID)
	_, err = repo.Claim(ctx, pending[0])
	require.ErrorIs(t, err, domain.ErrLeaseLost)
	require.Equal(t, before, invocationSnapshot(t, postgres.Pool, run.ID))
}

func TestChatInvocationRecoveryDefaultsOnOnlyForExistingWorkspaces(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	_, err := postgres.Pool.Exec(ctx, `DELETE FROM public.workspace_settings WHERE workspace_id=$1`, input.WorkspaceID)
	require.NoError(t, err)
	run := seedSurfaceInvocation(t, postgres.Pool, input, "chat_message")
	pending, err := repo.Pending(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{run.ID}, pending)
	_, err = repo.Claim(ctx, run.ID)
	require.NoError(t, err, "legacy workspace without settings retains default-on behavior")

	run = seedSurfaceInvocation(t, postgres.Pool, input, "chat_message")
	_, err = postgres.Pool.Exec(ctx, `UPDATE public.workspaces SET deleted_at=clock_timestamp() WHERE workspace_id=$1`, input.WorkspaceID)
	require.NoError(t, err)
	pending, err = repo.Pending(ctx, 1)
	require.NoError(t, err)
	require.Empty(t, pending)
	_, err = repo.Claim(ctx, run.ID)
	require.ErrorIs(t, err, domain.ErrLeaseLost)
}

// Seed persisted jobs directly: surface-specific source acceptance is exercised
// by the adapter suites; these tests isolate queue selection and lease behavior.
func seedSurfaceInvocation(t *testing.T, pool *pgxpool.Pool, input domain.Input, surface string) domain.Run {
	t.Helper()
	var teamID *uuid.UUID
	if surface != "document_comment" {
		teamID = &input.TeamID
	}
	var id uuid.UUID
	err := pool.QueryRow(t.Context(), `INSERT INTO public.agent_invocations
		(workspace_id,requester_id,team_id,surface,source_id,conversation_id,entity_id,
		 source_revision,source_hash,idempotency_key,request_hash,payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'1','test-source',$8,$9,'{}') RETURNING id`,
		input.WorkspaceID, input.RequesterID, teamID, surface, uuid.New(), uuid.New(), uuid.New(), uuid.NewString(), make([]byte, 32)).Scan(&id)
	require.NoError(t, err)
	run, err := New(pool).Get(t.Context(), input.WorkspaceID, id)
	require.NoError(t, err)
	return run
}

func setInvocationChatEnabled(t *testing.T, pool *pgxpool.Pool, workspaceID uuid.UUID, enabled bool) {
	t.Helper()
	_, err := pool.Exec(t.Context(), `INSERT INTO public.workspace_settings (workspace_id,chat_enabled)
		VALUES ($1,$2) ON CONFLICT (workspace_id) DO UPDATE SET chat_enabled=EXCLUDED.chat_enabled`, workspaceID, enabled)
	require.NoError(t, err)
}

func invocationSnapshot(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) []byte {
	t.Helper()
	var snapshot []byte
	err := pool.QueryRow(t.Context(), `SELECT to_jsonb(invocation) FROM public.agent_invocations invocation WHERE id=$1`, id).Scan(&snapshot)
	require.NoError(t, err)
	return snapshot
}
