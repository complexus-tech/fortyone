package migrations

import (
	"strings"
	"testing"
)

func TestUserOnboardingTourProgressMigrationScopesVersionedState(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000176_user_onboarding_tour_progress.up.sql")
	if err != nil {
		t.Fatalf("read onboarding progress migration: %v", err)
	}
	migration := strings.ToLower(string(data))
	for _, contract := range []string{
		"create table public.user_onboarding_tour_progress",
		"primary key (user_id, workspace_id, tour_key, tour_version)",
		"foreign key (user_id) references public.users(user_id) on delete cascade",
		"foreign key (workspace_id) references public.workspaces(workspace_id) on delete cascade",
		"completed_step_ids text[] not null default array[]::text[]",
		"completed_action_ids text[] not null default array[]::text[]",
		"status in ('active', 'completed', 'skipped')",
		"idx_user_onboarding_tour_progress_workspace_id",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("onboarding progress migration is missing contract %q", contract)
		}
	}
}

func TestUserOnboardingTourProgressRollbackPreservesDurableState(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000176_user_onboarding_tour_progress.down.sql")
	if err != nil {
		t.Fatalf("read onboarding progress rollback: %v", err)
	}
	rollback := strings.ToLower(string(data))
	for _, contract := range []string{
		"migration 000176 is forward-only",
		"select 1 from public.user_onboarding_tour_progress",
		"raise exception",
		"using errcode = '55000'",
	} {
		if !strings.Contains(rollback, contract) {
			t.Errorf("onboarding progress rollback is missing contract %q", contract)
		}
	}
	if strings.Index(rollback, "raise exception") > strings.Index(rollback, "drop table") {
		t.Fatal("onboarding progress rollback must guard before dropping the table")
	}
}
