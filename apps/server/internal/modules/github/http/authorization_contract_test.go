package githubhttp

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestWorkspaceGitHubMutationRoutesRequireExplicitRoles(t *testing.T) {
	t.Parallel()

	routes, err := os.Open("routes.go")
	if err != nil {
		t.Fatalf("open GitHub routes: %v", err)
	}
	defer routes.Close()

	expectedRoles := map[string]string{
		"/integrations/github/install-session":              "adminOnly",
		"/integrations/github/repositories/resync":          "adminOnly",
		"/integrations/github/settings\"":                   "adminOnly",
		"/integrations/github/issue-sync-links\"":           "adminOnly",
		"/integrations/github/issue-sync-links/{linkId}":    "adminOnly",
		"/teams/{teamId}/settings/github":                   "adminOnly",
		"/stories/{storyId}/github-links/{linkId}":          "memberOnly",
		"/stories/{storyId}/github-comments":                "memberOnly",
		"/integration-requests/{requestId}/github-comments": "memberOnly",
	}
	found := make(map[string]bool, len(expectedRoles))
	scanner := bufio.NewScanner(routes)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "app.Post(") && !strings.Contains(line, "app.Put(") && !strings.Contains(line, "app.Delete(") {
			continue
		}
		for route, role := range expectedRoles {
			if strings.Contains(line, route) {
				if !strings.Contains(line, role) {
					t.Fatalf("GitHub mutation route %q is missing %s: %s", route, role, line)
				}
				found[route] = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan GitHub routes: %v", err)
	}
	for route := range expectedRoles {
		if !found[route] {
			t.Fatalf("GitHub mutation route %q was not checked", route)
		}
	}
}
