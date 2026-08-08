package messagingrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type dailyUsageDBStub struct {
	rows       []dailyUsageRow
	err        error
	beginErr   error
	duplicate  bool
	committed  bool
	rolledBack bool
	queries    []string
	args       [][]any
}

func (s *dailyUsageDBStub) GetContext(_ context.Context, destination any, query string, args ...any) error {
	return s.getContext(destination, query, args...)
}

func (s *dailyUsageDBStub) Begin(context.Context) (dailyUsageTransaction, error) {
	if s.beginErr != nil {
		return nil, s.beginErr
	}
	return &dailyUsageTxStub{db: s}, nil
}

func (s *dailyUsageDBStub) getContext(destination any, query string, args ...any) error {
	s.queries = append(s.queries, query)
	s.args = append(s.args, append([]any(nil), args...))
	if s.err != nil {
		return s.err
	}
	if strings.Contains(query, "INSERT INTO messaging_assistant_usage_events") {
		if s.duplicate {
			return sql.ErrNoRows
		}
		inserted, ok := destination.(*int)
		if !ok {
			return errors.New("unexpected usage event destination")
		}
		*inserted = 1
		return nil
	}
	row, ok := destination.(*dailyUsageRow)
	if !ok {
		return errors.New("unexpected daily usage destination")
	}
	if len(s.rows) > 0 {
		*row = s.rows[0]
		s.rows = s.rows[1:]
	}
	return nil
}

type dailyUsageTxStub struct {
	db *dailyUsageDBStub
}

func (s *dailyUsageTxStub) GetContext(_ context.Context, destination any, query string, args ...any) error {
	return s.db.getContext(destination, query, args...)
}

func (s *dailyUsageTxStub) Commit() error {
	s.db.committed = true
	return nil
}

func (s *dailyUsageTxStub) Rollback() error {
	s.db.rolledBack = true
	return nil
}

func TestDailyUsageRepositoryCheckUsesDefaultCeiling(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{rows: []dailyUsageRow{{InputTokens: 800, OutputTokens: 200, TotalTokens: 1_000, RequestCount: 2}}}
	repository := &DailyUsageRepository{db: db}

	snapshot, err := repository.Check(context.Background(), dailyUsageWorkspaceID(), 0)

	require.NoError(t, err)
	require.True(t, snapshot.Allowed)
	require.Equal(t, DefaultDailyWorkspaceTokenLimit, snapshot.Limit)
	require.Equal(t, DefaultDailyWorkspaceTokenLimit-1_000, snapshot.Remaining)
	require.Contains(t, db.queries[0], "CAST(NOW() AT TIME ZONE 'UTC' AS date)")
}

func TestDailyUsageRepositoryCheckRejectsExhaustedWorkspace(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{rows: []dailyUsageRow{{InputTokens: 750, OutputTokens: 250, TotalTokens: 1_000, RequestCount: 3}}}
	repository := &DailyUsageRepository{db: db}

	snapshot, err := repository.Check(context.Background(), dailyUsageWorkspaceID(), 1_000)

	require.ErrorIs(t, err, ErrDailyWorkspaceTokenLimit)
	require.False(t, snapshot.Allowed)
	require.Zero(t, snapshot.Remaining)
	var limitError *DailyTokenLimitError
	require.ErrorAs(t, err, &limitError)
	require.Equal(t, int64(1_000), limitError.Used)
}

func TestDailyUsageRepositoryRecordPersistsResponsesUsage(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{rows: []dailyUsageRow{{InputTokens: 1_100, OutputTokens: 400, TotalTokens: 1_500, RequestCount: 4}}}
	repository := &DailyUsageRepository{db: db}
	usage := messaging.Usage{InputTokens: 300, OutputTokens: 100, TotalTokens: 400}

	snapshot, err := repository.Record(context.Background(), dailyUsageRecordInput(4, usage), 2_000)

	require.NoError(t, err)
	require.Equal(t, int64(1_500), snapshot.TotalTokens)
	require.Equal(t, int64(4), snapshot.RequestCount)
	require.True(t, snapshot.Allowed)
	require.True(t, db.committed)
	require.True(t, db.rolledBack, "deferred rollback should safely run after commit")
	require.Len(t, db.queries, 2)
	require.Contains(t, db.queries[0], "INSERT INTO messaging_assistant_usage_events")
	require.Contains(t, db.queries[0], "inbound_event_id,\n\t\t\tattempt_count\n\t\t) DO NOTHING")
	require.Contains(t, db.queries[0], "RETURNING 1")
	require.Contains(t, db.queries[1], "INSERT INTO messaging_assistant_daily_usage")
	require.Contains(t, db.queries[1], "ON CONFLICT (workspace_id, usage_date) DO UPDATE")
	require.Equal(t, []any{
		testUsageInboundEventID(),
		dailyUsageWorkspaceID(),
		"slack",
		"T1",
		"Ev1",
		4,
		int64(300),
		int64(100),
		int64(400),
	}, db.args[0])
	require.Equal(t, []any{
		dailyUsageWorkspaceID(),
		int64(300),
		int64(100),
		int64(400),
	}, db.args[1])
}

