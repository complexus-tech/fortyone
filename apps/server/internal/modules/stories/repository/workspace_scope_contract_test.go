package storiesrepository

import (
	"os"
	"strings"
	"testing"
)

func TestStorySubresourceQueriesCarryWorkspaceScope(t *testing.T) {
	data, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read queries.go: %v", err)
	}
	source := string(data)

	for _, contract := range []struct {
		name  string
		start string
		end   string
		parts []string
	}{
		{name: "links", start: "func (r *repo) GetStoryLinks(", end: "func (r *repo) MyStories(", parts: []string{"INNER JOIN stories", "s.workspace_id = :workspace_id"}},
		{name: "activities", start: "func (r *repo) GetActivitiesWithUser(", end: "func (r *repo) GetComments(", parts: []string{"INNER JOIN stories", "sa.workspace_id = :workspace_id", "s.workspace_id = :workspace_id"}},
		{name: "comments", start: "func (r *repo) GetComments(", end: "func (r *repo) GetComment(", parts: []string{"INNER JOIN stories", "s.workspace_id = :workspace_id", "sub.story_id = sc.story_id"}},
		{name: "comment", start: "func (r *repo) GetComment(", end: "func (r *repo) CountStoriesInWorkspace(", parts: []string{"sc.story_id = :story_id", "s.workspace_id = :workspace_id", "stories.ErrNotFound"}},
	} {
		t.Run(contract.name, func(t *testing.T) {
			body := sourceBetweenMarkers(t, source, contract.start, contract.end)
			for _, part := range contract.parts {
				if !strings.Contains(body, part) {
					t.Fatalf("%s query is missing %q", contract.name, part)
				}
			}
		})
	}
}

func sourceBetweenMarkers(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing %q", start)
	}
	endIndex := strings.Index(source[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("source after %q is missing %q", start, end)
	}
	return source[startIndex : startIndex+len(start)+endIndex]
}
