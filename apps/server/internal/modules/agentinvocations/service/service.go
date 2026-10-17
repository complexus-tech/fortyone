// Package agentinvocations coordinates durable assistant work independently of
// its conversation surface or model provider.
package agentinvocations

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	"github.com/google/uuid"
)

type Run = domain.Run
type Input = domain.Input
type Proposal = domain.Proposal
type Outcome = domain.Outcome
type Control = domain.Control
type ContinuationQuery = domain.ContinuationQuery
type Failure = domain.Failure
type Status = domain.Status

const (
	StatusQueued           = domain.StatusQueued
	StatusRunning          = domain.StatusRunning
	StatusNeedsInput       = domain.StatusNeedsInput
	StatusAwaitingApproval = domain.StatusAwaitingApproval
	StatusSucceeded        = domain.StatusSucceeded
	StatusFailed           = domain.StatusFailed
	StatusCancelled        = domain.StatusCancelled
)

var (
	ErrInvalidInput     = domain.ErrInvalidInput
	ErrNotFound         = domain.ErrNotFound
	ErrConflict         = domain.ErrConflict
	ErrLeaseLost        = domain.ErrLeaseLost
	ErrSourceChanged    = domain.ErrSourceChanged
	ErrPermissionDenied = domain.ErrPermissionDenied
	ErrUnavailable      = domain.ErrUnavailable
)

// Runtime owns source-specific authorization and model context. Apply and
// Publish run inside an atomic fenced database unit and MUST NOT perform network
// I/O. Prepare owns bounded provider I/O outside transactions. Never let model
// arguments choose Run identity, destination, or permission scope.
type Runtime interface {
	Authorize(context.Context, Run) error
	Prepare(context.Context, Run) (Proposal, error)
	Apply(context.Context, Run, Proposal) (Outcome, error)
	Publish(context.Context, Run, Outcome) (uuid.UUID, error)
}

// Delivery runs after the reply commits and must deduplicate by run/reply ID.
// Its successful acknowledgement is fenced before the run becomes terminal.
type Delivery interface {
	Deliver(context.Context, Run, Outcome, uuid.UUID) error
}

// ProposalPublisher writes an ordinary assistant question before approval.
// It runs inside the same fenced unit as its distinct proposal reply receipt.
type ProposalPublisher interface {
	PublishProposal(context.Context, Run, Proposal) (uuid.UUID, error)
}

type FailurePublisher interface {
	PublishFailure(context.Context, Run, string) (uuid.UUID, error)
}

type Store interface {
	Accept(context.Context, Input) (Run, bool, error)
	ResolveContinuation(context.Context, ContinuationQuery) (*uuid.UUID, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Run, error)
	List(context.Context, uuid.UUID, uuid.UUID) ([]Run, error)
	Control(context.Context, Control, string) (Run, error)
	Pending(context.Context, int) ([]uuid.UUID, error)
	Claim(context.Context, uuid.UUID) (Run, error)
	SaveProposal(context.Context, Run, Proposal) (Run, error)
	Execute(context.Context, Run, func(context.Context, Run) (Outcome, error)) (Run, error)
	Publish(context.Context, Run, func(context.Context, Run) (uuid.UUID, error)) (Run, error)
	PublishProposal(context.Context, Run, func(context.Context, Run) (uuid.UUID, error)) (Run, error)
	AwaitApproval(context.Context, Run) (Run, error)
	BeginFailure(context.Context, Run, string) (Run, error)
	PublishFailure(context.Context, Run, func(context.Context, Run) (uuid.UUID, error)) (Run, error)
	DeferFailure(context.Context, Run) error
	CompleteFailure(context.Context, Run) (Run, error)
	Complete(context.Context, Run) (Run, error)
	Fail(context.Context, Run, string, bool) error
}

type Service struct {
	store   Store
	runtime Runtime
}

func New(store Store, runtime Runtime) *Service { return &Service{store: store, runtime: runtime} }

func (s *Service) Accept(ctx context.Context, input Input) (Run, bool, error) {
	if err := input.Validate(); err != nil {
		return Run{}, false, err
	}
	if s == nil || s.store == nil {
		return Run{}, false, ErrUnavailable
	}
	return s.store.Accept(ctx, input)
}

