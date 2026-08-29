package reportsrepository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportQueriesUseTypedStaticFilters(t *testing.T) {
	t.Parallel()

	queryFiles, err := filepath.Glob(filepath.Join("queries", "*.sql"))
	if err != nil {
		t.Fatalf("glob report queries: %v", err)
	}
	if len(queryFiles) == 0 {
		t.Fatal("expected module-owned report query files")
	}

	var combined strings.Builder
	for _, queryFile := range queryFiles {
		contents, err := os.ReadFile(queryFile)
		if err != nil {
			t.Fatalf("read %s: %v", queryFile, err)
		}
		combined.Write(contents)
	}
	queries := combined.String()

	for _, expected := range []string{
		"sqlc.arg(team_ids)::uuid[]",
		"sqlc.arg(assignee_ids)::uuid[]",
		"sqlc.arg(sprint_ids)::uuid[]",
		"sqlc.arg(objective_ids)::uuid[]",
		"sqlc.narg(start_date)::timestamptz",
		"sqlc.narg(end_date)::timestamptz",
	} {
		if !strings.Contains(queries, expected) {
			t.Errorf("expected typed filter expression %q", expected)
		}
	}
	for _, forbidden := range []string{"fmt.Sprintf", "IN (%s)", ":workspace_id"} {
		if strings.Contains(queries, forbidden) {
			t.Errorf("unexpected dynamic query fragment %q", forbidden)
		}
	}
}

func TestReportAuthorizationQueryRequiresAuthorizedRoleAndVisibleTeamScope(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join("queries", "authorization.sql"))
	if err != nil {
		t.Fatalf("read authorization query: %v", err)
	}
	query := string(contents)
	for _, expected := range []string{
		"FROM workspace_members AS membership",
		"actor.is_active = TRUE",
		"actor.is_system = FALSE",
		"workspace.deleted_at IS NULL",
		"membership.workspace_id = sqlc.arg(workspace_id)::uuid",
		"membership.user_id = sqlc.arg(actor_id)::uuid",
		"membership.role IN ('admin', 'member')",
		"membership.role = 'admin'",
		"team.is_private = FALSE",
		"FROM team_members AS actor_team_membership",
		"cardinality(sqlc.arg(requested_team_ids)::uuid[]) = 0",
	} {
		if !strings.Contains(query, expected) {
			t.Errorf("authorization query missing %q", expected)
		}
	}
}
