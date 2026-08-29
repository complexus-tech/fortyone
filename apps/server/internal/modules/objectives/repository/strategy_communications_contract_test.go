package objectivesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrategyCommunicationQueriesUseApplicationOwnedTimeAndBoundedKeysets(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/strategy_communications.sql")
	require.NoError(t, err)
	query := string(contents)
	upperQuery := strings.ToUpper(query)

	for _, contract := range []string{
		"NOT CAST(sqlc.arg(has_cursor) AS boolean)",
		"membership.workspace_id > sqlc.arg(after_workspace_id)",
		"user_account.user_id > sqlc.arg(after_user_id)",
		"ORDER BY membership.workspace_id, user_account.user_id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
		"CAST(sqlc.arg(period_start) AS date)",
		"CAST(sqlc.arg(period_end) AS date)",
		"CAST(sqlc.arg(stale_before) AS timestamptz)",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, upperQuery, "CURRENT_DATE")
	require.NotContains(t, upperQuery, "CURRENT_TIMESTAMP")
	require.NotContains(t, upperQuery, "NOW()")
	require.NotContains(t, upperQuery, " OFFSET ")
	require.NotContains(t, query, "pg_timezone_names")
	require.NotContains(t, strings.ToLower(query), "select *")
}

func TestStrategyWeeklyRecipientQueryPagesMembersBeforeLoadingSignals(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/strategy_communications.sql")
	require.NoError(t, err)
	query := string(contents)
	start := strings.Index(query, "-- name: ListStrategyWeeklyCommunicationRecipients")
	end := strings.Index(query, "-- name: ListStrategyWeeklyCommunicationRecords")
	require.NotEqual(t, -1, start)
	require.Greater(t, end, start)
	recipientQuery := query[start:end]

	for _, contract := range []string{
		"membership.role IN ('admin', 'member', 'guest')",
		"user_account.is_active = true",
		"user_account.is_system = false",
		"workspace.deleted_at IS NULL",
		"ORDER BY membership.workspace_id, user_account.user_id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
	} {
		require.Contains(t, recipientQuery, contract)
	}
	require.NotContains(t, recipientQuery, "objectives")
	require.NotContains(t, recipientQuery, "key_results")
	require.NotContains(t, recipientQuery, "stale_before")
}

func TestStrategyWeeklySignalQueryRechecksLeadAccessAndCompletionSemantics(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/strategy_communications.sql")
	require.NoError(t, err)
	query := string(contents)
	start := strings.Index(query, "-- name: ListStrategyWeeklyCommunicationRecords")
	require.NotEqual(t, -1, start)
	signalQuery := query[start:]

	for _, contract := range []string{
		"membership.user_id = sqlc.arg(user_id)",
		"membership.workspace_id = sqlc.arg(workspace_id)",
		"user_account.is_active = true",
		"user_account.is_system = false",
		"workspace.deleted_at IS NULL",
		"objective.lead_user_id = recipient.user_id",
		"recipient.role = 'admin'",
		"FROM public.team_members AS team_membership",
		"team_membership.team_id = objective.team_id",
		"COALESCE(status.category, '') NOT IN ('completed', 'cancelled', 'paused')",
		"key_result.target_value >= key_result.start_value",
		"key_result.current_value >= key_result.target_value",
		"key_result.target_value < key_result.start_value",
		"key_result.current_value <= key_result.target_value",
		"CAST(key_result.measurement_type AS text) = 'boolean'",
		"objective.objective_id > sqlc.arg(after_objective_id)",
		"key_result.id IS NULL THEN 1 ELSE 0",
		"key_result.id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
	} {
		require.Contains(t, signalQuery, contract)
	}
	require.NotContains(t, signalQuery, "key_result.current_value IS DISTINCT FROM key_result.target_value")
}
