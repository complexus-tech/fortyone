package storiesrepository

import (
	"os"
	"strings"
	"testing"
)

func TestOAuthApplicationSQLAuthorizationIsLimitedToStoryCreation(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("queries/mutations.sql")
	if err != nil {
		t.Fatalf("read story mutation queries: %v", err)
	}
	queries := string(data)
	for _, queryName := range []string{
		"AuthorizeStoryCreate",
		"GetOAuthApplicationStoryCreationReplay",
	} {
		section := namedQuerySection(t, queries, queryName)
		for _, required := range []string{
			"'oauth_application'",
			"oauth_application_installations",
			"oauth_application_installation_scopes",
			"installation_scope.scope = 'stories:write'",
			"principal.status = 'active'",
			"application.expires_at > sqlc.arg(now)",
		} {
			if !strings.Contains(section, required) {
				t.Errorf("%s is missing %q", queryName, required)
			}
		}
	}
	for _, queryName := range []string{
		"GetStoryMutationSnapshot",
		"ApplyStoryPatch",
		"DeleteStoryMutation",
	} {
		section := namedQuerySection(t, queries, queryName)
		if strings.Contains(section, "'oauth_application'") {
			t.Errorf("%s must not authorize OAuth application actors", queryName)
		}
	}
}

func namedQuerySection(t *testing.T, queries, name string) string {
	t.Helper()
	marker := "-- name: " + name + " "
	start := strings.Index(queries, marker)
	if start < 0 {
		t.Fatalf("query %s is missing", name)
	}
	remainder := queries[start+len(marker):]
	end := strings.Index(remainder, "\n-- name: ")
	if end < 0 {
		return remainder
	}
	return remainder[:end]
}
