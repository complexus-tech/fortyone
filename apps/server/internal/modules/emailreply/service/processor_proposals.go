package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type decisionContext struct {
	Thread         Thread
	Inbound        InboundEmail
	InboundMessage Message
	EventID        string
	ReceiptID      uuid.UUID
	ReplyTokenHash []byte
	Decision       AgentDecision
}

type resolvedReply struct {
	Copy EmailCopy
	Kind string
	Key  string
}

func (processor *Processor) resolveDecision(ctx context.Context, input decisionContext) (resolvedReply, error) {
	decision := input.Decision
	switch decision.Intent {
	case AgentIntentAnswer, AgentIntentClarify, AgentIntentRefuse:
		if decision.Copy == nil {
			return resolvedReply{}, errors.New("email agent returned copy-less response")
		}
		copy := conversationReplyCopy(*decision.Copy, input.Inbound.Subject)
		kind := MessageKindAnswer
		if decision.Source == AgentDecisionSourceFallback {
			kind = MessageKindError
		}
		return resolvedReply{Copy: copy, Kind: kind, Key: "decision"}, nil
	case AgentIntentPropose:
		if decision.Copy == nil || decision.Proposal == nil {
			return resolvedReply{}, errors.New("email agent returned an incomplete proposal")
		}
		payload, err := json.Marshal(decision.Proposal)
		if err != nil {
			return resolvedReply{}, fmt.Errorf("encode email action proposal: %w", err)
		}
		target, err := proposalTarget(*decision.Proposal)
		if err != nil {
			return resolvedReply{}, err
		}
		_, _, err = processor.store.RegisterProposal(ctx, ProposalInput{
			ThreadID:              input.Thread.ID,
			WorkspaceID:           input.Thread.WorkspaceID,
			UserID:                input.Thread.UserID,
			SourceMessageID:       input.InboundMessage.ID,
			IdempotencyKey:        "proposal:" + input.EventID,
			ActionKind:            string(decision.Proposal.Kind),
			EntityType:            proposalEntityType(decision.Proposal.Kind),
			EntityID:              target.ID,
			ExpectedEntityVersion: target.ExpectedUpdatedAt.UTC().Format(time.RFC3339Nano),
			ProposedDiff:          payload,
			ExpiresAt:             processor.now().UTC().Add(emailReplyProposalLifetime),
			Now:                   processor.now().UTC(),
		})
		if err != nil {
			return resolvedReply{}, err
		}
		return resolvedReply{
			Copy: conversationReplyCopy(*decision.Copy, input.Inbound.Subject),
			Kind: MessageKindProposal,
			Key:  "proposal",
		}, nil
	case AgentIntentCancel:
		return processor.cancelProposal(ctx, input)
	case AgentIntentConfirm:
		return processor.confirmProposal(ctx, input)
	default:
		return resolvedReply{}, fmt.Errorf("unsupported email agent intent %q", decision.Intent)
	}
}

func conversationReplyCopy(copy EmailCopy, inboundSubject string) EmailCopy {
	copy.Subject = replyEmailSubject(inboundSubject)
	return copy
}

