package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoryAutomationQueriesAreTenantScopedBoundedAndApplicationClocked(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/story_automation.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: lockstoryautomation :exec",
		"hashtextextended(cast(sqlc.arg(lock_name) as text), 0)",
		"-- name: archiveeligiblestoriesbatch :many",
		"-- name: closeeligiblestoriesbatch :many",
		"-- name: migrateeligiblesprintstoriesbatch :many",
		"settings.workspace_id = story.workspace_id",
		"settings.team_id = story.team_id",
		"current_status.workspace_id = story.workspace_id",
		"current_status.team_id = story.team_id",
		"candidate_status.workspace_id = story.workspace_id",
		"candidate_status.team_id = story.team_id",
		"ended_sprint.workspace_id = story.workspace_id",
		"ended_sprint.team_id = story.team_id",
		"candidate_sprint.workspace_id = ended_sprint.workspace_id",
		"candidate_sprint.team_id = ended_sprint.team_id",
		"order by story.updated_at, story.id",
		"order by ended_sprint.end_date, ended_sprint.sprint_id, story.id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"for update of story skip locked",
		"cast(sqlc.arg(as_of) as timestamptz) at time zone 'utc'",
		"story.status_id = candidate.expected_status_id",
		"story.sprint_id = candidate.previous_sprint_id",
		"insert into public.story_activities",
		"insert into public.audit_events",
		"story.auto_moved_to_sprint",
		"previous_sprint_id",
		"new_sprint_id",
	} {
		require.Contains(t, query, contract)
	}

	require.Equal(t, 3, strings.Count(query, "for update of story skip locked"))
	require.Equal(t, 3, strings.Count(query, "limit cast(sqlc.arg(batch_size) as integer)"))
	require.Equal(t, 2, strings.Count(query, "insert into public.story_activities"))

	for _, forbidden := range []string{
		"current_timestamp",
		"current_date",
		"now()",
		" offset ",
		"::",
		"select *",
	} {
		require.NotContains(t, query, forbidden)
	}
}
