package workerbootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

type outboundWebhookDispatcherStub struct {
	work        []bool
	dispatchErr error
	recovered   int64
	recoverErr  error
	dispatches  int
}

type storyMutationEventDispatcherStub struct {
	processed int
	err       error
	calls     int
}

func (dispatcher *storyMutationEventDispatcherStub) DispatchBatch(context.Context) (int, error) {
	dispatcher.calls++
	return dispatcher.processed, dispatcher.err
}

func (dispatcher *outboundWebhookDispatcherStub) DispatchOne(context.Context) (bool, error) {
	dispatcher.dispatches++
	if dispatcher.dispatchErr != nil {
		return false, dispatcher.dispatchErr
	}
	if len(dispatcher.work) == 0 {
		return false, nil
	}
	worked := dispatcher.work[0]
	dispatcher.work = dispatcher.work[1:]
	return worked, nil
}

func (dispatcher *outboundWebhookDispatcherStub) RecoverExpiredLeases(context.Context) (int64, error) {
	return dispatcher.recovered, dispatcher.recoverErr
}

func TestOutboundWebhookTaskRecoversAndDrainsBoundedBatch(t *testing.T) {
	t.Parallel()
	dispatcher := &outboundWebhookDispatcherStub{recovered: 2, work: []bool{true, true, false}}
	storyEvents := &storyMutationEventDispatcherStub{processed: 3}
	if err := outboundWebhookTaskHandler(nil, storyEvents, dispatcher)(t.Context(), asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil)); err != nil {
		t.Fatalf("outbound webhook task error = %v", err)
	}
	if storyEvents.calls != 1 || dispatcher.dispatches != 3 {
		t.Fatalf("story event calls=%d dispatch calls=%d, want 1 and 3", storyEvents.calls, dispatcher.dispatches)
	}

	dispatcher = &outboundWebhookDispatcherStub{work: make([]bool, outboundWebhookDispatchBatchSize+1)}
	for index := range dispatcher.work {
		dispatcher.work[index] = true
	}
	if err := outboundWebhookTaskHandler(nil, &storyMutationEventDispatcherStub{}, dispatcher)(t.Context(), asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil)); err != nil {
		t.Fatalf("bounded outbound webhook task error = %v", err)
	}
	if dispatcher.dispatches != outboundWebhookDispatchBatchSize {
		t.Fatalf("bounded dispatch calls = %d, want %d", dispatcher.dispatches, outboundWebhookDispatchBatchSize)
	}
}

func TestOutboundWebhookTaskSurfacesInfrastructureErrors(t *testing.T) {
	t.Parallel()
	storyFailure := errors.New("story event dispatch failed")
	storyEvents := &storyMutationEventDispatcherStub{err: storyFailure}
	dispatcher := &outboundWebhookDispatcherStub{}
	err := outboundWebhookTaskHandler(nil, storyEvents, dispatcher)(t.Context(), asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil))
	if !errors.Is(err, storyFailure) || dispatcher.dispatches != 0 {
		t.Fatalf("story event error = %v, dispatches=%d", err, dispatcher.dispatches)
	}

	recoverFailure := errors.New("recover failed")
	dispatcher = &outboundWebhookDispatcherStub{recoverErr: recoverFailure}
	err = outboundWebhookTaskHandler(nil, &storyMutationEventDispatcherStub{}, dispatcher)(t.Context(), asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil))
	if !errors.Is(err, recoverFailure) || dispatcher.dispatches != 0 {
		t.Fatalf("recovery error = %v, dispatches=%d", err, dispatcher.dispatches)
	}

	dispatchFailure := errors.New("dispatch failed")
	dispatcher = &outboundWebhookDispatcherStub{dispatchErr: dispatchFailure}
	err = outboundWebhookTaskHandler(nil, &storyMutationEventDispatcherStub{}, dispatcher)(t.Context(), asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil))
	if !errors.Is(err, dispatchFailure) || dispatcher.dispatches != 1 {
		t.Fatalf("dispatch error = %v, dispatches=%d", err, dispatcher.dispatches)
	}
}

func TestRegisterOutboundWebhookTask(t *testing.T) {
	t.Parallel()
	mux := asynq.NewServeMux()
	if err := registerOutboundWebhookTask(mux, nil, &storyMutationEventDispatcherStub{}, &outboundWebhookDispatcherStub{}); err != nil {
		t.Fatalf("register outbound webhook task: %v", err)
	}
	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeOutboundWebhookDispatch, nil))
	if handler == nil || pattern != tasks.TypeOutboundWebhookDispatch {
		t.Fatalf("registered handler=%v pattern=%q", handler, pattern)
	}
}
