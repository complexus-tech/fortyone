package agentinvocations

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type failureRuntime struct {
	*lifecycleRuntime
	failurePublications int
}

func (r *failureRuntime) PublishFailure(ctx context.Context, _ Run, summary string) (uuid.UUID, error) {
	r.failurePublications++
	if ctx.Value(transactionMarker{}) != true {
		return uuid.Nil, errors.New("failure escaped transaction")
	}
	if summary == "" {
		return uuid.Nil, ErrInvalidInput
	}
	return uuid.New(), nil
}

func (s *lifecycleStore) BeginFailure(_ context.Context, run Run, summary string) (Run, error) {
	run.FailurePending = true
	run.Summary = summary
	s.run = run
	return run, nil
}
func (s *lifecycleStore) PublishFailure(ctx context.Context, run Run, callback func(context.Context, Run) (uuid.UUID, error)) (Run, error) {
	id, err := callback(context.WithValue(ctx, transactionMarker{}, true), run)
	if err != nil {
		return Run{}, err
	}
	run.FailureReplyID = &id
	s.run = run
	return run, nil
}
func (s *lifecycleStore) DeferFailure(_ context.Context, run Run) error {
	run.Status = StatusQueued
	s.run = run
	return nil
}
func (s *lifecycleStore) CompleteFailure(_ context.Context, run Run) (Run, error) {
	run.Status = StatusFailed
	run.FailurePending = false
	s.run = run
	return run, nil
}

func TestFailurePublicationAndDeliveryRecoverWithoutModelReplay(t *testing.T) {
	store := &lifecycleStore{run: Run{ID: uuid.New(), RetryAttemptCount: 3}}
	runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{prepareError: errors.New("raw provider secret"), deliveryError: errors.New("temporary delivery outage")}}
	service := New(store, runtime)
	if err := service.Process(t.Context(), store.run.ID); err == nil {
		t.Fatal("expected provider failure")
	}
	if !store.run.FailurePending || store.run.FailureReplyID == nil || store.run.Status != StatusQueued || store.run.Outcome != nil {
		t.Fatalf("failure was not durable: %+v", store.run)
	}
	if strings.Contains(store.run.Summary, "secret") {
		t.Fatal("raw provider failure leaked")
	}
	firstID := *store.run.FailureReplyID
	runtime.deliveryError = nil
	if err := service.Process(t.Context(), store.run.ID); err != nil {
		t.Fatal(err)
	}
	if store.run.Status != StatusFailed || store.run.FailurePending || runtime.prepareCalls != 1 || runtime.applyCalls != 0 || runtime.failurePublications != 1 || *store.run.FailureReplyID != firstID {
		t.Fatalf("failure recovery reran work: run=%+v runtime=%+v", store.run, runtime)
	}
}

func TestFailurePublicationDoesNotBypassRevokedAccess(t *testing.T) {
	for _, failure := range []error{ErrPermissionDenied, ErrSourceChanged} {
		store := &lifecycleStore{run: Run{ID: uuid.New()}}
		runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{prepareError: failure}}
		if err := New(store, runtime).Process(t.Context(), store.run.ID); !errors.Is(err, failure) {
			t.Fatal(err)
		}
		if runtime.failurePublications != 0 || store.run.FailurePending || store.run.Status != StatusFailed {
			t.Fatalf("unsafe failure publication: %+v", store.run)
		}
	}
	store := &lifecycleStore{run: Run{ID: uuid.New(), FailurePending: true, Summary: "Safe failure"}}
	runtime := &failureRuntime{lifecycleRuntime: &lifecycleRuntime{denyTransaction: true}}
	if err := New(store, runtime).Process(t.Context(), store.run.ID); !errors.Is(err, ErrPermissionDenied) {
		t.Fatal(err)
	}
	if runtime.failurePublications != 0 || store.run.FailurePending || store.run.Status != StatusFailed {
		t.Fatalf("revoked recovery published: %+v", store.run)
	}
}
