package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoryGuidanceQueriesPreserveEligibilityAndAccessContracts(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	for _, contract := range []string{
		"membership.role IN ('admin', 'member', 'guest')",
		"membership.role = 'admin'",
		"FROM team_members AS team_membership",
		"team_membership.team_id = story.team_id",
		"workspace.deleted_at IS NULL",
		"story.deleted_at IS NULL",
		"story.archived_at IS NULL",
		"story.completed_at IS NULL",
		"assignee.is_active = true",
		"assignee.is_system = false",
		"preference.preferences -> 'reminders' ->> 'email'",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, query, "JOIN team_members AS team_membership")
}

func TestStoryGuidanceRecipientQueryUsesBoundedCompositeKeyset(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	for _, contract := range []string{
		"NOT CAST(sqlc.arg(has_cursor) AS boolean)",
		"story.assignee_id > sqlc.arg(after_assignee_id)",
		"story.workspace_id > sqlc.arg(after_workspace_id)",
		"ORDER BY story.assignee_id, workspace.workspace_id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, strings.ToUpper(query), "OFFSET")
}

func TestStoryGuidanceDetailQueryIsAssigneeAndWorkspaceScoped(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	require.Contains(t, query, "story.assignee_id = sqlc.arg(assignee_id)")
	require.Contains(t, query, "story.workspace_id = sqlc.arg(workspace_id)")
	require.Contains(t, query, "story.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '3 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '3 days'")
	require.NotContains(t, query, "CURRENT_DATE")
	require.NotContains(t, strings.ToLower(query), "select *")
}
