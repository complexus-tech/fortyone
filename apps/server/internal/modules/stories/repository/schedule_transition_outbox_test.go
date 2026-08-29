package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImmediateScheduleTransitionRetryRequiresLatestFingerprintAndCurrentState(t *testing.T) {
	t.Parallel()
	reason := "Scheduled after the stand-up."
	current := scheduleTransitionStoryState{Status: "scheduled", Reason: &reason, Locked: false}
	latest := latestScheduleTransition{SemanticFingerprint: "same", TransitionSequence: 7}

	require.True(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, false, "scheduled", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, false, "intervening"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "at_risk", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, true, "same"))
}

func TestStoryScheduleTransitionWriteSerializesAndUsesMonotonicLatestEvent(t *testing.T) {
	t.Parallel()
	queries := scheduleTransitionQueries(t)

	require.Contains(t, queries, "story.updated_at = sqlc.arg(expected_updated_at)")
	require.Contains(t, queries, "FOR UPDATE")
	require.Contains(t, queries, "ORDER BY outbox.transition_sequence DESC")
	require.NotContains(t, queries, "ORDER BY outbox.created_at DESC")
	require.Contains(t, queries, "transition_sequence")
	require.Contains(t, queries, "sqlc.arg(event_payload)")
	require.Contains(t, queries, "sqlc.narg(claim_token)")
}

func TestStoryScheduleTransitionClaimsAreRecoverableAndClaimGuarded(t *testing.T) {
	t.Parallel()

	claim := strings.ToLower(scheduleTransitionQueries(t))
	for _, contract := range []string{
		"for update skip locked",
		"status in ('pending', 'retrying')",
		"status = 'processing'",
		"claimed_at <= current_timestamp - make_interval",
		"attempt_count = outbox.attempt_count + 1",
		"claim_token = gen_random_uuid()",
	} {
		require.Contains(t, claim, contract)
	}

	require.GreaterOrEqual(t, strings.Count(claim, "claim_token = sqlc.arg(claim_token)"), 3)
	require.GreaterOrEqual(t, strings.Count(claim, "status = 'processing'"), 4)
}

func TestStoryScheduleTransitionPublisherFailuresRemainRetryableAndCleanupIsBounded(t *testing.T) {
	t.Parallel()

	queries := scheduleTransitionQueries(t)
	require.Contains(t, queries, "status = 'retrying'")
	require.Contains(t, queries, "status = 'completed'")
	require.Contains(t, queries, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, queries, "LIMIT sqlc.arg(batch_size)")
}

func scheduleTransitionQueries(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("queries/schedule_transition.sql")
	require.NoError(t, err)
	return string(data)
}
