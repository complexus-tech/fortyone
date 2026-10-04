//go:build integration

package useruow

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccountDeletionErasesIdentifiedEventReceiptsBeforeDeferredHardDelete(t *testing.T) {
	f := newDeletionFixture(t)
	f.cleanup.pending = true
	ctx := t.Context()

	for _, event := range []struct {
		key              string
		recipient, actor uuid.UUID
		assignee         *uuid.UUID
	}{
		{"personal-receipt", f.user, f.other, nil},
		{"departing-actor", f.other, f.user, nil},
		{"departing-assignee", f.other, f.other, &f.user},
		{"same-named-member", f.other, f.other, &f.other},
		{"ambiguous-legacy-name", f.other, f.other, nil},
	} {
		insertDeletionReceipt(t, f, event.key, event.recipient, event.actor, event.assignee)
	}

	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	// The account still exists while remote cleanup is pending, so FK cascades
	// cannot be responsible for removing these immutable original snapshots.
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND NOT is_active`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM notification_event_receipts WHERE dedupe_key IN ('personal-receipt','departing-actor','departing-assignee')`, 0)
	deletionCount(t, f, `SELECT count(*) FROM notification_event_receipts WHERE dedupe_key IN ('same-named-member','ambiguous-legacy-name')`, 2)
}

func TestAccountDeletionReceiptCleanupRollsBackWithAccountErasure(t *testing.T) {
	f := newDeletionFixture(t)
	insertDeletionReceipt(t, f, "departing-assignee", f.other, f.other, &f.user)
	execDeletion(t, t.Context(), f.db.Pool, `CREATE FUNCTION reject_receipt_erasure_test() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced erasure failure'; END $$; CREATE TRIGGER reject_receipt_erasure_test BEFORE DELETE ON users FOR EACH ROW EXECUTE FUNCTION reject_receipt_erasure_test()`)

	_, err := f.manager.Delete(t.Context(), f.command())
	require.ErrorContains(t, err, "forced erasure failure")
	deletionCount(t, f, `SELECT count(*) FROM notification_event_receipts WHERE dedupe_key='departing-assignee'`, 1)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND is_active`, 1, f.user)
}

func insertDeletionReceipt(t *testing.T, f deletionFixture, key string, recipient, actor uuid.UUID, assignee *uuid.UUID) {
	t.Helper()
	execDeletion(t, t.Context(), f.db.Pool, `
		INSERT INTO notification_event_receipts (
			dedupe_key, recipient_id, workspace_id, actor_id, entity_id,
			notification_type, entity_type, coalescing_key, occurred_at, payload
		)
		VALUES ($1,$2,$3,$4,$5,'story_update','story','story_update',now(),
			jsonb_build_array('story_update','story',CAST($5 AS uuid),CAST($4 AS uuid),'Task assignment',
				CAST('{"variables":{"assignee":{"value":"departing","type":"assignee"}}}' AS jsonb)
				|| CASE WHEN CAST($6 AS uuid) IS NULL THEN CAST('{}' AS jsonb)
				   ELSE jsonb_build_object('identityReferences',jsonb_build_object('assignee',CAST($6 AS uuid))) END))`,
		key, recipient, f.workspace, actor, f.story, assignee)
}
