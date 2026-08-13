package migrations

import (
	"strings"
	"testing"
)

const feedbackPublicationOutboxMigration = "000123_feedback_update_publication_outbox"

func TestFeedbackPublicationOutboxMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackPublicationOutboxMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	assertPublicationMigrationContains(t, migration,
		"ADD COLUMN publication_sequence bigint NOT NULL DEFAULT 0",
		"SET publication_sequence = 1",
		"status <> 'published' OR publication_sequence > 0",
		"CREATE TABLE public.feedback_update_publication_outbox",
		"publication_event_id uuid NOT NULL DEFAULT gen_random_uuid()",
		"event_payload jsonb NOT NULL",
		"feedback_update_publication_outbox_update_sequence_key",
		"UNIQUE (update_id, publication_sequence)",
		"status IN ('pending', 'processing', 'retrying', 'completed', 'failed')",
		"feedback_update_publication_outbox_lifecycle_check",
		"feedback_update_publication_outbox_claim_token_key",
		"idx_feedback_update_publication_outbox_ready",
		"idx_feedback_update_publication_outbox_stale_claim",
		"idx_feedback_update_publication_outbox_retention",
		"intentionally has no foreign keys",
	)

	if strings.Contains(migration, "REFERENCES public.feedback_updates") {
		t.Fatal("publication outbox must survive deletion of its source Update")
	}
	if strings.Contains(migration, "feedback_contributor_merges") {
		t.Fatal("identity merge policy is intentionally deferred from this migration")
	}
}

func TestFeedbackPublicationOutboxMigrationIsForwardOnly(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackPublicationOutboxMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	assertPublicationMigrationContains(t, migration,
		"migration 000123 is forward-only",
		"publication event identities and delivery state are immutable",
	)
	if strings.Contains(migration, "DROP TABLE") || strings.Contains(migration, "DROP COLUMN") {
		t.Fatal("forward-only migration must not destructively roll back publication state")
	}
}

func assertPublicationMigrationContains(t *testing.T, migration string, contracts ...string) {
	t.Helper()

	for _, contract := range contracts {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing contract %q", contract)
		}
	}
}
