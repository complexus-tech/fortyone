package storiesrepository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMyStoriesAssignmentAttributionUsesAuditHistory(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/stories.sql")
	require.NoError(t, err)
	query := string(contents)

	for _, contract := range []string{
		"activity.field_changed = 'assignee_id'",
		"activity.new_value = to_jsonb(story.assignee_id)",
		"event.event_type = 'story.created'",
		"event.payload ->> 'assigneeId' = CAST(story.assignee_id AS text)",
		"ORDER BY candidate.assigned_at DESC, candidate.source_priority",
		"assignment_actor.user_id AS assigned_by_id",
	} {
		require.Contains(t, query, contract)
	}
}
