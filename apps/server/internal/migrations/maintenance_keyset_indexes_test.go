package migrations

import (
	"strings"
	"testing"
)

func TestMaintenanceKeysetIndexesMatchBoundedJobOrdering(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000173_maintenance_keyset_indexes.up.sql")
	if err != nil {
		t.Fatalf("read maintenance keyset index migration: %v", err)
	}
	migration := string(forward)
	for _, contract := range []string{
		"idx_stories_retention_deleted_page",
		"ON public.stories (deleted_at, id)",
		"idx_workspaces_inactivity_warning_page",
		"ON public.workspaces (last_accessed_at, workspace_id)",
		"inactivity_warning_sent_at IS NULL",
		"idx_workspaces_inactivity_deletion_page",
		"inactivity_warning_sent_at IS NOT NULL",
		"idx_workspaces_deleted_retention_page",
		"ON public.workspaces (deleted_at, workspace_id)",
		"idx_users_inactivity_warning_page",
		"ON public.users (last_login_at, user_id)",
		"idx_users_inactivity_deactivation_page",
		"ON public.users (inactivity_warning_sent_at, last_login_at, user_id)",
		"is_active = TRUE",
		"is_system = FALSE",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("maintenance keyset index migration is missing %q", contract)
		}
	}
}

func TestMaintenanceKeysetIndexRollbackDropsOnlyAddedIndexes(t *testing.T) {
	t.Parallel()

	rollback, err := FS.ReadFile("000173_maintenance_keyset_indexes.down.sql")
	if err != nil {
		t.Fatalf("read maintenance keyset index rollback: %v", err)
	}
	migration := string(rollback)
	if strings.Count(migration, "DROP INDEX IF EXISTS") != 6 {
		t.Fatalf("maintenance keyset rollback must drop exactly six indexes:\n%s", migration)
	}
	for _, index := range []string{
		"public.idx_stories_retention_deleted_page",
		"public.idx_workspaces_inactivity_warning_page",
		"public.idx_workspaces_inactivity_deletion_page",
		"public.idx_workspaces_deleted_retention_page",
		"public.idx_users_inactivity_warning_page",
		"public.idx_users_inactivity_deactivation_page",
	} {
		if !strings.Contains(migration, index) {
			t.Errorf("maintenance keyset rollback is missing %q", index)
		}
	}
}
