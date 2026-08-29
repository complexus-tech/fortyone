package storiesrepository

import (
	"os"
	"strings"
	"testing"
)

func TestStorySubresourceQueriesCarryWorkspaceScope(t *testing.T) {
	data, err := os.ReadFile("queries/support_reads.sql")
	if err != nil {
		t.Fatalf("read support queries: %v", err)
	}
	source := string(data)

	for _, part := range []string{
		"story.workspace_id = sqlc.arg(workspace_id)",
		"workspace_member.workspace_id = story.workspace_id",
		"team_member.team_id = story.team_id",
		"actor.user_id = sqlc.arg(actor_id)",
		"story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))",
		"story.id = activity.story_id",
		"story.id = story_link.story_id",
	} {
		if !strings.Contains(source, part) {
			t.Fatalf("typed story support queries are missing %q", part)
		}
	}

	commentQueries, err := os.ReadFile("queries/comment_reads.sql")
	if err != nil {
		t.Fatalf("read comment queries: %v", err)
	}
	for _, part := range []string{
		"workspace_member.workspace_id = story.workspace_id",
		"team_member.team_id = story.team_id",
		"comment.story_id = sqlc.arg(story_id)",
		"reply.parent_id = ANY",
		"story.workspace_id = sqlc.arg(workspace_id)",
		"story.deleted_at IS NULL",
	} {
		if !strings.Contains(string(commentQueries), part) {
			t.Fatalf("typed comment queries are missing %q", part)
		}
	}
}
