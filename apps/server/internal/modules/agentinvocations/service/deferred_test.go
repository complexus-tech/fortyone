package agentinvocations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type deferredLifecycleStore struct {
	lifecycleStore
	starts, defers int
}

func (s *deferredLifecycleStore) BeginAction(ctx context.Context, run Run, callback func(context.Context, Run) error) (Run, error) {
	s.starts++
	if err := callback(context.WithValue(ctx, transactionMarker{}, true), run); err != nil {
		return Run{}, err
	}
	now := time.Now()
	run.ActionStartedAt = &now
	s.run = run
	return run, nil
}
func (s *deferredLifecycleStore) DeferAction(_ context.Context, run Run) error {
	s.defers++
	run.Status = StatusQueued
	run.RetryAttemptCount = 0
	s.run = run
	return nil
}

type deferredLifecycleRuntime struct {
	lifecycleRuntime
	starts, polls int
	ready         bool
	pollError     error
}

func (r *deferredLifecycleRuntime) UsesDeferredAction(Run) bool { return true }
func (r *deferredLifecycleRuntime) BeginApply(ctx context.Context, _ Run, _ Proposal) error {
	if ctx.Value(transactionMarker{}) != true {
		return errors.New("staging escaped transaction")
	}
	r.starts++
	return nil
}
func (r *deferredLifecycleRuntime) PollApply(ctx context.Context, _ Run, _ Proposal) (Outcome, bool, error) {
	if ctx.Value(transactionMarker{}) != nil {
		return Outcome{}, false, errors.New("poll held transaction")
	}
	r.polls++
	return Outcome{Reply: "Updated.", Receipt: json.RawMessage(`{"revision":2}`)}, r.ready, r.pollError
}

func TestDeferredActionWaitsWithoutRepeatingModelOrPublishingSuccess(t *testing.T) {
	store := &deferredLifecycleStore{lifecycleStore: lifecycleStore{run: Run{ID: uuid.New()}}}
	runtime := &deferredLifecycleRuntime{lifecycleRuntime: lifecycleRuntime{proposal: Proposal{Action: json.RawMessage(`{"edit":true}`)}}}
	service := New(store, runtime)
	for range 5 {
		if err := service.Process(t.Context(), store.run.ID); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.prepareCalls != 1 || runtime.starts != 1 || runtime.applyCalls != 0 || runtime.publishCalls != 0 || store.defers != 5 || store.run.Status != StatusQueued {
		t.Fatalf("pending action repeated or published: runtime=%+v store=%+v", runtime, store)
	}
	runtime.ready = true
	if err := service.Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if store.run.Status != StatusSucceeded || store.run.Outcome == nil || runtime.applyCalls != 0 || runtime.publishCalls != 1 {
		t.Fatalf("receipt not completed: %+v", store.run)
	}
}

func TestDeferredActionRecoveryKeepsDurableStageAcrossTransientReceiptFailure(t *testing.T) {
	now := time.Now()
	store := &deferredLifecycleStore{lifecycleStore: lifecycleStore{run: Run{ID: uuid.New(), ActionStartedAt: &now, RetryAttemptCount: 10, Proposal: &Proposal{Action: json.RawMessage(`{"edit":true}`)}}}}
	runtime := &deferredLifecycleRuntime{pollError: errors.New("temporary database outage")}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); err == nil {
		t.Fatal("missing temporary failure")
	}
	if store.run.Status != StatusQueued || runtime.starts != 0 || runtime.prepareCalls != 0 || runtime.publishCalls != 0 {
		t.Fatalf("durable edit abandoned or replayed: %+v", store.run)
	}
}
