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

func TestGlobalUserOnboardingTourProgressMigrationBackfillsAndMirrorsLegacyWrites(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000182_global_user_onboarding_tour_progress.up.sql")
	if err != nil {
		t.Fatalf("read global onboarding progress migration: %v", err)
	}
	migration := strings.ToLower(string(data))
	for _, contract := range []string{
		"create table public.user_onboarding_tour_progress_global",
		"primary key (user_id, tour_key, tour_version)",
		"foreign key (user_id) references public.users(user_id) on delete cascade",
		"group by progress.user_id, progress.tour_key, progress.tour_version",
		"bool_or(progress.status = 'completed')",
		"bool_or(progress.status = 'skipped')",
		"array_agg(distinct step.value order by step.value)",
		"array_agg(distinct action.value order by action.value)",
		"where account.has_seen_walkthrough = true",
		"'workspace-getting-started'",
		"where progress.tour_key = 'workspace-module-team'",
		"'workspace-module-sprints'",
		"create trigger mirror_user_onboarding_tour_progress_global",
		"on public.user_onboarding_tour_progress",
		"on conflict (user_id, tour_key, tour_version) do update",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("global onboarding progress migration is missing contract %q", contract)
		}
	}
}

func TestGlobalUserOnboardingTourProgressRollbackPreservesCanonicalState(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000182_global_user_onboarding_tour_progress.down.sql")
	if err != nil {
		t.Fatalf("read global onboarding progress rollback: %v", err)
	}
	rollback := strings.ToLower(string(data))
	for _, contract := range []string{
		"migration 000182 is forward-only",
		"select 1 from public.user_onboarding_tour_progress_global",
		"raise exception",
		"using errcode = '55000'",
	} {
		if !strings.Contains(rollback, contract) {
			t.Errorf("global onboarding progress rollback is missing contract %q", contract)
		}
	}
	if strings.Index(rollback, "raise exception") > strings.Index(rollback, "drop table") {
		t.Fatal("global onboarding rollback must guard before dropping canonical state")
	}
}