func (s *Service) ResolveContinuation(ctx context.Context, query ContinuationQuery) (*uuid.UUID, error) {
	if s == nil || s.store == nil {
		return nil, ErrUnavailable
	}
	if query.WorkspaceID == uuid.Nil || query.RequesterID == uuid.Nil || query.EntityID == uuid.Nil || query.ConversationID == uuid.Nil || query.ReplyToID == uuid.Nil || query.SourceID == uuid.Nil || strings.TrimSpace(query.Surface) == "" || len(query.Surface) > 64 {
		return nil, ErrInvalidInput
	}
	return s.store.ResolveContinuation(ctx, query)
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (Run, error) {
	if s == nil || s.store == nil {
		return Run{}, ErrUnavailable
	}
	if workspaceID == uuid.Nil || id == uuid.Nil {
		return Run{}, ErrInvalidInput
	}
	return s.store.Get(ctx, workspaceID, id)
}

// The surface adapter must authorize the destination before listing shared runs.
func (s *Service) List(ctx context.Context, workspaceID, entityID uuid.UUID) ([]Run, error) {
	if s == nil || s.store == nil {
		return nil, ErrUnavailable
	}
	if workspaceID == uuid.Nil || entityID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return s.store.List(ctx, workspaceID, entityID)
}

func (s *Service) Cancel(ctx context.Context, control Control) (Run, error) {
	return s.control(ctx, control, "cancel")
}
func (s *Service) Retry(ctx context.Context, control Control) (Run, error) {
	return s.control(ctx, control, "retry")
}
func (s *Service) Approve(ctx context.Context, control Control) (Run, error) {
	return s.control(ctx, control, "approve")
}

func (s *Service) control(ctx context.Context, control Control, operation string) (Run, error) {
	if s == nil || s.store == nil {
		return Run{}, ErrUnavailable
	}
	if control.WorkspaceID == uuid.Nil || control.RequesterID == uuid.Nil || control.RunID == uuid.Nil || control.ExpectedVersion < 1 {
		return Run{}, ErrInvalidInput
	}
	return s.store.Control(ctx, control, operation)
}

func (s *Service) ProcessPending(ctx context.Context, limit int) (int, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return 0, ErrUnavailable
	}
	if limit < 1 || limit > 100 {
		return 0, ErrInvalidInput
	}
	ids, err := s.store.Pending(ctx, limit)
	if err != nil {
		return 0, err
	}
	var combined error
	processed := 0
	for _, id := range ids {
		if ctx.Err() != nil {
			return processed, errors.Join(combined, ctx.Err())
		}
		if err := s.Process(ctx, id); err != nil {
			combined = errors.Join(combined, err)
		}
		processed++
	}
	return processed, combined
}

func (s *Service) Process(ctx context.Context, id uuid.UUID) error {
	if s == nil || s.store == nil || s.runtime == nil {
		return ErrUnavailable
	}
	if id == uuid.Nil {
		return ErrInvalidInput
	}
	run, err := s.store.Claim(ctx, id)
	if errors.Is(err, ErrLeaseLost) {
		return nil
	}
	if err != nil {
		return err
	}
	// The database lease is two minutes. Bound work below that duration; a dead
	// process is recovered without an unbounded heartbeat goroutine.
	workCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	err = s.processClaim(workCtx, &run)
	if err == nil || errors.Is(err, ErrLeaseLost) {
		return nil
	}
	summary, retry := safeFailure(err)
	cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cleanupCancel()
	if run.FailurePending {
		return errors.Join(err, s.deferFailure(cleanupCtx, run, err))
	}
	_, canPublish := s.runtime.(FailurePublisher)
	terminal := !retry || (run.RetryAttemptCount >= domain.MaximumAttempts && !run.HasDurableResponse())
	if canPublish && terminal && run.Outcome == nil && run.ReplyCommentID == nil && !cannotPublishFailure(err) {
		updated, publishErr := s.store.BeginFailure(cleanupCtx, run, summary)
		if publishErr != nil {
			return errors.Join(err, publishErr)
		}
		run = updated
		publishCtx, cancel := context.WithTimeout(cleanupCtx, 5*time.Second)
		publishErr = s.publishFailure(publishCtx, &run)
		cancel()
		if publishErr == nil || errors.Is(publishErr, ErrLeaseLost) {
			return err
		}
		return errors.Join(err, publishErr, s.deferFailure(cleanupCtx, run, publishErr))
	}
	if persistErr := s.store.Fail(cleanupCtx, run, summary, retry); persistErr != nil && !errors.Is(persistErr, ErrLeaseLost) {
		return errors.Join(err, fmt.Errorf("persist agent failure: %w", persistErr))
	}
	return err
}

