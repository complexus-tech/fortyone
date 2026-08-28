package invitationsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestInvitationAdminLockRequiresActiveAdminAndLocksAuthorizationRows(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/invitations.sql")
	if err != nil {
		t.Fatalf("read invitation queries: %v", err)
	}
	query := queryContractSection(
		t,
		string(source),
		"-- name: LockActiveWorkspaceAdmin :one",
		"-- name: ListWorkspaceInvitations :many",
	)
	for _, clause := range []string{
		"FROM public.workspace_members AS member",
		"INNER JOIN public.users AS actor",
		"actor.is_active = TRUE",
		"member.workspace_id = sqlc.arg(workspace_id)",
		"member.user_id = sqlc.arg(actor_id)",
		"member.role = 'admin'",
		"FOR UPDATE OF member, actor",
	} {
		if !strings.Contains(query, clause) {
			t.Fatalf("admin lock query is missing %q", clause)
		}
	}
}

func queryContractSection(t *testing.T, source, start, end string) string {
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
