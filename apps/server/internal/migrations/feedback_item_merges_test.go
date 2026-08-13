package migrations

import (
	"strings"
	"testing"
)

const feedbackItemMergesMigration = "000124_feedback_item_merges"

func TestFeedbackItemMergesMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackItemMergesMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	assertItemMergeMigrationContains(t, migration,
		"ADD COLUMN merged_into_item_id uuid",
		"ADD COLUMN merged_at timestamptz",
		"ADD COLUMN merged_by_user_id uuid",
		"UNIQUE (workspace_id, portal_id, id)",
		"FOREIGN KEY (workspace_id, portal_id, merged_into_item_id)",
		"REFERENCES public.feedback_items(workspace_id, portal_id, id)",
		"ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED",
		"merged_into_item_id <> id",
		"feedback_items_merge_lifecycle_check",
		"idx_feedback_items_merged_into_item",
		"WHERE deleted_at IS NULL AND merged_into_item_id IS NULL",
		"CREATE TABLE public.feedback_item_merge_outbox",
		"merge_event_id uuid NOT NULL DEFAULT gen_random_uuid()",
		"event_type = 'feedback.item.merged'",
		"feedback_item_merge_outbox_source_item_key UNIQUE (source_item_id)",
		"feedback_item_merge_outbox_lifecycle_check",
		"feedback_item_merge_outbox_claim_token_key",
		"idx_feedback_item_merge_outbox_ready",
		"idx_feedback_item_merge_outbox_stale_claim",
		"idx_feedback_item_merge_outbox_retention",
		"intentionally has no foreign",
	)

	if strings.Contains(migration, "REFERENCES public.feedback_item_merge_outbox") {
		t.Fatal("feedback item merge state must not depend on outbox retention")
	}
}

func TestFeedbackItemMergesRollbackProtectsHistory(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackItemMergesMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	assertItemMergeMigrationContains(t, migration,
		"cannot be rolled back while merged feedback items exist",
		"cannot be rolled back while feedback item merge events exist",
		"DROP TABLE IF EXISTS public.feedback_item_merge_outbox",
		"DROP COLUMN IF EXISTS merged_into_item_id",
	)
}

func assertItemMergeMigrationContains(t *testing.T, migration string, contracts ...string) {
	t.Helper()

	for _, contract := range contracts {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing contract %q", contract)
		}
	}
}