func (s *Service) processClaim(ctx context.Context, run *Run) error {
	if run.FailurePending {
		return s.publishFailure(ctx, run)
	}
	if err := s.runtime.Authorize(ctx, *run); err != nil {
		return err
	}
	if run.Proposal == nil {
		prepareCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		proposal, err := s.runtime.Prepare(prepareCtx, *run)
		cancel()
		if err != nil {
			return err
		}
		if err := proposal.Validate(); err != nil {
			return err
		}
		updated, err := s.store.SaveProposal(ctx, *run, proposal)
		if err != nil {
			return err
		}
		*run = updated
	}
	if run.Proposal.RequiresApproval && run.ApprovedAt == nil {
		return s.publishApproval(ctx, run)
	}
	if run.Outcome == nil && run.Proposal.HasAction() {
		waiting, err := s.prepareDeferredAction(ctx, run)
		if err != nil || waiting {
			return err
		}
	}
	if run.Outcome == nil {
		updated, err := s.store.Execute(ctx, *run, func(txCtx context.Context, current Run) (Outcome, error) {
			if err := s.runtime.Authorize(txCtx, current); err != nil {
				return Outcome{}, err
			}
			if current.Proposal == nil {
				return Outcome{}, ErrConflict
			}
			if current.Proposal.HasAction() {
				return s.runtime.Apply(txCtx, current, *current.Proposal)
			}
			return Outcome{Reply: current.Proposal.Reply}, nil
		})
		if err != nil {
			return err
		}
		*run = updated
	}
	if run.ReplyCommentID == nil {
		updated, err := s.store.Publish(ctx, *run, func(txCtx context.Context, current Run) (uuid.UUID, error) {
			if err := s.runtime.Authorize(txCtx, current); err != nil {
				return uuid.Nil, err
			}
			if current.Outcome == nil {
				return uuid.Nil, ErrConflict
			}
			return s.runtime.Publish(txCtx, current, *current.Outcome)
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
		if err := delivery.Deliver(ctx, *run, *run.Outcome, *run.ReplyCommentID); err != nil {
			return err
		}
	}
	updated, err := s.store.Complete(ctx, *run)
	if err == nil {
		*run = updated
	}
	return err
}

func (s *Service) publishApproval(ctx context.Context, run *Run) error {
	if run.ProposalReplyID == nil {
		publisher, ok := s.runtime.(ProposalPublisher)
		if !ok {
			return ErrUnavailable
		}
		updated, err := s.store.PublishProposal(ctx, *run, func(txCtx context.Context, current Run) (uuid.UUID, error) {
			if err := s.runtime.Authorize(txCtx, current); err != nil {
				return uuid.Nil, err
			}
			if current.Proposal == nil {
				return uuid.Nil, ErrConflict
			}
			return publisher.PublishProposal(txCtx, current, *current.Proposal)
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
		if err := delivery.Deliver(ctx, *run, Outcome{Reply: run.Proposal.Reply}, *run.ProposalReplyID); err != nil {
			return err
		}
	}
	updated, err := s.store.AwaitApproval(ctx, *run)
	if err == nil {
		*run = updated
	}
	return err
}

func safeFailure(err error) (string, bool) {
	var failure *Failure
	if errors.As(err, &failure) && strings.TrimSpace(failure.Summary) != "" && len(failure.Summary) <= 2000 {
		return failure.Summary, false
	}
	switch {
	case errors.Is(err, ErrPermissionDenied):
		return "Maya can no longer access this request.", false
	case errors.Is(err, ErrSourceChanged):
		return "This request changed. Ask Maya again using the current text.", false
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrConflict):
		return "Maya could not safely apply this request. Review it and try again.", false
	default:
		return "Maya could not finish this request. Please try again.", true
	}
}
