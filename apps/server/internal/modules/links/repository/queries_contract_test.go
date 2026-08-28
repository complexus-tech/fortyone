package linksrepository

import (
	"os"
	"strings"
	"testing"
)

func TestLinkMutationQueriesAuthorizeAndLockTheActorInsideEachStatement(t *testing.T) {
	source, err := os.ReadFile("queries/links.sql")
	if err != nil {
		t.Fatalf("read links queries: %v", err)
	}

	queryFile := string(source)
	cases := []struct {
		name  string
		start string
		end   string
		parts []string
	}{
		{
			name:  "create",
			start: "-- name: CreateLinkForWorkspace :one",
			end:   "-- name: UpdateLinkForWorkspace :execrows",
			parts: []string{"FROM public.stories AS story", "story.workspace_id = sqlc.arg(workspace_id)"},
		},
		{
			name:  "update",
			start: "-- name: UpdateLinkForWorkspace :execrows",
			end:   "-- name: DeleteLinkForWorkspace :execrows",
			parts: []string{"FROM public.stories AS story, authorized_actor", "story.workspace_id = authorized_actor.workspace_id"},
		},
		{
			name:  "delete",
			start: "-- name: DeleteLinkForWorkspace :execrows",
			parts: []string{"USING public.stories AS story, authorized_actor", "story.workspace_id = authorized_actor.workspace_id"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			section := querySection(t, queryFile, tc.start, tc.end)
			authorizationClauses := []string{
				"FROM public.workspace_members AS member",
				"INNER JOIN public.users AS actor",
				"actor.is_active = TRUE",
				"member.workspace_id = sqlc.arg(workspace_id)",
				"member.user_id = sqlc.arg(actor_id)",
				"member.role IN ('member', 'admin')",
				"FOR UPDATE OF member, actor",
			}
			for _, part := range append(tc.parts, authorizationClauses...) {
				if !strings.Contains(section, part) {
					t.Fatalf("query is missing tenant-scope clause %q", part)
				}
			}
		})
	}
}

func querySection(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("query marker %q not found", start)
	}
	section := source[startIndex:]
	if end == "" {
		return section
	}
	endIndex := strings.Index(section, end)
	if endIndex < 0 {
		t.Fatalf("query marker %q not found", end)
	}
	return section[:endIndex]
}
