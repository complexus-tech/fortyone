package workerbootstrap

import (
	"context"
	"errors"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
)

// emailReplyStoreAdapter is the composition-root translation between the
// messaging persistence model and emailreply's caller-owned worker port.
type emailReplyStoreAdapter struct {
	repository *messagingrepository.Repository
}

var _ emailreply.ProcessorStore = emailReplyStoreAdapter{}

func (adapter emailReplyStoreAdapter) AcquireThreadLease(
	ctx context.Context,
	input emailreply.ThreadLease,
) (func() error, error) {
	release, err := adapter.repository.AcquireEmailThreadProcessingLease(ctx, messaging.EmailThreadProcessingLease{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
	})
	return release, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) GetThread(ctx context.Context, input emailreply.ThreadKey) (emailreply.Thread, error) {
	record, err := adapter.repository.GetEmailThread(ctx, messaging.EmailThreadKey{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
	})
	return toEmailReplyThread(record), mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) AppendMessage(
	ctx context.Context,
	input emailreply.MessageInput,
) (emailreply.Message, bool, error) {
	record, created, err := adapter.repository.AppendEmailMessage(ctx, messaging.EmailMessageInput{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		InboundEventID: input.InboundEventID, IdempotencyKey: input.IdempotencyKey,
		Direction: input.Direction, Role: input.Role, Kind: input.Kind,
		ProviderMessageID: input.ProviderMessageID, InternetMessageID: input.InternetMessageID,
		InReplyToMessageID: input.InReplyToMessageID, Subject: input.Subject, Content: input.Content,
		Context: input.Context, ProviderMetadata: input.ProviderMetadata,
	})
	return toEmailReplyMessage(record), created, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) ListMessages(
	ctx context.Context,
	input emailreply.MessagePageInput,
) (emailreply.MessagePage, error) {
	page, err := adapter.repository.ListEmailMessages(ctx, messaging.EmailMessagePageInput{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		AfterSequence: input.AfterSequence, Limit: input.Limit,
	})
	if err != nil {
		return emailreply.MessagePage{}, mapEmailReplyMessagingError(err)
	}
	messages := make([]emailreply.Message, 0, len(page.Messages))
	for _, message := range page.Messages {
		messages = append(messages, toEmailReplyMessage(message))
	}
	return emailreply.MessagePage{Messages: messages, NextSequence: page.NextSequence, HasMore: page.HasMore}, nil
}

func (adapter emailReplyStoreAdapter) UpdateThreadSummary(
	ctx context.Context,
	input emailreply.ThreadSummaryUpdate,
) (emailreply.Thread, error) {
	record, err := adapter.repository.UpdateEmailThreadSummary(ctx, messaging.EmailThreadSummaryUpdate{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		ExpectedSummaryThroughSequence: input.ExpectedSummaryThroughSequence,
		Summary:                        input.Summary, ThroughSequence: input.ThroughSequence,
	})
	return toEmailReplyThread(record), mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) RegisterProposal(
	ctx context.Context,
	input emailreply.ProposalInput,
) (emailreply.Proposal, bool, error) {
	record, created, err := adapter.repository.RegisterEmailActionProposal(ctx, messaging.EmailActionProposalInput{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		SourceMessageID: input.SourceMessageID, IdempotencyKey: input.IdempotencyKey,
		ActionKind: input.ActionKind, EntityType: input.EntityType, EntityID: input.EntityID,
		ExpectedEntityVersion: input.ExpectedEntityVersion, ProposedDiff: input.ProposedDiff,
		ExpiresAt: input.ExpiresAt, Now: input.Now,
	})
	return toEmailReplyProposal(record), created, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) ListPendingProposals(
	ctx context.Context,
	input emailreply.ProposalListInput,
) ([]emailreply.Proposal, error) {
	records, err := adapter.repository.ListPendingEmailActionProposals(ctx, messaging.EmailActionProposalListInput{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID, Now: input.Now,
	})
	if err != nil {
		return nil, mapEmailReplyMessagingError(err)
	}
	result := make([]emailreply.Proposal, 0, len(records))
	for _, record := range records {
		result = append(result, toEmailReplyProposal(record))
	}
	return result, nil
}

func (adapter emailReplyStoreAdapter) FindLatestProposalForControl(
	ctx context.Context,
	input emailreply.ProposalControlLookup,
) (emailreply.Proposal, bool, error) {
	record, found, err := adapter.repository.FindLatestEmailActionProposalForControl(ctx, messaging.EmailActionProposalControlLookup{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		Control: messaging.EmailActionProposalStatus(input.Control),
	})
	return toEmailReplyProposal(record), found, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) DecideProposal(
	ctx context.Context,
	input emailreply.ProposalDecision,
) (emailreply.Proposal, bool, error) {
	record, changed, err := adapter.repository.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: input.ProposalID, ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID,
		UserID: input.UserID, ReplyTokenHash: input.ReplyTokenHash,
		Decision: messaging.EmailActionProposalStatus(input.Decision), Now: input.Now,
	})
	return toEmailReplyProposal(record), changed, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) ClaimProposalApply(
	ctx context.Context,
	input emailreply.ProposalApplyClaim,
) (emailreply.Proposal, bool, error) {
	record, apply, err := adapter.repository.ClaimEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyClaim{
		ProposalID: input.ProposalID, ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID,
		UserID: input.UserID, Now: input.Now, RetryAfter: input.RetryAfter,
	})
	return toEmailReplyProposal(record), apply, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) CompleteProposalApply(
	ctx context.Context,
	input emailreply.ProposalApplyCompletion,
) (emailreply.Proposal, bool, error) {
	record, changed, err := adapter.repository.CompleteEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyCompletion{
		ProposalID: input.ProposalID, ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID,
		UserID: input.UserID, ApplyAttempt: input.ApplyAttempt,
		Status: messaging.EmailActionProposalStatus(input.Status), Result: input.Result,
		ErrorMessage: input.ErrorMessage, Now: input.Now,
	})
	return toEmailReplyProposal(record), changed, mapEmailReplyMessagingError(err)
}

