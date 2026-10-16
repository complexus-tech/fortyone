//go:build integration

package messagingrepository

import (
	"context"
	"errors"
	"sync"
	"testing"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestNativeUsagePersistsWithoutExternalInboundEventAndDeduplicatesAttempts(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, userID, teamID, runID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, seed := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users (user_id,username,email) VALUES ($1,$2,$3)`, []any{userID, "native-" + userID.String(), userID.String() + "@example.test"}},
		{`INSERT INTO workspaces (workspace_id,name,slug,created_by) VALUES ($1,'Native usage',$2,$3)`, []any{workspaceID, "usage-" + workspaceID.String(), userID}},
		{`INSERT INTO teams (team_id,workspace_id,name,code,color) VALUES ($1,$2,'Web','WEB','#111111')`, []any{teamID, workspaceID}},
		{`INSERT INTO agent_invocations (id,workspace_id,requester_id,team_id,surface,source_id,conversation_id,entity_id,source_revision,source_hash,idempotency_key,request_hash) VALUES ($1,$2,$3,$4,'story_comment',$5,$5,$6,'version','hash',$7,$8)`, []any{runID, workspaceID, userID, teamID, uuid.New(), uuid.New(), uuid.NewString(), make([]byte, 32)}},
	} {
		_, err := postgres.Pool.Exec(ctx, seed.query, seed.args...)
		require.NoError(t, err)
	}
	repository := NewDailyUsageRepository(postgres.Pool)
	input := DailyUsageRecordInput{AgentInvocationID: runID, WorkspaceID: workspaceID, Provider: "native", AttemptCount: 1, Usage: messaging.Usage{InputTokens: 100, OutputTokens: 25, TotalTokens: 125}}
	first, err := repository.Record(ctx, input, 0)
	require.NoError(t, err)
	require.EqualValues(t, 125, first.TotalTokens)
	require.EqualValues(t, 1, first.RequestCount)
	duplicate, err := repository.Record(ctx, input, 0)
	require.NoError(t, err)
	require.Equal(t, first, duplicate)
	input.AttemptCount = 2
	second, err := repository.Record(ctx, input, 0)
	require.NoError(t, err)
	require.EqualValues(t, 250, second.TotalTokens)
	require.EqualValues(t, 2, second.RequestCount)
	input.AttemptCount = 3
	var group sync.WaitGroup
	failures := make(chan error, 8)
	for range 8 {
		group.Go(func() { _, err := repository.Record(ctx, input, 0); failures <- err })
	}
	group.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	snapshot, err := repository.Check(ctx, workspaceID, 0)
	require.NoError(t, err)
	require.EqualValues(t, 375, snapshot.TotalTokens)
	require.EqualValues(t, 3, snapshot.RequestCount)
	wrongWorkspace := input
	wrongWorkspace.WorkspaceID = uuid.New()
	_, err = repository.Record(ctx, wrongWorkspace, 0)
	require.ErrorContains(t, err, "missing or outside the workspace")
	missing := input
	missing.AgentInvocationID = uuid.New()
	_, err = repository.Record(ctx, missing, 0)
	require.ErrorContains(t, err, "missing or outside the workspace")
	input.AttemptCount = 4
	rollback := errors.New("rollback usage and aggregate")
	err = database.WithinUnitOfWork(ctx, postgres.Pool, func(txctx context.Context, _ pgx.Tx) error {
		_, err := repository.Record(txctx, input, 0)
		if err != nil {
			return err
		}
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	afterRollback, err := repository.Check(ctx, workspaceID, 0)
	require.NoError(t, err)
	require.Equal(t, snapshot, afterRollback)
	var nativeCount, externalCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM agent_invocation_assistant_usage_events WHERE invocation_id=$1`, runID).Scan(&nativeCount))
	require.Equal(t, 3, nativeCount)
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM messaging_assistant_usage_events WHERE workspace_id=$1`, workspaceID).Scan(&externalCount))
	require.Zero(t, externalCount)
}
