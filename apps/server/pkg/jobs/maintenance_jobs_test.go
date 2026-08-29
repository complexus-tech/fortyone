package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/stretchr/testify/require"
)

func TestPurgeExpiredTokensUsesOneClockForTokenAndFeedbackRetention(t *testing.T) {
	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	tokens := &verificationTokenPurgerStub{results: []int64{11}}
	feedbackStore := &tokenFeedbackMaintenanceStub{
		artifactResult: feedback.CoreContributorArtifactPurgeResult{
			VerificationsDeleted: 3,
			SessionsDeleted:      2,
		},
	}

	err := purgeExpiredTokensAt(context.Background(), tokens, feedbackStore, newTestJobLogger(), now)

	require.NoError(t, err)
	retainedBefore := now.UTC().Add(-7 * 24 * time.Hour)
	require.Equal(t, []time.Time{retainedBefore}, tokens.retainedBefore)
	require.Equal(t, []int{maintenancePurgeBatchSize}, tokens.batchSizes)
	require.Equal(t, 1, feedbackStore.artifactCalls)
	require.Equal(t, feedback.CoreContributorArtifactCutoffs{
		RetainedBefore: retainedBefore,
		ExpiredBefore:  now.UTC(),
	}, feedbackStore.cutoffs)
}

func TestPurgeExpiredTokensPreservesFeedbackCleanupErrors(t *testing.T) {
	feedbackErr := errors.New("feedback cleanup unavailable")
	tokens := &verificationTokenPurgerStub{results: []int64{0}}
	feedbackStore := &tokenFeedbackMaintenanceStub{artifactErr: feedbackErr}

	err := purgeExpiredTokensAt(
		context.Background(),
		tokens,
		feedbackStore,
		newTestJobLogger(),
		time.Date(2026, time.August, 28, 8, 15, 0, 0, time.UTC),
	)

	require.ErrorIs(t, err, feedbackErr)
	require.Equal(t, 1, feedbackStore.artifactCalls)
}

func TestProcessUserDeactivationUsesIndependentPolicyCutoffs(t *testing.T) {
	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	store := &inactiveUserDeactivatorStub{results: []int64{maintenancePurgeBatchSize, 4}}

	err := processUserDeactivationAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Len(t, store.calls, 2)
	expectedNow := now.UTC()
	for _, call := range store.calls {
		require.Equal(t, expectedNow.AddDate(0, -8, 0), call.inactiveBefore)
		require.Equal(t, expectedNow.Add(-30*24*time.Hour), call.warningSentBefore)
		require.Equal(t, expectedNow, call.deactivatedAt)
		require.Equal(t, maintenancePurgeBatchSize, call.batchSize)
	}
}

func TestPurgeOldStripeWebhookEventsUsesTerminalRetentionCutoff(t *testing.T) {
	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	store := &stripeWebhookEventPurgerStub{results: []int64{9}}

	err := purgeOldStripeWebhookEventsAt(context.Background(), store, newTestJobLogger(), now)

	require.NoError(t, err)
	require.Equal(t, []time.Time{now.UTC().Add(-30 * 24 * time.Hour)}, store.terminalBefore)
	require.Equal(t, []int{maintenancePurgeBatchSize}, store.batchSizes)
}

type verificationTokenPurgerStub struct {
	results        []int64
	retainedBefore []time.Time
	batchSizes     []int
}

func (store *verificationTokenPurgerStub) PurgeExpiredVerificationTokens(
	_ context.Context,
	retainedBefore time.Time,
	batchSize int,
) (int64, error) {
	store.retainedBefore = append(store.retainedBefore, retainedBefore)
	store.batchSizes = append(store.batchSizes, batchSize)
	if len(store.results) == 0 {
		return 0, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}

type tokenFeedbackMaintenanceStub struct {
	cutoffs        feedback.CoreContributorArtifactCutoffs
	artifactResult feedback.CoreContributorArtifactPurgeResult
	artifactErr    error
	artifactCalls  int
}

func (store *tokenFeedbackMaintenanceStub) PurgeExpiredContributorArtifacts(
	_ context.Context,
	cutoffs feedback.CoreContributorArtifactCutoffs,
) (feedback.CoreContributorArtifactPurgeResult, error) {
	store.artifactCalls++
	store.cutoffs = cutoffs
	return store.artifactResult, store.artifactErr
}

func (*tokenFeedbackMaintenanceStub) PurgeDeletedFeedback(
	context.Context,
	time.Time,
) (feedback.CoreDeletedFeedbackPurgeResult, error) {
	return feedback.CoreDeletedFeedbackPurgeResult{}, nil
}

type inactiveUserDeactivationCall struct {
	inactiveBefore    time.Time
	warningSentBefore time.Time
	deactivatedAt     time.Time
	batchSize         int
}

type inactiveUserDeactivatorStub struct {
	results []int64
	calls   []inactiveUserDeactivationCall
}

func (store *inactiveUserDeactivatorStub) DeactivateInactiveUsers(
	_ context.Context,
	inactiveBefore time.Time,
	warningSentBefore time.Time,
	deactivatedAt time.Time,
	batchSize int,
) (int64, error) {
	store.calls = append(store.calls, inactiveUserDeactivationCall{
		inactiveBefore:    inactiveBefore,
		warningSentBefore: warningSentBefore,
		deactivatedAt:     deactivatedAt,
		batchSize:         batchSize,
	})
	if len(store.results) == 0 {
		return 0, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}

type stripeWebhookEventPurgerStub struct {
	results        []int64
	terminalBefore []time.Time
	batchSizes     []int
}

func (store *stripeWebhookEventPurgerStub) PurgeTerminalStripeWebhookEvents(
	_ context.Context,
	terminalBefore time.Time,
	batchSize int,
) (int64, error) {
	store.terminalBefore = append(store.terminalBefore, terminalBefore)
	store.batchSizes = append(store.batchSizes, batchSize)
	if len(store.results) == 0 {
		return 0, nil
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}
