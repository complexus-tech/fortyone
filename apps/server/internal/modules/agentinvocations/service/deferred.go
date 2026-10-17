package agentinvocations

import (
	"context"
)

// DeferredAction separates an external authoritative writer from the fenced
// invocation transaction. BeginApply only stages durable work. PollApply reads
// its receipt; it never repeats the write or claims success while still pending.
type DeferredAction interface {
	UsesDeferredAction(Run) bool
	BeginApply(context.Context, Run, Proposal) error
	PollApply(context.Context, Run, Proposal) (Outcome, bool, error)
}

type deferredStore interface {
	BeginAction(context.Context, Run, func(context.Context, Run) error) (Run, error)
	DeferAction(context.Context, Run) error
}

func (s *Service) prepareDeferredAction(ctx context.Context, run *Run) (bool, error) {
	runtime, ok := s.runtime.(DeferredAction)
	if !ok || !runtime.UsesDeferredAction(*run) {
		return false, nil
	}
	store, ok := s.store.(deferredStore)
	if !ok {
		return false, ErrUnavailable
	}
	if run.ActionStartedAt == nil {
		updated, err := store.BeginAction(ctx, *run, func(txCtx context.Context, current Run) error {
			if err := s.runtime.Authorize(txCtx, current); err != nil {
				return err
			}
			return runtime.BeginApply(txCtx, current, *current.Proposal)
		})
		if err != nil {
			return false, err
		}
		*run = updated
	}
	outcome, ready, err := runtime.PollApply(ctx, *run, *run.Proposal)
	if err != nil {
		return false, err
	}
	if !ready {
		return true, store.DeferAction(ctx, *run)
	}
	updated, err := s.store.Execute(ctx, *run, func(txCtx context.Context, current Run) (Outcome, error) {
		if err := s.runtime.Authorize(txCtx, current); err != nil {
			return Outcome{}, err
		}
		return outcome, nil
	})
	if err == nil {
		*run = updated
	}
	return false, err
}