func (adapter emailReplyStoreAdapter) HasEarlierInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID string,
	currentID uuid.UUID,
) (bool, error) {
	return adapter.repository.HasEarlierInboundEvent(ctx, provider, externalWorkspaceID, currentID)
}

func (adapter emailReplyStoreAdapter) GetInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID, externalEventID string,
) (emailreply.InboundEvent, error) {
	record, err := adapter.repository.GetInboundEvent(ctx, provider, externalWorkspaceID, externalEventID)
	return toEmailReplyInboundEvent(record), err
}

func (adapter emailReplyStoreAdapter) StartInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID, externalEventID string,
) (emailreply.InboundEvent, bool, error) {
	record, claimed, err := adapter.repository.StartInboundEvent(ctx, provider, externalWorkspaceID, externalEventID)
	return toEmailReplyInboundEvent(record), claimed, err
}

func (adapter emailReplyStoreAdapter) CompleteInboundEvent(
	ctx context.Context,
	id uuid.UUID,
	status, message string,
) error {
	return adapter.repository.CompleteInboundEvent(ctx, id, status, message)
}

func (adapter emailReplyStoreAdapter) StartOutboundDelivery(
	ctx context.Context,
	input emailreply.OutboundDeliveryInput,
) (emailreply.OutboundDelivery, bool, error) {
	record, claimed, err := adapter.repository.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		ExternalWorkspaceID:     input.ExternalWorkspaceID,
		ExternalRecipientUserID: input.ExternalRecipientUserID,
		InboundEventID:          input.InboundEventID, IdempotencyKey: input.IdempotencyKey,
		ExternalChannelID: input.ExternalChannelID, ExternalThreadID: input.ExternalThreadID,
		Purpose: input.Purpose, ExpiresAt: input.ExpiresAt,
	})
	return toEmailReplyOutboundDelivery(record), claimed, err
}

func (adapter emailReplyStoreAdapter) SetOutboundDeliveryContentAndProviderPayload(
	ctx context.Context,
	id uuid.UUID,
	content string,
	payload []byte,
) error {
	return adapter.repository.SetOutboundDeliveryContentAndProviderPayload(ctx, id, content, payload)
}

func (adapter emailReplyStoreAdapter) CompleteOutboundDelivery(
	ctx context.Context,
	id uuid.UUID,
	externalMessageID string,
) error {
	return adapter.repository.CompleteOutboundDelivery(ctx, id, externalMessageID)
}

func (adapter emailReplyStoreAdapter) FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	return adapter.repository.FailOutboundDelivery(ctx, id, message)
}

func toEmailReplyThread(record messaging.EmailThreadRecord) emailreply.Thread {
	return emailreply.Thread{
		ID: record.ID, WorkspaceID: record.WorkspaceID, UserID: record.UserID,
		RecipientEmail: record.RecipientEmail, ExternalThreadID: record.ExternalThreadID,
		RootInternetMessageID:   record.RootInternetMessageID,
		LatestInternetMessageID: record.LatestInternetMessageID,
		Context:                 append([]byte(nil), record.Context...), Summary: record.Summary,
		SummaryThroughSequence: record.SummaryThroughSequence,
		NextMessageSequence:    record.NextMessageSequence,
	}
}

func toEmailReplyMessage(record messaging.EmailMessageRecord) emailreply.Message {
	return emailreply.Message{
		ID: record.ID, Sequence: record.Sequence, Role: record.Role, Content: record.Content,
		InternetMessageID: record.InternetMessageID, CreatedAt: record.CreatedAt,
	}
}

func toEmailReplyProposal(record messaging.EmailActionProposalRecord) emailreply.Proposal {
	return emailreply.Proposal{
		ID: record.ID, ThreadID: record.ThreadID, WorkspaceID: record.WorkspaceID, UserID: record.UserID,
		SourceMessageID: record.SourceMessageID, ActionKind: record.ActionKind,
		ProposedDiff: append([]byte(nil), record.ProposedDiff...),
		Status:       emailreply.ProposalStatus(record.Status), ApplyAttempt: record.ApplyAttempt,
	}
}

func toEmailReplyInboundEvent(record messagingrepository.InboundEventRecord) emailreply.InboundEvent {
	return emailreply.InboundEvent{
		ID: record.ID, WorkspaceID: record.WorkspaceID,
		ExternalWorkspaceID: record.ExternalWorkspaceID, ExternalEventID: record.ExternalEventID,
		Status: record.Status, PayloadEncrypted: record.PayloadEncrypted,
	}
}

func toEmailReplyOutboundDelivery(record messagingrepository.OutboundDeliveryRecord) emailreply.OutboundDelivery {
	return emailreply.OutboundDelivery{
		ID: record.ID, IdempotencyKey: record.IdempotencyKey,
		ProviderPayload: append([]byte(nil), record.ProviderPayload...), ExpiresAt: record.ExpiresAt,
	}
}

func mapEmailReplyMessagingError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, messaging.ErrInvalidEmailConversation):
		return errors.Join(emailreply.ErrInvalidConversation, err)
	case errors.Is(err, messaging.ErrEmailProposalConflict):
		return errors.Join(emailreply.ErrProposalConflict, err)
	case errors.Is(err, messaging.ErrEmailProposalExpired):
		return errors.Join(emailreply.ErrProposalExpired, err)
	default:
		return err
	}
}
