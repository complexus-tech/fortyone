package storieshttp

import (
	"os"
	"strings"
	"testing"
)

func TestSingleDeleteBindsAuthenticatedActorAndWorkspaceRole(t *testing.T) {
	data, err := os.ReadFile("mutation_handlers.go")
	if err != nil {
		t.Fatalf("read story handlers: %v", err)
	}
	body := sourceBetweenMarkers(t, string(data), "func (h *Handlers) Delete(", "func (h *Handlers) BulkDelete(")

	for _, part := range []string{
		"mid.GetUserID(ctx)",
		"ActorID: userID",
		"IsAdmin: workspace.UserRole == string(mid.RoleAdmin)",
		"h.stories.Delete(ctx, storyId, workspace.ID, authorization)",
	} {
		if !strings.Contains(body, part) {
			t.Fatalf("single story deletion authorization is missing %q", part)
		}
	}
}

func sourceBetweenMarkers(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source marker %q not found", start)
	}
	endIndex := strings.Index(source[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("source marker %q not found after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endIndex]
}
