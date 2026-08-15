package migrations

import (
	"strings"
	"testing"
)

const continuousCalendarSchedulingMigration = "000131_continuous_calendar_scheduling"

func TestContinuousCalendarSchedulingForwardMigration(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(continuousCalendarSchedulingMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	t.Run("account scoped connection lifecycle", func(t *testing.T) {
		for _, contract := range []string{
			"DROP CONSTRAINT IF EXISTS calendar_connections_workspace_member_fkey",
			"DROP CONSTRAINT IF EXISTS calendar_connections_workspace_id_fkey",
			"ADD COLUMN cleanup_pending_at timestamptz",
			"DROP CONSTRAINT calendar_events_connection_scope_fkey",
			"ON UPDATE CASCADE",
			"CREATE FUNCTION public.reanchor_account_calendar_connection()",
			"BEFORE DELETE ON public.workspace_members",
			"pg_advisory_xact_lock",
			"CONCAT('calendar:', CAST(OLD.user_id AS text))",
			"member.workspace_id <> OLD.workspace_id",
			"SET cleanup_pending_at = CURRENT_TIMESTAMP",
			"FROM public.calendar_schedule_event_outbox outbox",
			"last_error = 'Retrying provider cleanup after final workspace removal.'",
			"UPDATE public.calendar_busy_windows busy_window",
			"SET workspace_id = replacement_workspace_id",
			"SET workspace_id = replacement_workspace_id",
			"DELETE FROM public.calendar_connections connection",
		} {
			if !strings.Contains(migration, contract) {
				t.Errorf("migration is missing account calendar lifecycle contract %q", contract)
			}
		}

		busyWindowsReanchor := strings.Index(migration, "UPDATE public.calendar_busy_windows busy_window")
		connectionReanchor := strings.LastIndex(migration, "UPDATE public.calendar_connections connection")
		if busyWindowsReanchor < 0 || connectionReanchor < 0 || busyWindowsReanchor >= connectionReanchor {
			t.Fatal("busy windows must move before the connection composite scope cascades to detailed events")
		}
		outboxTable := strings.Index(migration, "CREATE TABLE public.calendar_schedule_event_outbox")
		reanchorFunction := strings.Index(migration, "CREATE FUNCTION public.reanchor_account_calendar_connection()")
		if outboxTable < 0 || reanchorFunction < 0 || outboxTable >= reanchorFunction {
			t.Fatal("the outbox relation must exist before the reanchor trigger function references it")
		}
	})

	t.Run("Maya ownership and provider cleanup", func(t *testing.T) {
		for _, contract := range []string{
			"FOREIGN KEY (story_id, workspace_id)",
			"REFERENCES public.stories(id, workspace_id) ON DELETE CASCADE",
			"FOREIGN KEY (workspace_id, user_id)",
			"REFERENCES public.workspace_members(workspace_id, user_id) ON DELETE CASCADE",
			"CREATE TRIGGER calendar_schedule_blocks_enqueue_provider_delete",
			"BEFORE DELETE ON public.calendar_schedule_blocks",
			"OLD.external_event_id",
			"CREATE TRIGGER stories_cleanup_retired_maya_schedule",
			"AFTER UPDATE OF deleted_at, archived_at ON public.stories",
			"CREATE TRIGGER team_members_cleanup_maya_schedule",
			"AFTER DELETE ON public.team_members",
			"recovery_attempted_at timestamptz",
			"idx_calendar_maya_schedule_ownerships_recovery",
			"(updated_at, workspace_id, story_id)",
			"idx_maya_agent_runs_schedule_recovery",
			"WHERE status = 'running'",
		} {
			if !strings.Contains(migration, contract) {
				t.Errorf("migration is missing Maya lifecycle contract %q", contract)
			}
		}

		providerDeleteTrigger := strings.Index(migration, "BEFORE DELETE ON public.calendar_schedule_blocks")
		storyCleanupTrigger := strings.Index(migration, "AFTER UPDATE OF deleted_at, archived_at ON public.stories")
		teamCleanupTrigger := strings.Index(migration, "AFTER DELETE ON public.team_members")
		if providerDeleteTrigger < 0 || storyCleanupTrigger < 0 || teamCleanupTrigger < 0 || providerDeleteTrigger >= storyCleanupTrigger || providerDeleteTrigger >= teamCleanupTrigger {
			t.Fatal("provider deletion must be captured from each OLD schedule block before lifecycle cleanup triggers can remove it")
		}
	})

	t.Run("durable outbox", func(t *testing.T) {
		for _, contract := range []string{
			"dead_lettered_at timestamptz",
			"attempt_count integer NOT NULL DEFAULT 0",
			"CHECK (attempt_count >= 0)",
			"WHERE processed_at IS NULL AND dead_lettered_at IS NULL",
			"dead_lettered_at = NULL",
			"attempt_count = 0",
		} {
			if !strings.Contains(migration, contract) {
				t.Errorf("migration is missing durable outbox contract %q", contract)
			}
		}
		if strings.Contains(migration, "calendar_schedule_event_outbox_workspace_id_fkey") {
			t.Fatal("outbox rows must survive workspace deletion long enough to remove provider events")
		}
	})
}

func TestContinuousCalendarSchedulingRollbackRestoresConnectionOwnership(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(continuousCalendarSchedulingMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"DROP TRIGGER IF EXISTS workspace_members_reanchor_account_calendar_connection ON public.workspace_members",
		"DROP FUNCTION IF EXISTS public.reanchor_account_calendar_connection()",
		"DROP CONSTRAINT calendar_events_connection_scope_fkey",
		"WHERE cleanup_pending_at IS NOT NULL",
		"DROP COLUMN cleanup_pending_at",
		"ADD CONSTRAINT calendar_connections_workspace_id_fkey",
		"REFERENCES public.workspaces(workspace_id)",
		"ADD CONSTRAINT calendar_connections_workspace_member_fkey",
		"REFERENCES public.workspace_members(workspace_id, user_id)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("rollback is missing legacy connection ownership contract %q", contract)
		}
	}
	reanchorDrop := strings.Index(migration, "DROP FUNCTION IF EXISTS public.reanchor_account_calendar_connection()")
	outboxDrop := strings.Index(migration, "DROP TABLE IF EXISTS public.calendar_schedule_event_outbox")
	if reanchorDrop < 0 || outboxDrop < 0 || reanchorDrop >= outboxDrop {
		t.Fatal("rollback must remove the reanchor function before dropping its referenced outbox relation")
	}
}
