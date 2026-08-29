package feedbackrepository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFeedbackUpdateQueries(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile("queries/updates.sql")
	require.NoError(t, err)
	return string(contents)
}

func TestPublicationWriteQueriesKeepStateAndOutboxAtomic(t *testing.T) {
	t.Parallel()
	queries := readFeedbackUpdateQueries(t)

	require.Contains(t, queries, "FOR UPDATE OF update_record")
	require.Contains(t, queries, "FOR SHARE OF item")
	require.Contains(t, queries, "publication_sequence = publication_sequence + 1")
	require.Contains(t, queries, "update_record.status = 'draft'")
	require.Contains(t, queries, "INSERT INTO feedback_update_publication_outbox")
	require.Contains(t, queries, "event_payload")
}

func TestPublicationClaimQueryRecoversStaleClaimsWithoutDoubleClaiming(t *testing.T) {
	t.Parallel()
	queries := readFeedbackUpdateQueries(t)

	require.Contains(t, queries, "status IN ('pending', 'retrying')")
	require.Contains(t, queries, "status = 'processing'")
	require.Contains(t, queries, "claimed_at <= sqlc.arg(stale_before)")
	require.Contains(t, queries, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, queries, "claim_token = gen_random_uuid()")
	require.Contains(t, queries, "attempt_count = attempt_count + 1")
}

func TestPublicationLifecycleWritesAreClaimTokenGuarded(t *testing.T) {
	t.Parallel()
	queries := readFeedbackUpdateQueries(t)

	require.Contains(t, queries, "claim_token = sqlc.arg(claim_token)")
	require.Contains(t, queries, "status = 'completed'")
	require.Contains(t, queries, "last_error = LEFT(sqlc.arg(failure), 4000)")
}

func TestPublicationAudienceSnapshotRetainsCurrentEligibilityGuards(t *testing.T) {
	t.Parallel()
	queries := readFeedbackUpdateQueries(t)

	require.Contains(t, queries, "feedback_item_followers")
	require.Contains(t, queries, "feedback_portal_followers")
	require.Contains(t, queries, "unsubscribed_at IS NULL")
	require.Contains(t, queries, "contributor.blocked_at IS NULL")
	require.Contains(t, queries, "preference.email_unsubscribed_at IS NULL")
	require.Contains(t, queries, "account.is_active = true")
	require.Contains(t, queries, "COALESCE(published_item.merged_into_item_id, published_item.id)")
}
