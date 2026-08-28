package invitationshttp

import (
	"os"
	"strings"
	"testing"
)

func TestWorkspaceInvitationAdministrationRoutesRequireAdmin(t *testing.T) {
	t.Parallel()

	routes, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read invitation routes: %v", err)
	}

	for _, route := range []string{
		`app.Post("/workspaces/{workspaceSlug}/invitations"`,
		`app.Get("/workspaces/{workspaceSlug}/invitations"`,
		`app.Delete("/workspaces/{workspaceSlug}/invitations/{id}"`,
	} {
		start := strings.Index(string(routes), route)
		if start == -1 {
			t.Fatalf("invitation administration route %q was not registered", route)
		}
		lineEnd := strings.IndexByte(string(routes)[start:], '\n')
		if lineEnd == -1 {
			lineEnd = len(routes) - start
		}
		line := string(routes)[start : start+lineEnd]
		if !strings.Contains(line, "adminOnly") {
			t.Fatalf("invitation administration route is missing adminOnly: %s", line)
		}
	}
}
