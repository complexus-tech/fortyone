package notificationsrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWeeklyDigestRecipientQueryPreservesEligibilityAndPreferenceContracts(t *testing.T) {
	t.Parallel()

	query := weeklyDigestQuerySection(t, "ListWeeklyDigestRecipients", "GetWeeklyDigestStats")
	for _, contract := range []string{
		"recipient.is_active = TRUE",
		"recipient.is_system = FALSE",
		"membership.role IN ('admin', 'member')",
		"workspace.deleted_at IS NULL",
		"NULLIF(BTRIM(recipient.email), '') IS NOT NULL",
		"preference.preferences -> 'weekly_digest' ->> 'email'",
	} {
		require.Contains(t, query, contract)
	}
}

func TestWeeklyDigestRecipientQueryUsesBoundedCompositeKeyset(t *testing.T) {
	t.Parallel()

	query := weeklyDigestQuerySection(t, "ListWeeklyDigestRecipients", "GetWeeklyDigestStats")
	for _, contract := range []string{
		"NOT CAST(sqlc.arg(has_cursor) AS boolean)",
		"workspace.workspace_id > CAST(sqlc.arg(after_workspace_id) AS uuid)",
		"membership.user_id > CAST(sqlc.arg(after_user_id) AS uuid)",
		"ORDER BY workspace.workspace_id, membership.user_id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, strings.ToUpper(query), "OFFSET")
}

func TestWeeklyDigestStatsUseExplicitPeriodAndCurrentAccess(t *testing.T) {
	t.Parallel()

	query := weeklyDigestQuerySection(t, "GetWeeklyDigestStats", "")
	for _, contract := range []string{
		"CAST(sqlc.arg(as_of) AS timestamptz)",
		"AT TIME ZONE 'UTC'",
		"AND EXISTS (SELECT 1 FROM recipient_access)",
		"team.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)",
		"EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')",
		"story.team_id IN (SELECT team_id FROM visible_teams)",
		"objective.team_id IN (SELECT team_id FROM visible_teams)",
		"notification.message -> 'strategy' ->> 'kind' = 'weekly_check_in'",
		"notification.created_at >= period.window_start",
		"notification.created_at <= period.as_of",
		"comment.created_at >= period.window_start",
		"comment.created_at <= period.as_of",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, query, "CURRENT_DATE")
	require.NotContains(t, query, "NOW()")
	require.NotContains(t, strings.ToLower(query), "select *")
}

func readWeeklyDigestQuerySource(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile("queries/weekly_digest.sql")
	require.NoError(t, err)
	return string(contents)
}

func weeklyDigestQuerySection(t *testing.T, queryName, nextQueryName string) string {
	t.Helper()
	source := readWeeklyDigestQuerySource(t)
	marker := "-- name: " + queryName + " "
	start := strings.Index(source, marker)
	require.GreaterOrEqual(t, start, 0, "query %s not found", queryName)
	section := source[start:]
	if nextQueryName == "" {
		return section
	}
	nextMarker := "-- name: " + nextQueryName + " "
	end := strings.Index(section[len(marker):], nextMarker)
	require.GreaterOrEqual(t, end, 0, "query %s not found", nextQueryName)
	return section[:len(marker)+end]
}
