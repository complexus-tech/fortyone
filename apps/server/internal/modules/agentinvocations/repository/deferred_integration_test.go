//go:build integration

package agentinvocationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestDeferredActionStageIsAtomicFencedAndCannotBeCancelled(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	run := acceptInvocation(t, repo, input)
	run, err := repo.Claim(ctx, run.ID)
	require.NoError(t, err)
	run, err = repo.SaveProposal(ctx, run, domain.Proposal{Action: json.RawMessage(`{"edit":true}`)})
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `CREATE TABLE stage_probe(run_id uuid PRIMARY KEY)`)
	require.NoError(t, err)
	begin := func(txCtx context.Context, current domain.Run) error {
		tx, ok := database.UnitTransaction(txCtx)
		require.True(t, ok)
		_, err := tx.Exec(txCtx, `INSERT INTO stage_probe(run_id) VALUES($1)`, current.ID)
		return err
	}
	rollback := errors.New("outer transaction rolled back")
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txCtx context.Context, _ pgx.Tx) error {
		_, err := repo.BeginAction(txCtx, run, begin)
		if err != nil {
			return err
		}
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	stored, err := repo.Get(ctx, input.WorkspaceID, run.ID)
	require.NoError(t, err)
	require.Nil(t, stored.ActionStartedAt)
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM stage_probe`).Scan(&count))
	require.Zero(t, count)
	run, err = repo.BeginAction(ctx, run, begin)
	require.NoError(t, err)
	require.NotNil(t, run.ActionStartedAt)
	_, err = repo.BeginAction(ctx, run, func(context.Context, domain.Run) error { t.Fatal("durable stage replayed"); return nil })
	require.NoError(t, err)
	_, err = repo.Control(ctx, domain.Control{WorkspaceID: input.WorkspaceID, RequesterID: input.RequesterID, RunID: run.ID, ExpectedVersion: run.Version}, "cancel")
	require.ErrorIs(t, err, domain.ErrConflict)
	require.NoError(t, repo.DeferAction(ctx, run))
	stored, err = repo.Get(ctx, input.WorkspaceID, run.ID)
	require.NoError(t, err)
	require.Equal(t, domain.StatusQueued, stored.Status)
	require.Zero(t, stored.RetryAttemptCount)
	_, err = repo.BeginAction(ctx, run, begin)
	require.ErrorIs(t, err, domain.ErrLeaseLost)
}

func TestRecoveryPrioritizesReadyWorkOverOlderDeferredRequests(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	repo := New(postgres.Pool)
	input := invocationFixture(t, postgres.Pool)
	older := acceptInvocation(t, repo, input)
	input.SourceID = uuid.New()
	input.IdempotencyKey = uuid.NewString()
	newer := acceptInvocation(t, repo, input)
	_, err := postgres.Pool.Exec(ctx, `UPDATE agent_invocations SET available_at=clock_timestamp()-interval '1 second' WHERE id=$1`, older.ID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE agent_invocations SET available_at=clock_timestamp()-interval '1 minute' WHERE id=$1`, newer.ID)
	require.NoError(t, err)
	ready, err := repo.Pending(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{newer.ID}, ready, "a repeatedly deferred older request must not starve ready work behind it")
}
