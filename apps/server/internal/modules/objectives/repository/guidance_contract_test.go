package objectivesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectiveGuidanceQueriesPreserveEligibilityAndAccessContracts(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	for _, contract := range []string{
		"membership.role IN ('admin', 'member', 'guest')",
		"membership.role = 'admin'",
		"FROM team_members AS team_membership",
		"team_membership.team_id = objective.team_id",
		"workspace.deleted_at IS NULL",
		"lead.is_active = true",
		"lead.is_system = false",
		"settings.objective_enabled = true",
		"preference.preferences -> 'reminders' ->> 'email'",
		"key_result.target_value >= key_result.start_value",
		"key_result.current_value >= key_result.target_value",
		"key_result.target_value < key_result.start_value",
		"key_result.current_value <= key_result.target_value",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, query, "JOIN team_members AS team_membership")
}

func TestObjectiveGuidanceRecipientQueryUsesBoundedCompositeKeyset(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	for _, contract := range []string{
		"NOT CAST(sqlc.arg(has_cursor) AS boolean)",
		"objective.lead_user_id > sqlc.arg(after_lead_user_id)",
		"objective.workspace_id > sqlc.arg(after_workspace_id)",
		"ORDER BY objective.lead_user_id, workspace.workspace_id",
		"LIMIT CAST(sqlc.arg(result_limit) AS integer)",
	} {
		require.Contains(t, query, contract)
	}
	require.NotContains(t, strings.ToUpper(query), "OFFSET")
}

func TestObjectiveGuidanceDetailQueryIsLeadAndWorkspaceScoped(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/guidance.sql")
	require.NoError(t, err)
	query := string(contents)

	require.Contains(t, query, "objective.lead_user_id = sqlc.arg(lead_user_id)")
	require.Contains(t, query, "objective.workspace_id = sqlc.arg(workspace_id)")
	require.Contains(t, query, "jsonb_array_length(key_results) > 0")
	require.Contains(t, query, "CAST(sqlc.arg(as_of) AS date)")
	require.NotContains(t, query, "CURRENT_DATE")
	require.NotContains(t, strings.ToLower(query), "select *")
}