func (processor *Processor) cancelProposal(ctx context.Context, input decisionContext) (resolvedReply, error) {
	proposal, _, err := processor.store.DecideProposal(ctx, ProposalDecision{
		ProposalID: input.Decision.Command.ProposalID, ThreadID: input.Thread.ID,
		WorkspaceID: input.Thread.WorkspaceID, UserID: input.Thread.UserID,
		ReplyTokenHash: input.ReplyTokenHash, Decision: ProposalCancelled, Now: processor.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, ErrProposalExpired) || errors.Is(err, ErrProposalConflict) {
			return deterministicReply(input.Inbound.Subject, "That proposed change is no longer pending, so I haven't changed anything.", "Tell me what you want to update and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, err
	}
	return deterministicReply(input.Inbound.Subject, "Cancelled. I haven't applied the proposed change: "+proposalSummary(proposal)+".", "Tell me if you'd like to take a different next step."), nil
}

func (processor *Processor) confirmProposal(ctx context.Context, input decisionContext) (resolvedReply, error) {
	proposalRecord, _, err := processor.store.DecideProposal(ctx, ProposalDecision{
		ProposalID: input.Decision.Command.ProposalID, ThreadID: input.Thread.ID,
		WorkspaceID: input.Thread.WorkspaceID, UserID: input.Thread.UserID,
		ReplyTokenHash: input.ReplyTokenHash, Decision: ProposalConfirmed, Now: processor.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, ErrProposalExpired) || errors.Is(err, ErrProposalConflict) {
			return deterministicReply(input.Inbound.Subject, "That proposed change is no longer available, so I haven't changed anything.", "Tell me the latest value and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, err
	}
	if proposalRecord.Status == ProposalApplied {
		return deterministicReply(input.Inbound.Subject, "That change has already been applied: "+proposalSummary(proposalRecord)+".", "Nothing else was changed."), nil
	}

	claimed, apply, err := processor.store.ClaimProposalApply(ctx, ProposalApplyClaim{
		ProposalID: proposalRecord.ID, ThreadID: input.Thread.ID, WorkspaceID: input.Thread.WorkspaceID,
		UserID: input.Thread.UserID, Now: processor.now().UTC(), RetryAfter: emailReplyApplyRetryAfter,
	})
	if err != nil {
		return resolvedReply{}, err
	}
	if !apply {
		return deterministicReply(input.Inbound.Subject, "That change has already been applied: "+proposalSummary(claimed)+".", "Nothing else was changed."), nil
	}
	var proposal ActionProposal
	if err := json.Unmarshal(claimed.ProposedDiff, &proposal); err != nil {
		return resolvedReply{}, processor.failProposalApply(ctx, claimed, fmt.Errorf("decode confirmed email proposal: %w", err))
	}
	if proposal.WorkspaceID != claimed.WorkspaceID || proposal.ActorID != claimed.UserID {
		return resolvedReply{}, processor.failProposalApply(ctx, claimed, errors.New("confirmed proposal actor binding is invalid"))
	}
	if err := processor.context.AuthorizeProposal(ctx, proposal); err != nil {
		if errors.Is(err, ErrActionUnauthorized) || errors.Is(err, ErrActionConflict) {
			if completionErr := processor.completeProposalFailure(ctx, claimed, err); completionErr != nil {
				return resolvedReply{}, errors.Join(err, completionErr)
			}
			return deterministicReply(input.Inbound.Subject, "I couldn't apply that change because the item or your access has changed since the preview.", "Tell me the latest state you want and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, processor.failProposalApply(ctx, claimed, err)
	}
	if err := processor.mutations.Apply(ctx, proposal); err != nil {
		if errors.Is(err, ErrActionConflict) {
			alreadyApplied, reconcileErr := processor.context.ProposalAlreadyApplied(ctx, proposal)
			if reconcileErr != nil {
				return resolvedReply{}, processor.failProposalApply(ctx, claimed, errors.Join(err, reconcileErr))
			}
			if alreadyApplied {
				return processor.completeAppliedProposal(ctx, input, claimed, proposal)
			}
		}
		if errors.Is(err, ErrActionConflict) || errors.Is(err, ErrActionUnauthorized) {
			if completionErr := processor.completeProposalFailure(ctx, claimed, err); completionErr != nil {
				return resolvedReply{}, errors.Join(err, completionErr)
			}
			return deterministicReply(input.Inbound.Subject, "I couldn't apply that change because the item changed after the preview.", "Tell me the latest state you want and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, processor.failProposalApply(ctx, claimed, err)
	}
	return processor.completeAppliedProposal(ctx, input, claimed, proposal)
}

func (processor *Processor) completeAppliedProposal(
	ctx context.Context,
	input decisionContext,
	claimed Proposal,
	proposal ActionProposal,
) (resolvedReply, error) {
	result, _ := json.Marshal(map[string]any{"summary": proposal.Summary})
	if _, _, err := processor.store.CompleteProposalApply(ctx, ProposalApplyCompletion{
		ProposalID: claimed.ID, ThreadID: claimed.ThreadID, WorkspaceID: claimed.WorkspaceID,
		UserID: claimed.UserID, ApplyAttempt: claimed.ApplyAttempt, Status: ProposalApplied,
		Result: result, Now: processor.now().UTC(),
	}); err != nil {
		return resolvedReply{}, err
	}
	return deterministicReply(input.Inbound.Subject, "Done — I applied this change: "+proposal.Summary+".", "Nothing else was changed. Reply if you'd like me to help with the next update."), nil
}

func (processor *Processor) failProposalApply(ctx context.Context, proposal Proposal, cause error) error {
	if err := processor.completeProposalFailure(ctx, proposal, cause); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (processor *Processor) completeProposalFailure(ctx context.Context, proposal Proposal, cause error) error {
	stateCtx, cancel := processorStateContext(ctx)
	defer cancel()
	_, _, err := processor.store.CompleteProposalApply(stateCtx, ProposalApplyCompletion{
		ProposalID: proposal.ID, ThreadID: proposal.ThreadID, WorkspaceID: proposal.WorkspaceID,
		UserID: proposal.UserID, ApplyAttempt: proposal.ApplyAttempt, Status: ProposalFailed,
		Result: json.RawMessage(`{}`), ErrorMessage: truncateProcessorError(cause), Now: processor.now().UTC(),
	})
	return err
}
