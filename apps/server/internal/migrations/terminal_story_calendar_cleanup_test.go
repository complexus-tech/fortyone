package migrations

import (
	"strings"
	"testing"
)

const terminalStoryCalendarCleanupMigration = "000138_terminal_story_calendar_cleanup"

func TestTerminalStoryCalendarCleanupForwardMigration(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(terminalStoryCalendarCleanupMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"DELETE FROM public.calendar_schedule_blocks AS block",
		"block.block_type = 'work'",
		"story.completed_at IS NOT NULL",
		"status.category IN ('completed', 'cancelled')",
		"CREATE FUNCTION public.cleanup_terminal_story_calendar_schedule()",
		"DELETE FROM public.calendar_schedule_blocks",
		"block_type = 'work'",
		"CREATE TRIGGER stories_cleanup_terminal_calendar_schedule",
		"AFTER UPDATE OF status_id, completed_at ON public.stories",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("forward migration is missing terminal cleanup contract %q", contract)
		}
	}
	if strings.Contains(migration, "source = 'maya'") {
		t.Fatal("terminal cleanup must remove both manually scheduled and Maya-managed task blocks")
	}

	backfill := strings.Index(migration, "DELETE FROM public.calendar_schedule_blocks AS block")
	trigger := strings.Index(migration, "CREATE TRIGGER stories_cleanup_terminal_calendar_schedule")
	if backfill < 0 || trigger < 0 || backfill >= trigger {
		t.Fatal("existing terminal blocks must be retired before the lifecycle trigger is installed")
	}
}

func TestTerminalStoryCalendarCleanupRollback(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(terminalStoryCalendarCleanupMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"DROP TRIGGER IF EXISTS stories_cleanup_terminal_calendar_schedule ON public.stories",
		"DROP FUNCTION IF EXISTS public.cleanup_terminal_story_calendar_schedule()",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("rollback is missing terminal cleanup contract %q", contract)
		}
	}
}
