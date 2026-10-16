package agentinvocations

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) publishFailure(ctx context.Context, run *Run) error {
	if err := s.runtime.Authorize(ctx, *run); err != nil {
		return err
	}
	if run.FailureReplyID == nil {
		publisher, ok := s.runtime.(FailurePublisher)
		if !ok {
			return ErrUnavailable
		}
		updated, err := s.store.PublishFailure(ctx, *run, func(txCtx context.Context, current Run) (uuid.UUID, error) {
			if err := s.runtime.Authorize(txCtx, current); err != nil {
				return uuid.Nil, err
			}
			return publisher.PublishFailure(txCtx, current, current.Summary)
		})
		if err != nil {
			return err
		}
		*run = updated
	}
	if delivery, ok := s.runtime.(Delivery); ok {
		if err := s.runtime.Authorize(ctx, *run); err != nil {
			return err
		}
		if err := delivery.Deliver(ctx, *run, Outcome{Reply: run.Summary}, *run.FailureReplyID); err != nil {
			return err
		}
	}
	updated, err := s.store.CompleteFailure(ctx, *run)
	if err == nil {
		*run = updated
	}
	return err
}

func (s *Service) deferFailure(ctx context.Context, run Run, cause error) error {
	if errors.Is(cause, ErrLeaseLost) {
		return nil
	}
	if cannotPublishFailure(cause) || errors.Is(cause, ErrInvalidInput) {
		return s.store.Fail(ctx, run, run.Summary, false)
	}
	return s.store.DeferFailure(ctx, run)
}

func cannotPublishFailure(err error) bool {
	return errors.Is(err, ErrPermissionDenied) || errors.Is(err, ErrSourceChanged)
}
