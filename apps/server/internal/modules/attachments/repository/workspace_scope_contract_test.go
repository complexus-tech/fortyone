package attachmentsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestStoryAttachmentReadsScopeStoryAndAttachmentWorkspace(t *testing.T) {
	queriesData, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read queries.go: %v", err)
	}
	queries := string(queriesData)
	start := strings.Index(queries, "func (r *Repository) GetAttachmentsByStoryID(")
	if start < 0 {
		t.Fatal("GetAttachmentsByStoryID is missing")
	}
	listBody := queries[start:]
	for _, part := range []string{"JOIN stories s", "s.workspace_id = :workspace_id", "a.workspace_id = :workspace_id"} {
		if !strings.Contains(listBody, part) {
			t.Fatalf("attachment list query is missing %q", part)
		}
	}

	mediaData, err := os.ReadFile("story_media.go")
	if err != nil {
		t.Fatalf("read story_media.go: %v", err)
	}
	media := string(mediaData)
	start = strings.Index(media, "func (r *Repository) AuthorizeStoryAttachment(")
	if start < 0 {
		t.Fatal("AuthorizeStoryAttachment is missing")
	}
	authorizationBody := media[start:]
	for _, part := range []string{"story_attachments", "sa.story_id = $1", "a.attachment_id = $2", "s.workspace_id = $3", "attachments.ErrNotFound"} {
		if !strings.Contains(authorizationBody, part) {
			t.Fatalf("attachment authorization is missing %q", part)
		}
	}
}
