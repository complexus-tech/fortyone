package migrations

import (
	"strings"
	"testing"
)

func TestAutomationKeysetIndexesMatchScheduledQueryShapes(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000175_automation_keyset_indexes.up.sql")
	if err != nil {
		t.Fatalf("read automation index migration: %v", err)
	}
	normalized := strings.ToLower(string(forward))
	for _, contract := range []string{
		"idx_stories_automation_updated_page",
		"on public.stories (updated_at, id)",
		"where deleted_at is null",
		"and archived_at is null",
		"idx_team_sprint_settings_auto_create_page",
		"on public.team_sprint_settings (workspace_id, team_id)",
		"include (updated_at)",
		"where auto_create_sprints = true",
		"idx_sprints_automation_end_page",
		"on public.sprints (end_date, sprint_id, workspace_id, team_id)",
	} {
		if !strings.Contains(normalized, contract) {
			t.Errorf("automation index migration is missing %q", contract)
		}
	}
}

func TestAutomationKeysetIndexesRollbackEveryAddedIndex(t *testing.T) {
	t.Parallel()

	rollback, err := FS.ReadFile("000175_automation_keyset_indexes.down.sql")
	if err != nil {
		t.Fatalf("read automation index rollback: %v", err)
	}
	normalized := strings.ToLower(string(rollback))
	for _, index := range []string{
		"idx_stories_automation_updated_page",
		"idx_team_sprint_settings_auto_create_page",
		"idx_sprints_automation_end_page",
	} {
		if !strings.Contains(normalized, "drop index if exists public."+index) {
			t.Errorf("automation index rollback does not drop %s", index)
		}
	}
}