func TestDailyUsageRepositoryRecordReturnsExhaustedSnapshotAfterIncurredUsage(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{rows: []dailyUsageRow{{InputTokens: 900, OutputTokens: 200, TotalTokens: 1_100, RequestCount: 2}}}
	repository := &DailyUsageRepository{db: db}

	snapshot, err := repository.Record(context.Background(), dailyUsageRecordInput(2, messaging.Usage{
		InputTokens: 100, OutputTokens: 100, TotalTokens: 200,
	}), 1_000)

	require.NoError(t, err)
	require.False(t, snapshot.Allowed)
	require.Zero(t, snapshot.Remaining)
}

func TestDailyUsageRepositoryRejectsInvalidUsageBeforeDatabaseWrite(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{}
	repository := &DailyUsageRepository{db: db}

	_, err := repository.Record(context.Background(), dailyUsageRecordInput(1, messaging.Usage{
		InputTokens: 10, OutputTokens: 5, TotalTokens: 14,
	}), 1_000)

	require.ErrorContains(t, err, "total usage")
	require.Empty(t, db.queries)
}

func TestDailyUsageRepositoryUsesCastSyntaxInApplicationSQL(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{}
	repository := &DailyUsageRepository{db: db}

	_, err := repository.Check(context.Background(), dailyUsageWorkspaceID(), 1_000)

	require.NoError(t, err)
	require.NotContains(t, strings.Join(db.queries, "\n"), "::")
}

func TestDailyUsageRepositoryRecordRequiresDurableExecutionIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*DailyUsageRecordInput)
	}{
		{name: "provider", mutate: func(input *DailyUsageRecordInput) { input.Provider = "" }},
		{name: "external workspace", mutate: func(input *DailyUsageRecordInput) { input.ExternalWorkspaceID = "" }},
		{name: "external event", mutate: func(input *DailyUsageRecordInput) { input.ExternalEventID = "" }},
		{name: "inbound event", mutate: func(input *DailyUsageRecordInput) { input.InboundEventID = uuid.Nil }},
		{name: "attempt", mutate: func(input *DailyUsageRecordInput) { input.AttemptCount = 0 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			db := &dailyUsageDBStub{}
			repository := &DailyUsageRepository{db: db}
			input := dailyUsageRecordInput(1, messaging.Usage{InputTokens: 1, OutputTokens: 1, TotalTokens: 2})
			test.mutate(&input)

			_, err := repository.Record(context.Background(), input, 1_000)

			require.Error(t, err)
			require.Empty(t, db.queries)
		})
	}
}

func TestDailyUsageRepositoryDuplicateExecutionReturnsCurrentSnapshot(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{duplicate: true, rows: []dailyUsageRow{{
		InputTokens: 900, OutputTokens: 100, TotalTokens: 1_000, RequestCount: 3,
	}}}
	repository := &DailyUsageRepository{db: db}

	snapshot, err := repository.Record(context.Background(), dailyUsageRecordInput(3, messaging.Usage{
		InputTokens: 300, OutputTokens: 100, TotalTokens: 400,
	}), 2_000)

	require.NoError(t, err)
	require.Equal(t, int64(1_000), snapshot.TotalTokens)
	require.Equal(t, int64(3), snapshot.RequestCount)
	require.True(t, db.committed)
	require.Len(t, db.queries, 2)
	require.Contains(t, db.queries[0], "INSERT INTO messaging_assistant_usage_events")
	require.Contains(t, db.queries[1], "FROM messaging_assistant_daily_usage")
	require.NotContains(t, db.queries[1], "INSERT INTO messaging_assistant_daily_usage")
}

func TestDailyUsageRepositoryLaterAttemptUsesDistinctLedgerKey(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{rows: []dailyUsageRow{{
		InputTokens: 1_200, OutputTokens: 300, TotalTokens: 1_500, RequestCount: 2,
	}}}
	repository := &DailyUsageRepository{db: db}

	_, err := repository.Record(context.Background(), dailyUsageRecordInput(2, messaging.Usage{
		InputTokens: 200, OutputTokens: 50, TotalTokens: 250,
	}), 2_000)

	require.NoError(t, err)
	require.Equal(t, 2, db.args[0][5])
	require.Contains(t, db.queries[0], "attempt_count")
}

func TestDailyUsageRepositoryRollsBackWhenAggregateWriteFails(t *testing.T) {
	t.Parallel()

	db := &dailyUsageDBStub{err: errors.New("database unavailable")}
	repository := &DailyUsageRepository{db: db}

	_, err := repository.Record(context.Background(), dailyUsageRecordInput(1, messaging.Usage{
		InputTokens: 1, OutputTokens: 1, TotalTokens: 2,
	}), 2_000)

	require.ErrorContains(t, err, "claim messaging assistant usage event")
	require.False(t, db.committed)
	require.True(t, db.rolledBack)
}

func dailyUsageWorkspaceID() uuid.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func testUsageInboundEventID() uuid.UUID {
	return uuid.MustParse("22222222-2222-2222-2222-222222222222")
}

func dailyUsageRecordInput(attempt int, usage messaging.Usage) DailyUsageRecordInput {
	return DailyUsageRecordInput{
		InboundEventID:      testUsageInboundEventID(),
		WorkspaceID:         dailyUsageWorkspaceID(),
		Provider:            "slack",
		ExternalWorkspaceID: "T1",
		ExternalEventID:     "Ev1",
		AttemptCount:        attempt,
		Usage:               usage,
	}
}
