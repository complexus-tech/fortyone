package teamsettingsrepository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTeamSettingsQueriesRetainTenantAndTeamScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		file  string
		parts []string
	}{
		{
			file: "queries/authorization.sql",
			parts: []string{
				"team.workspace_id = sqlc.arg(workspace_id)",
				"team_member.user_id = sqlc.arg(actor_id)",
				"workspace_member.workspace_id = team.workspace_id",
				"actor.is_active = TRUE",
			},
		},
		{
			file: "queries/sprint_settings.sql",
			parts: []string{
				"team.team_id = settings.team_id",
				"team.workspace_id = settings.workspace_id",
				"settings.team_id = sqlc.arg(team_id)",
				"settings.workspace_id = sqlc.arg(workspace_id)",
			},
		},
		{
			file: "queries/story_automation_settings.sql",
			parts: []string{
				"team.team_id = settings.team_id",
				"team.workspace_id = settings.workspace_id",
				"settings.team_id = sqlc.arg(team_id)",
				"settings.workspace_id = sqlc.arg(workspace_id)",
			},
		},
		{
			file: "queries/estimation_settings.sql",
			parts: []string{
				"team.team_id = settings.team_id",
				"team.workspace_id = settings.workspace_id",
				"settings.team_id = sqlc.arg(team_id)",
				"settings.workspace_id = sqlc.arg(workspace_id)",
			},
		},
		{
			file: "queries/sprint_schedule.sql",
			parts: []string{
				"sprint.team_id = sqlc.arg(team_id)",
				"sprint.workspace_id = sqlc.arg(workspace_id)",
				"team.workspace_id = sprint.workspace_id",
				"sprint.schedule_managed_by_automation = TRUE",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.file, func(t *testing.T) {
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

func TestTeamSettingsQueriesUseExplicitProjections(t *testing.T) {
	t.Parallel()

	queryFiles, err := filepath.Glob("queries/*.sql")
	if err != nil {
		t.Fatalf("list query files: %v", err)
	}
	if len(queryFiles) == 0 {
		t.Fatal("no team settings query files found")
	}
	for _, queryFile := range queryFiles {
		source, err := os.ReadFile(queryFile)
		if err != nil {
			t.Fatalf("read %s: %v", queryFile, err)
		}
		if strings.Contains(strings.ToUpper(string(source)), "SELECT *") {
			t.Fatalf("%s contains a wildcard projection", queryFile)
		}
	}
}

func TestTeamSettingsRepositoryContainsNoSQLXOrHandwrittenSQL(t *testing.T) {
	t.Parallel()

	productionFiles, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list repository files: %v", err)
	}
	for _, productionFile := range productionFiles {
		if strings.HasSuffix(productionFile, "_test.go") {
			continue
		}
		source, err := os.ReadFile(productionFile)
		if err != nil {
			t.Fatalf("read %s: %v", productionFile, err)
		}
		text := string(source)
		if strings.Contains(text, "jmoiron/sqlx") {
			t.Fatalf("%s imports SQLx", productionFile)
		}
		for _, marker := range []string{"`SELECT ", "`INSERT ", "`UPDATE ", "`DELETE "} {
			if strings.Contains(strings.ToUpper(text), marker) {
				t.Fatalf("%s contains handwritten SQL marker %q", productionFile, marker)
			}
		}
	}
}
