package migrations

import (
	"strings"
	"testing"
)

const storyScheduleTransitionOutboxMigration = "000133_story_schedule_transition_outbox"

func TestStoryScheduleTransitionOutboxMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyScheduleTransitionOutboxMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	assertStoryScheduleOutboxMigrationContains(t, migration,
		"CREATE TABLE public.story_schedule_transition_outbox",
		"schedule_transition_event_id uuid NOT NULL DEFAULT gen_random_uuid()",
		"actor_id uuid NOT NULL",
		"story_id uuid NOT NULL",
		"workspace_id uuid NOT NULL",
		"event_type text NOT NULL DEFAULT 'story.updated'",
		"event_payload jsonb NOT NULL",
		"event_payload ? 'type'",
		"event_payload ? 'payload'",
		"event_payload ? 'timestamp'",
		"event_payload ? 'actor_id'",
		"event_payload ->> 'type' = event_type",
		"event_payload #>> '{payload,story_id}' = CAST(story_id AS text)",
		"event_payload #>> '{payload,workspace_id}' = CAST(workspace_id AS text)",
		"semantic_fingerprint text NOT NULL",
		"transition_sequence bigint NOT NULL",
		"story_schedule_transition_outbox_sequence_check",
		"story_schedule_transition_outbox_story_sequence_key",
		"UNIQUE (workspace_id, story_id, transition_sequence)",
		"status IN ('pending', 'processing', 'retrying', 'completed', 'failed')",
		"story_schedule_transition_outbox_lifecycle_check",
		"status = 'pending'",
		"status = 'processing'",
		"status = 'retrying'",
		"status = 'completed'",
		"status = 'failed'",
		"next_attempt_at timestamptz DEFAULT now()",
		"claim_token uuid",
		"claimed_at timestamptz",
		"completed_at timestamptz",
		"last_error text",
		"created_at timestamptz NOT NULL DEFAULT now()",
		"updated_at timestamptz NOT NULL DEFAULT now()",
		"story_schedule_transition_outbox_claim_token_key",
		"idx_story_schedule_transition_outbox_ready",
		"idx_story_schedule_transition_outbox_stale_claim",
		"idx_story_schedule_transition_outbox_story_latest",
		"transition_sequence DESC",
		"idx_story_schedule_transition_outbox_semantic_fingerprint",
		"idx_story_schedule_transition_outbox_retention",
		"WHERE status = 'completed'",
		"intentionally has no",
		"complete\n-- events.Event JSON snapshot",
		"deliberately non-unique",
		"bounded DELETE batches",
	)

	if strings.Contains(migration, "FOREIGN KEY") || strings.Contains(migration, "REFERENCES public.") {
		t.Fatal("schedule transition outbox must survive deletion of its mutable source rows")
	}
	if strings.Contains(migration, "CREATE UNIQUE INDEX idx_story_schedule_transition_outbox_semantic_fingerprint") {
		t.Fatal("semantic fingerprint must remain non-unique so a transition may recur after intervening states")
	}
	if strings.Contains(migration, "WHERE status IN ('completed', 'failed')") {
		t.Fatal("automatic retention must preserve permanently malformed rows for operator diagnosis")
	}
}

func TestStoryScheduleTransitionOutboxMigrationIsForwardOnly(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyScheduleTransitionOutboxMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	assertStoryScheduleOutboxMigrationContains(t, migration,
		"migration 000133 is forward-only",
		"schedule transition event snapshots and delivery state are immutable",
	)
	if strings.Contains(migration, "DROP TABLE") || strings.Contains(migration, "DROP INDEX") {
		t.Fatal("forward-only migration must not destructively roll back schedule transition state")
	}
}

func assertStoryScheduleOutboxMigrationContains(t *testing.T, migration string, contracts ...string) {
	t.Helper()

	for _, contract := range contracts {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing schedule transition outbox contract %q", contract)
		}
	}
}
