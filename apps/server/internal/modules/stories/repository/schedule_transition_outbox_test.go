package storiesrepository

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestImmediateScheduleTransitionRetryRequiresLatestFingerprintAndCurrentState(t *testing.T) {
	t.Parallel()
	reason := "Scheduled after the stand-up."
	current := dbScheduleTransitionStoryState{Status: "scheduled", Reason: &reason, Locked: false}
	latest := dbLatestScheduleTransition{SemanticFingerprint: "same", TransitionSequence: 7}

	require.True(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, false, "scheduled", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, false, "intervening"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "at_risk", &reason, false, "same"))
	require.False(t, isImmediateScheduleTransitionRetry(current, latest, true, "scheduled", &reason, true, "same"))
}

func TestStoryScheduleTransitionWriteSerializesAndUsesMonotonicLatestEvent(t *testing.T) {
	t.Parallel()

	require.Contains(t, lockStoryForScheduleTransitionQuery, "updated_at = $3")
	require.Contains(t, lockStoryForScheduleTransitionQuery, "FOR UPDATE")
	require.Contains(t, latestStoryScheduleTransitionFingerprintQuery, "transition_sequence")
	require.Contains(t, latestStoryScheduleTransitionFingerprintQuery, "ORDER BY transition_sequence DESC")
	require.NotContains(t, latestStoryScheduleTransitionFingerprintQuery, "ORDER BY created_at")
	require.Contains(t, insertStoryScheduleTransitionOutboxQuery, "transition_sequence")
	require.Contains(t, insertStoryScheduleTransitionOutboxQuery, "CAST($5 AS jsonb)")
	require.Contains(t, updateStoryAutoSchedulingStateForTransitionQuery, "updated_at = $3")
}

func TestStoryScheduleTransitionClaimsAreRecoverableAndClaimGuarded(t *testing.T) {
	t.Parallel()

	claim := strings.ToLower(claimStoryScheduleTransitionOutboxQuery)
	for _, contract := range []string{
		"for update skip locked",
		"status in ('pending', 'retrying')",
		"status = 'processing'",
		"claimed_at <= current_timestamp - cast($2 as interval)",
		"attempt_count = outbox.attempt_count + 1",
		"claim_token = gen_random_uuid()",
	} {
		require.Contains(t, claim, contract)
	}

	for _, lifecycleQuery := range []string{
		completeStoryScheduleTransitionOutboxQuery,
		retryStoryScheduleTransitionOutboxQuery,
		failStoryScheduleTransitionOutboxQuery,
	} {
		require.Contains(t, lifecycleQuery, "claim_token = $2")
		require.Contains(t, lifecycleQuery, "status = 'processing'")
	}
	require.Equal(t, "600.000000 seconds", scheduleTransitionIntervalLiteral(10*time.Minute))
}

func TestStoryScheduleTransitionPublisherFailuresRemainRetryableAndCleanupIsBounded(t *testing.T) {
	t.Parallel()

	require.Contains(t, retryStoryScheduleTransitionOutboxQuery, "status = 'retrying'")
	require.NotContains(t, retryStoryScheduleTransitionOutboxQuery, "status = 'failed'")
	require.Contains(t, deleteCompletedStoryScheduleTransitionOutboxQuery, "status = 'completed'")
	require.Contains(t, deleteCompletedStoryScheduleTransitionOutboxQuery, "FOR UPDATE SKIP LOCKED")
	require.Contains(t, deleteCompletedStoryScheduleTransitionOutboxQuery, "LIMIT $2")
	require.NotContains(t, deleteCompletedStoryScheduleTransitionOutboxQuery, "status = 'failed'")
}
