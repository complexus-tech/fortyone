package agentinvocations

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDurableRepliesRecoverBeyondPreparationRetryLimit(t *testing.T) {
	for _, approval := range []bool{false, true} {
		t.Run(map[bool]string{false: "result", true: "approval"}[approval], func(t *testing.T) {
			outage := errors.New("delivery is temporarily unavailable")
			store := &lifecycleStore{run: Run{ID: uuid.New()}}
			runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{
				proposal:      Proposal{Reply: "Should I apply this?", Review: "Status: Open → Closed", Action: json.RawMessage(`{"token":"opaque"}`), RequiresApproval: approval},
				deliveryError: outage,
			}}
			service := New(store, runtime)
			for range 5 {
				require.ErrorIs(t, service.Process(t.Context(), store.run.ID), outage)
				require.Equal(t, StatusQueued, store.run.Status, "a durable reply must remain recoverable after the model attempt limit")
				require.False(t, store.run.FailurePending)
				require.Zero(t, runtime.failurePublications)
			}
			runtime.deliveryError = nil
			require.NoError(t, service.Process(t.Context(), store.run.ID))
			require.Equal(t, 1, runtime.prepareCalls)
			require.Equal(t, 6, runtime.deliveryCalls)
			if approval {
				require.Equal(t, StatusAwaitingApproval, store.run.Status)
				require.Equal(t, 1, runtime.proposalPublishCalls)
				require.Zero(t, runtime.applyCalls)
				require.Zero(t, runtime.publishCalls)
			} else {
				require.Equal(t, StatusSucceeded, store.run.Status)
				require.Equal(t, 1, runtime.applyCalls)
				require.Equal(t, 1, runtime.publishCalls)
			}
		})
	}
}

func TestDurableReplyRecoveryStillStopsOnPermissionOrSourceRevocation(t *testing.T) {
	for _, denial := range []error{ErrPermissionDenied, ErrSourceChanged} {
		store := &lifecycleStore{run: Run{ID: uuid.New(), RetryAttemptCount: 5,
			Proposal: &Proposal{Reply: "Saved response"}, Outcome: &Outcome{Reply: "Saved response"}}}
		runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{authorizationError: denial}}
		require.ErrorIs(t, New(store, runtime).Process(t.Context(), store.run.ID), denial)
		require.Equal(t, StatusFailed, store.run.Status)
		require.Zero(t, runtime.prepareCalls)
		require.Zero(t, runtime.publishCalls)
		require.Zero(t, runtime.deliveryCalls)
		require.Zero(t, runtime.failurePublications)
	}
}

func TestCommittedOutcomePublicationRetriesWithoutRepeatingAction(t *testing.T) {
	outage := errors.New("comment storage temporarily unavailable")
	store := &lifecycleStore{run: Run{ID: uuid.New(), RetryAttemptCount: 5,
		Proposal: &Proposal{Action: json.RawMessage(`{}`)}, Outcome: &Outcome{Reply: "Updated", Receipt: json.RawMessage(`{"applied":true}`)}}}
	runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{publishError: outage}}
	service := New(store, runtime)
	require.ErrorIs(t, service.Process(t.Context(), store.run.ID), outage)
	require.Equal(t, StatusQueued, store.run.Status)
	require.Nil(t, store.run.ReplyCommentID)
	runtime.publishError = nil
	require.NoError(t, service.Process(t.Context(), store.run.ID))
	require.Equal(t, StatusSucceeded, store.run.Status)
	require.Zero(t, runtime.prepareCalls)
	require.Zero(t, runtime.applyCalls)
	require.Equal(t, 2, runtime.publishCalls)
	require.Equal(t, 1, runtime.deliveryCalls)
}
