package migrations

import (
	"strings"
	"testing"
)

func TestCompletedCalendarScheduleBlocksMigrationPreservesHistory(t *testing.T) {
	t.Parallel()

	migration, err := FS.ReadFile("000145_completed_calendar_schedule_blocks.up.sql")
	if err != nil {
		t.Fatalf("read completed calendar schedule blocks migration: %v", err)
	}
	source := strings.ToLower(string(migration))
	for _, contract := range []string{
		"add column completed_at timestamptz",
		"and completed_at is null",
		"terminal_category = 'completed'",
		"block.start_at < effective_completed_at",
		"set end_at = least(end_at, effective_completed_at)",
		"completed_at = effective_completed_at",
		"external_provider = null",
		"calendar_schedule_event_outbox",
		"terminal_category = 'cancelled'",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("completed block migration is missing %q", contract)
		}
	}
}

func TestCompletedCalendarScheduleBlocksRollbackRestoresTerminalDeletion(t *testing.T) {
	t.Parallel()

	migration, err := FS.ReadFile("000145_completed_calendar_schedule_blocks.down.sql")
	if err != nil {
		t.Fatalf("read completed calendar schedule blocks rollback: %v", err)
	}
	source := strings.ToLower(string(migration))
	for _, contract := range []string{
		"delete from public.calendar_schedule_blocks",
		"where completed_at is not null",
		"status.category in ('completed', 'cancelled')",
		"drop column completed_at",
	} {
		if !strings.Contains(source, contract) {
			t.Fatalf("completed block rollback is missing %q", contract)
		}
	}
}
