package teamsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestTeamQueriesRetainWorkspaceActorAndMembershipScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		file  string
		parts []string
	}{
		{
			name: "team reads",
			file: "queries/teams.sql",
			parts: []string{
				"team.workspace_id = sqlc.arg(workspace_id)",
				"actor_membership.workspace_id = team.workspace_id",
				"actor_membership.user_id = sqlc.arg(actor_id)",
				"actor.is_active = TRUE",
				"actor_team_membership.team_id = team.team_id",
			},
		},
		{
			name: "membership commands",
			file: "queries/memberships.sql",
			parts: []string{
				"workspace_membership.workspace_id = team.workspace_id",
				"member.is_active = TRUE",
				"team.workspace_id = sqlc.arg(workspace_id)",
				"team.is_private = FALSE",
				"team_membership.user_id = sqlc.arg(actor_id)",
				"target_workspace_membership.workspace_id = team.workspace_id",
				"target_user.is_active = TRUE",
			},
		},
		{
			name: "ordering commands",
			file: "queries/orderings.sql",
			parts: []string{
				"actor_membership.workspace_id = sqlc.arg(workspace_id)",
				"actor.is_active = TRUE",
				"team.workspace_id = sqlc.arg(workspace_id)",
				"actor_team_membership.user_id = sqlc.arg(actor_id)",
				"admin_membership.role = 'admin'",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(test.file)
			if err != nil {
				t.Fatalf("read %s: %v", test.file, err)
			}
			for _, part := range test.parts {
				if !strings.Contains(string(source), part) {
					t.Fatalf("%s is missing security clause %q", test.file, part)
				}
			}
		})
	}
}

func TestTeamQueryFilesContainNoWildcardProjection(t *testing.T) {
	t.Parallel()

	for _, file := range []string{"queries/teams.sql", "queries/memberships.sql", "queries/orderings.sql"} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(strings.ToUpper(string(source)), "SELECT *") {
			t.Fatalf("%s contains a wildcard projection", file)
		}
	}
}
