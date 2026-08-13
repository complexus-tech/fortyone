package feedbackrepository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicationWriteQueriesKeepStateAndOutboxAtomic(t *testing.T) {
	t.Parallel()

	require.Contains(t, lockUpdateForPublicationQuery, "FOR UPDATE OF fu")
	require.Contains(t, lockUpdateForPublicationQuery, "feedback_portals portal")
	require.Contains(t, lockPublicationItemsQuery, "FOR SHARE OF item")
	require.Contains(t, lockPublicationItemsQuery, "feedback_update_items")
	require.Contains(t, mutateUpdateForPublicationQuery, "status = 'published'")
	require.Contains(t, mutateUpdateForPublicationQuery, "publication_sequence = publication_sequence + 1")
	require.Contains(t, mutateUpdateForPublicationQuery, "status = 'draft'")
	require.Contains(t, insertPublicationOutboxQuery, "feedback_update_publication_outbox")
	require.Contains(t, insertPublicationOutboxQuery, "publication_sequence")
	require.Contains(t, insertPublicationOutboxQuery, "event_payload")
}

func TestPublicationClaimQueryRecoversStaleClaimsWithoutDoubleClaiming(t *testing.T) {
	t.Parallel()

	require.Contains(t, claimPublicationOutboxQuery, "status IN ('pending', 'retrying')")
	require.Contains(t, claimPublicationOutboxQuery, "status = 'processing'")
	require.Contains(t, claimPublicationOutboxQuery, "claimed_at <= NOW() - CAST($2 AS interval)")
	require.Contains(t, claimPublicationOutboxQuery, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, claimPublicationOutboxQuery, "claim_token = gen_random_uuid()")
	require.Contains(t, claimPublicationOutboxQuery, "attempt_count = attempt_count + 1")
}

func TestPublicationLifecycleWritesAreClaimTokenGuarded(t *testing.T) {
	t.Parallel()

	for _, query := range []string{completePublicationOutboxQuery, retryPublicationOutboxQuery} {
		require.Contains(t, query, "publication_event_id = $1")
		require.Contains(t, query, "claim_token = $2")
		require.Contains(t, query, "status = 'processing'")
	}
	require.Contains(t, completePublicationOutboxQuery, "status = 'completed'")
	require.Contains(t, retryPublicationOutboxQuery, "last_error = LEFT($5, 4000)")
}

func TestPublicationAudienceIsSnapshottedAndOnlyCurrentEligibleRecipientsDispatch(t *testing.T) {
	t.Parallel()

	for _, query := range []string{
		snapshotPublicationContributorAudienceQuery,
		snapshotPublicationAccountAudienceQuery,
	} {
		require.Contains(t, query, "feedback_item_followers")
		require.Contains(t, query, "feedback_portal_followers")
		require.Contains(t, query, "unsubscribed_at IS NULL")
		require.Contains(t, query, "contributor.blocked_at IS NULL")
	}
	require.Contains(t, snapshotPublicationContributorAudienceQuery, "preference.email_unsubscribed_at IS NULL")
	require.Contains(t, snapshotPublicationAccountAudienceQuery, "account.is_active = true")

	for _, query := range []string{
		listPublicationDeliveryRecipientsQuery,
		listAccountPublicationRecipientsQuery,
	} {
		require.Contains(t, query, "unnest(CAST($2 AS uuid[]))")
		require.Contains(t, query, "feedback_item_followers")
		require.Contains(t, query, "feedback_portal_followers")
		require.Contains(t, query, "unsubscribed_at IS NULL")
		require.Contains(t, query, "contributor.blocked_at IS NULL")
		require.Contains(t, query, "COALESCE(published_item.merged_into_item_id, published_item.id)")
	}
	require.Contains(t, listPublicationDeliveryRecipientsQuery, "preference.email_unsubscribed_at IS NULL")
	require.Contains(t, listAccountPublicationRecipientsQuery, "account.is_active = true")
}
