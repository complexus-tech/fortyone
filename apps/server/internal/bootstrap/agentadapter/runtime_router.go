package agentadapter

import (
	"context"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/google/uuid"
)

// RuntimeRouter keeps surface context and publication policy in their adapters.
// Unknown surfaces fail closed instead of falling through to a story runtime.
type RuntimeRouter struct {
	Story    *CommentRuntime
	Document *DocumentRuntime
	Chat     *ChatRuntime
}

func (r *RuntimeRouter) target(run agentruns.Run) (agentruns.Runtime, error) {
	switch run.Surface {
	case StoryCommentSurface:
		if r.Story != nil {
			return r.Story, nil
		}
	case DocumentCommentSurface:
		if r.Document != nil {
			return r.Document, nil
		}
	case ChatMessageSurface:
		if r.Chat != nil {
			return r.Chat, nil
		}
	}
	return nil, agentruns.ErrPermissionDenied
}
func (r *RuntimeRouter) Authorize(ctx context.Context, run agentruns.Run) error {
	target, err := r.target(run)
	if err != nil {
		return err
	}
	return target.Authorize(ctx, run)
}
func (r *RuntimeRouter) Prepare(ctx context.Context, run agentruns.Run) (agentruns.Proposal, error) {
	target, err := r.target(run)
	if err != nil {
		return agentruns.Proposal{}, err
	}
	return target.Prepare(ctx, run)
}
func (r *RuntimeRouter) Apply(ctx context.Context, run agentruns.Run, p agentruns.Proposal) (agentruns.Outcome, error) {
	target, err := r.target(run)
	if err != nil {
		return agentruns.Outcome{}, err
	}
	return target.Apply(ctx, run, p)
}
func (r *RuntimeRouter) Publish(ctx context.Context, run agentruns.Run, o agentruns.Outcome) (uuid.UUID, error) {
	target, err := r.target(run)
	if err != nil {
		return uuid.Nil, err
	}
	return target.Publish(ctx, run, o)
}
func (r *RuntimeRouter) PublishProposal(ctx context.Context, run agentruns.Run, p agentruns.Proposal) (uuid.UUID, error) {
	target, err := r.target(run)
	if err != nil {
		return uuid.Nil, err
	}
	publisher, ok := target.(agentruns.ProposalPublisher)
	if !ok {
		return uuid.Nil, agentruns.ErrUnavailable
	}
	return publisher.PublishProposal(ctx, run, p)
}
func (r *RuntimeRouter) PublishFailure(ctx context.Context, run agentruns.Run, summary string) (uuid.UUID, error) {
	target, err := r.target(run)
	if err != nil {
		return uuid.Nil, err
	}
	publisher, ok := target.(agentruns.FailurePublisher)
	if !ok {
		return uuid.Nil, agentruns.ErrUnavailable
	}
	return publisher.PublishFailure(ctx, run, summary)
}
func (r *RuntimeRouter) Deliver(ctx context.Context, run agentruns.Run, o agentruns.Outcome, id uuid.UUID) error {
	target, err := r.target(run)
	if err != nil {
		return err
	}
	if delivery, ok := target.(agentruns.Delivery); ok {
		return delivery.Deliver(ctx, run, o, id)
	}
	return nil
}
func (r *RuntimeRouter) UsesDeferredAction(run agentruns.Run) bool {
	return run.Surface == DocumentCommentSurface
}
func (r *RuntimeRouter) BeginApply(ctx context.Context, run agentruns.Run, p agentruns.Proposal) error {
	if r.Document == nil || run.Surface != DocumentCommentSurface {
		return agentruns.ErrPermissionDenied
	}
	return r.Document.BeginApply(ctx, run, p)
}
func (r *RuntimeRouter) PollApply(ctx context.Context, run agentruns.Run, p agentruns.Proposal) (agentruns.Outcome, bool, error) {
	if r.Document == nil || run.Surface != DocumentCommentSurface {
		return agentruns.Outcome{}, false, agentruns.ErrPermissionDenied
	}
	return r.Document.PollApply(ctx, run, p)
}
