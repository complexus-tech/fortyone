package attachmentsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestStoryAttachmentReadsScopeStoryAndAttachmentWorkspace(t *testing.T) {
	queriesData, err := os.ReadFile("queries/attachments.sql")
	if err != nil {
		t.Fatalf("read queries.go: %v", err)
	}
	queries := string(queriesData)
	start := strings.Index(queries, "-- name: ListStoryAttachments")
	if start < 0 {
		t.Fatal("GetAttachmentsByStoryID is missing")
	}
	listBody := queries[start:]
	for _, part := range []string{"INNER JOIN public.stories", "story.workspace_id = sqlc.arg(workspace_id)", "attachment.workspace_id = story.workspace_id"} {
		if !strings.Contains(listBody, part) {
			t.Fatalf("attachment list query is missing %q", part)
		}
	}

	start = strings.Index(queries, "-- name: AuthorizeWorkspaceStoryAttachment")
	if start < 0 {
		t.Fatal("AuthorizeStoryAttachment is missing")
	}
	authorizationBody := queries[start:]
	for _, part := range []string{"story_attachments", "relation.story_id = sqlc.arg(story_id)", "relation.attachment_id = sqlc.arg(attachment_id)", "story.workspace_id = sqlc.arg(workspace_id)"} {
		if !strings.Contains(authorizationBody, part) {
			t.Fatalf("attachment authorization is missing %q", part)
		}
	}
}
