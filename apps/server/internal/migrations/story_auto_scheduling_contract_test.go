package migrations

import (
	"strings"
	"testing"
)

const storyAutoSchedulingContractMigration = "000132_story_auto_scheduling_contract"

func TestStoryAutoSchedulingContractForwardMigration(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyAutoSchedulingContractMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	t.Run("story contract", func(t *testing.T) {
		for _, contract := range []string{
			"ADD COLUMN auto_scheduling_enabled boolean NOT NULL DEFAULT false",
			"ADD COLUMN auto_scheduling_locked boolean NOT NULL DEFAULT false",
			"ADD COLUMN auto_scheduling_status text NOT NULL DEFAULT 'off'",
			"ADD COLUMN auto_scheduling_reason text",
			"ADD COLUMN auto_scheduling_updated_at timestamptz",
			"CHECK (NOT auto_scheduling_locked OR auto_scheduling_enabled)",
			"'off'",
			"'needs_owner'",
			"'needs_time'",
			"'planning'",
			"'scheduled'",
			"'at_risk'",
			"'cannot_fit'",
			"'locked'",
		} {
			if !strings.Contains(migration, contract) {
				t.Errorf("migration is missing story auto-scheduling contract %q", contract)
			}
		}
		for _, nonNullableColumn := range []string{
			"ADD COLUMN auto_scheduling_reason text NOT NULL",
			"ADD COLUMN auto_scheduling_updated_at timestamptz NOT NULL",
		} {
			if strings.Contains(migration, nonNullableColumn) {
				t.Errorf("story auto-scheduling metadata must remain nullable: found %q", nonNullableColumn)
			}
		}
	})

	t.Run("legacy enrollment backfill", func(t *testing.T) {
		for _, contract := range []string{
			"FROM public.calendar_maya_schedule_ownerships AS ownership",
			"FROM public.calendar_schedule_blocks AS block",
			"block.source = 'maya'",
			"actor.email = 'maya@fortyone.app'",
			"actor.is_system = TRUE",
			"story.deleted_at IS NULL",
			"story.archived_at IS NULL",
			"story.completed_at IS NULL",
			"story.is_draft = FALSE",
			"status.category IN ('completed', 'cancelled')",
			"WHEN managed.has_maya_blocks THEN 'scheduled'",
			"WHEN managed.assignee_id IS NULL OR managed.assigned_to_maya THEN 'needs_owner'",
			"WHEN managed.estimated_duration_minutes IS NULL THEN 'needs_time'",
			"ELSE 'planning'",
			"Choose an owner before Maya can schedule this story.",
			"Maya is selecting an eligible owner for this story.",
			"Add a time estimate before Maya can schedule this story.",
			"Maya is checking availability and scheduling this story.",
			"SET auto_scheduling_enabled = TRUE",
			"auto_scheduling_reason = classified.scheduling_reason",
			"auto_scheduling_updated_at = CURRENT_TIMESTAMP",
		} {
			if !strings.Contains(migration, contract) {
				t.Errorf("migration is missing legacy enrollment backfill contract %q", contract)
			}
		}

		blocksStatus := strings.Index(migration, "WHEN managed.has_maya_blocks THEN 'scheduled'")
		ownerStatus := strings.Index(migration, "WHEN managed.assignee_id IS NULL OR managed.assigned_to_maya THEN 'needs_owner'")
		timeStatus := strings.Index(migration, "WHEN managed.estimated_duration_minutes IS NULL THEN 'needs_time'")
		if blocksStatus < 0 || ownerStatus < 0 || timeStatus < 0 || blocksStatus >= ownerStatus || ownerStatus >= timeStatus {
			t.Fatal("backfill must prefer persisted blocks, then owner readiness, then time readiness")
		}
	})
}

func TestStoryAutoSchedulingContractRollback(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyAutoSchedulingContractMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"DROP CONSTRAINT IF EXISTS stories_auto_scheduling_status_check",
		"DROP CONSTRAINT IF EXISTS stories_auto_scheduling_locked_requires_enabled_check",
		"DROP COLUMN IF EXISTS auto_scheduling_updated_at",
		"DROP COLUMN IF EXISTS auto_scheduling_reason",
		"DROP COLUMN IF EXISTS auto_scheduling_status",
		"DROP COLUMN IF EXISTS auto_scheduling_locked",
		"DROP COLUMN IF EXISTS auto_scheduling_enabled",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("rollback is missing story auto-scheduling contract %q", contract)
		}
	}
}
