package messagingrepository

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"testing"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type dailyUsageQueriesStub struct {
	claimResult int32
	claimErr    error
	addResult   messagingsql.AddAssistantDailyUsageRow
	addErr      error
	getResults  []messagingsql.GetAssistantDailyUsageRow
	getErr      error
	claims      []messagingsql.ClaimAssistantUsageEventParams
	adds        []messagingsql.AddAssistantDailyUsageParams
	gets        []messagingsql.GetAssistantDailyUsageParams
}

func (stub *dailyUsageQueriesStub) ClaimAssistantUsageEvent(
	_ context.Context,
	params messagingsql.ClaimAssistantUsageEventParams,
) (int32, error) {
	stub.claims = append(stub.claims, params)
	return stub.claimResult, stub.claimErr
}

func (stub *dailyUsageQueriesStub) AddAssistantDailyUsage(
	_ context.Context,
	params messagingsql.AddAssistantDailyUsageParams,
) (messagingsql.AddAssistantDailyUsageRow, error) {
	stub.adds = append(stub.adds, params)
	return stub.addResult, stub.addErr
}

func (stub *dailyUsageQueriesStub) GetAssistantDailyUsage(
	_ context.Context,
	params messagingsql.GetAssistantDailyUsageParams,
) (messagingsql.GetAssistantDailyUsageRow, error) {
	stub.gets = append(stub.gets, params)
	if len(stub.getResults) == 0 {
		return messagingsql.GetAssistantDailyUsageRow{}, stub.getErr
	}
	row := stub.getResults[0]
	stub.getResults = stub.getResults[1:]
	return row, stub.getErr
}

func TestDailyUsageRepositoryCheckUsesDefaultCeiling(t *testing.T) {
	t.Parallel()

	queries := &dailyUsageQueriesStub{getResults: []messagingsql.GetAssistantDailyUsageRow{{
		InputTokens: 800, OutputTokens: 200, TotalTokens: 1_000, RequestCount: 2,
	}}}
	repository := newDailyUsageRepositoryWithQueries(queries)

	snapshot, err := repository.Check(t.Context(), dailyUsageWorkspaceID(), 0)

	require.NoError(t, err)
	require.True(t, snapshot.Allowed)
	require.Equal(t, DefaultDailyWorkspaceTokenLimit, snapshot.Limit)
	require.Equal(t, DefaultDailyWorkspaceTokenLimit-1_000, snapshot.Remaining)
	require.Equal(t, dailyUsageWorkspaceID(), queries.gets[0].WorkspaceID)
}

func TestDailyUsageRepositoryCheckRejectsExhaustedWorkspace(t *testing.T) {
	t.Parallel()

	queries := &dailyUsageQueriesStub{getResults: []messagingsql.GetAssistantDailyUsageRow{{
		InputTokens: 750, OutputTokens: 250, TotalTokens: 1_000, RequestCount: 3,
	}}}
	repository := newDailyUsageRepositoryWithQueries(queries)

	snapshot, err := repository.Check(t.Context(), dailyUsageWorkspaceID(), 1_000)

	require.ErrorIs(t, err, ErrDailyWorkspaceTokenLimit)
	require.False(t, snapshot.Allowed)
	require.Zero(t, snapshot.Remaining)
	var limitError *DailyTokenLimitError
	require.ErrorAs(t, err, &limitError)
	require.Equal(t, int64(1_000), limitError.Used)
}

func TestDailyUsageRepositoryRecordPersistsTypedUsage(t *testing.T) {
	t.Parallel()

	queries := &dailyUsageQueriesStub{
		claimResult: 1,
		addResult: messagingsql.AddAssistantDailyUsageRow{
			InputTokens: 1_100, OutputTokens: 400, TotalTokens: 1_500, RequestCount: 4,
		},
	}
	repository := newDailyUsageRepositoryWithQueries(queries)
	usage := messaging.Usage{InputTokens: 300, OutputTokens: 100, TotalTokens: 400}

	snapshot, err := repository.Record(t.Context(), dailyUsageRecordInput(4, usage), 2_000)

	require.NoError(t, err)
	require.Equal(t, int64(1_500), snapshot.TotalTokens)
	require.True(t, snapshot.Allowed)
	require.Len(t, queries.claims, 1)
	require.Len(t, queries.adds, 1)
	require.Equal(t, int32(4), queries.claims[0].AttemptCount)
	require.Equal(t, int64(300), queries.claims[0].InputTokens)
	require.Equal(t, int64(400), queries.adds[0].TotalTokens)
}

func TestDailyUsageRepositoryDuplicateExecutionReturnsCurrentSnapshot(t *testing.T) {
	t.Parallel()

	queries := &dailyUsageQueriesStub{
		claimErr: pgx.ErrNoRows,
		getResults: []messagingsql.GetAssistantDailyUsageRow{{
			InputTokens: 900, OutputTokens: 100, TotalTokens: 1_000, RequestCount: 3,
		}},
	}
	repository := newDailyUsageRepositoryWithQueries(queries)

	snapshot, err := repository.Record(t.Context(), dailyUsageRecordInput(3, messaging.Usage{
		InputTokens: 300, OutputTokens: 100, TotalTokens: 400,
	}), 2_000)

	require.NoError(t, err)
	require.Equal(t, int64(1_000), snapshot.TotalTokens)
	require.Len(t, queries.claims, 1)
	require.Empty(t, queries.adds)
	require.Len(t, queries.gets, 1)
}

func TestDailyUsageRepositoryRejectsInvalidInputBeforeQueries(t *testing.T) {
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
		{name: "attempt overflow", mutate: func(input *DailyUsageRecordInput) { input.AttemptCount = math.MaxInt32 + 1 }},
		{name: "mismatched totals", mutate: func(input *DailyUsageRecordInput) { input.Usage.TotalTokens++ }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			queries := &dailyUsageQueriesStub{}
			repository := newDailyUsageRepositoryWithQueries(queries)
			input := dailyUsageRecordInput(1, messaging.Usage{InputTokens: 1, OutputTokens: 1, TotalTokens: 2})
			test.mutate(&input)

			_, err := repository.Record(t.Context(), input, 1_000)

			require.Error(t, err)
			require.Empty(t, queries.claims)
			require.Empty(t, queries.adds)
			require.Empty(t, queries.gets)
		})
	}
}

func TestDailyUsageRepositoryWrapsQueryFailures(t *testing.T) {
	t.Parallel()

	queries := &dailyUsageQueriesStub{claimErr: errors.New("database unavailable")}
	repository := newDailyUsageRepositoryWithQueries(queries)

	_, err := repository.Record(t.Context(), dailyUsageRecordInput(1, messaging.Usage{
		InputTokens: 1, OutputTokens: 1, TotalTokens: 2,
	}), 2_000)

	require.ErrorContains(t, err, "claim messaging assistant usage event")
	require.Empty(t, queries.adds)
}

func TestDailyUsageQueriesAreNamedStaticSQL(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/daily_usage.sql")
	require.NoError(t, err)
	queryText := string(source)
	require.Contains(t, queryText, "-- name: ClaimAssistantUsageEvent :one")
	require.Contains(t, queryText, "-- name: AddAssistantDailyUsage :one")
	require.Contains(t, queryText, "-- name: GetAssistantDailyUsage :one")
	require.NotContains(t, strings.ToLower(queryText), "fmt.sprintf")
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
