package developeroauthhttp

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestOAuthApplicationManagementRoutesKeepTheFullAuthorizationChain(t *testing.T) {
	t.Parallel()

	routes, err := os.Open("routes.go")
	if err != nil {
		t.Fatalf("open developer OAuth routes: %v", err)
	}
	defer routes.Close()

	expected := map[string]bool{
		`app.Post("/workspaces/{workspaceSlug}/oauth-applications"`:                                      true,
		`app.Get("/workspaces/{workspaceSlug}/oauth-applications"`:                                       false,
		`app.Get("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets"`:               false,
		`app.Post("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets/rotate"`:       true,
		`app.Delete("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets/{secretId}"`: true,
		`app.Post("/workspaces/{workspaceSlug}/oauth-application-installations"`:                         true,
		`app.Get("/workspaces/{workspaceSlug}/oauth-application-installations"`:                          false,
		`app.Put("/workspaces/{workspaceSlug}/oauth-application-installations/{installationId}"`:         true,
		`app.Delete("/workspaces/{workspaceSlug}/oauth-application-installations/{installationId}"`:      true,
	}
	found := make(map[string]bool, len(expected))
	scanner := bufio.NewScanner(routes)
	for scanner.Scan() {
		line := scanner.Text()
		for route, mutation := range expected {
			if !strings.Contains(line, route) {
				continue
			}
			for _, middleware := range []string{"auth", "workspace", "adminOnly", "integrationScope"} {
				if !strings.Contains(line, middleware) {
					t.Fatalf("OAuth management route is missing %s: %s", middleware, line)
				}
			}
			if mutation && !strings.Contains(line, "mutationLimit") {
				t.Fatalf("OAuth management mutation is missing mutationLimit: %s", line)
			}
			found[route] = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan developer OAuth routes: %v", err)
	}
	for route := range expected {
		if !found[route] {
			t.Fatalf("OAuth management route %q was not registered", route)
		}
	}
}
