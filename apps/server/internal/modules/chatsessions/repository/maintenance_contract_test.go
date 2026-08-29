package chatsessionsrepository

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/stretchr/testify/require"
)

func TestChatSessionMaintenanceQueryIsBoundedAndPreservesUnresolvedApprovals(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/maintenance.sql")
	if err != nil {
		t.Fatalf("read chat-session maintenance queries: %v", err)
	}
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: purgedeletedchatsessions :execrows",
		"with candidates as materialized",
		"session.deleted_at < sqlc.arg(deleted_before)",
		"not exists",
		"execution.session_id = session.id",
		"execution.status in ('ready', 'retry_ready', 'executing', 'failed_uncertain')",
		"order by session.deleted_at, session.id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"for update of session skip locked",
		"delete from public.chat_sessions as session using candidates",
		"session.id = candidates.id",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("chat-session maintenance query is missing contract %q", contract)
		}
	}

	if strings.Contains(query, "current_timestamp") || strings.Contains(query, "now()") {
		t.Fatal("chat-session retention cutoff must be supplied by the application clock")
	}
}

func TestChatSessionMaintenanceRepositoryMapsUTCAndBatchToSQLC(t *testing.T) {
	t.Parallel()

	queries := &chatSessionMaintenanceQueries{rows: 3}
	repository := &repo{queries: queries}
	deletedBefore := time.Date(2026, time.July, 29, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))

	deleted, err := repository.PurgeDeletedChatSessions(context.Background(), deletedBefore, 500)

	require.NoError(t, err)
	require.Equal(t, int64(3), deleted)
	require.Equal(t, 1, queries.calls)
	require.NotNil(t, queries.params.DeletedBefore)
	require.Equal(t, deletedBefore.UTC(), *queries.params.DeletedBefore)
	require.Equal(t, int32(500), queries.params.BatchSize)
}

func TestChatSessionMaintenanceRepositoryRejectsBatchOutsideSQLCRange(t *testing.T) {
	t.Parallel()

	queries := &chatSessionMaintenanceQueries{}
	repository := &repo{queries: queries}

	_, err := repository.PurgeDeletedChatSessions(
		context.Background(),
		time.Date(2026, time.July, 29, 8, 15, 0, 0, time.UTC),
		int(math.MaxInt32)+1,
	)

	require.ErrorIs(t, err, safecast.ErrOutOfRange)
	require.Zero(t, queries.calls)
}

type chatSessionMaintenanceQueries struct {
	chatsessionssql.Querier
	params chatsessionssql.PurgeDeletedChatSessionsParams
	rows   int64
	err    error
	calls  int
}

func (queries *chatSessionMaintenanceQueries) PurgeDeletedChatSessions(
	_ context.Context,
	params chatsessionssql.PurgeDeletedChatSessionsParams,
) (int64, error) {
	queries.calls++
	queries.params = params
	return queries.rows, queries.err
}
