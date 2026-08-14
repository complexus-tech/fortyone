package emailreply

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

const (
	emailReplyPurpose            = "email_reply"
	emailReplyHistoryPageSize    = 100
	emailReplyRecentTurnCount    = 12
	emailReplyProposalLifetime   = 24 * time.Hour
	emailReplyDeliveryLifetime   = 7 * 24 * time.Hour
	emailReplyApplyRetryAfter    = 2 * time.Minute
	maximumStoredErrorRunes      = 2_000
	maximumSummaryTurnRunes      = 7_500
	maximumSummaryBatchRunes     = 60_000
	maximumSummaryBatchTurnCount = 50
)

// ProcessorStore is the durable inbox, conversation, proposal, and delivery
// boundary used by the worker. The messaging repository implements it.
type ProcessorStore interface {
	messaging.EmailConversationStore
	HasEarlierInboundEvent(ctx context.Context, provider, externalWorkspaceID string, currentID uuid.UUID) (bool, error)
	GetInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (messagingrepository.InboundEventRecord, error)
	StartInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (messagingrepository.InboundEventRecord, bool, error)
	CompleteInboundEvent(ctx context.Context, id uuid.UUID, status, message string) error
	StartOutboundDelivery(ctx context.Context, input messagingrepository.OutboundDeliveryInput) (messagingrepository.OutboundDeliveryRecord, bool, error)
	SetOutboundDeliveryContentAndProviderPayload(ctx context.Context, id uuid.UUID, content string, providerPayload []byte) error
	CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type inboundPayloadOpener interface {
	OpenStoredInboundEmail(sealed string) (StoredInboundEmail, error)
	SealProcessorState(payload []byte) (string, error)
	OpenProcessorState(sealed string) ([]byte, error)
}

type emailDecisionService interface {
	Decide(ctx context.Context, request emailagent.Request) (emailagent.Decision, error)
}

type replyThreadPreparer interface {
	NewReplyToken(ctx context.Context, thread messaging.EmailThreadRecord) (string, error)
	PrepareReply(ctx context.Context, input emailthread.ReplyInput) (emailthread.PreparedReply, error)
}

// AuthorizedContext is rebuilt from current database state for every inbound
// turn. It contains no target the actor cannot currently access.
type AuthorizedContext struct {
	AllowedTeamIDs []uuid.UUID
	Facts          []emailagent.GroundedFact
	Targets        []emailagent.AuthorizedTarget
	Choices        []emailagent.AuthorizedChoice
}

// ContextLoader reauthorizes the actor and reloads the original guidance
// targets, their current versions, and safe choices.
type ContextLoader interface {
	Load(ctx context.Context, thread messaging.EmailThreadRecord) (AuthorizedContext, error)
	AuthorizeProposal(ctx context.Context, proposal emailagent.ActionProposal) error
	CurrentVersion(ctx context.Context, proposal emailagent.ActionProposal) (time.Time, error)
	ProposalAlreadyApplied(ctx context.Context, proposal emailagent.ActionProposal) (bool, error)
}

// MutationApplier invokes domain services only after confirmation-time
// authorization. Implementations must use compare-and-swap domain methods.
type MutationApplier interface {
	Apply(ctx context.Context, proposal emailagent.ActionProposal) error
}

// ProcessorConfig composes the production email reply worker without exposing
// provider payloads to the task queue.
type ProcessorConfig struct {
	Log        *logger.Logger
	Store      ProcessorStore
	Inbound    inboundPayloadOpener
	Agent      emailDecisionService
	Summarizer emailagent.Summarizer
	Threads    replyThreadPreparer
	Mailer     mailer.Service
	Context    ContextLoader
	Mutations  MutationApplier
}

// Processor turns one durable Brevo inbox row into one idempotent threaded
// email response. Every mutation requires an exact CONFIRM reply.
type Processor struct {
	log        *logger.Logger
	store      ProcessorStore
	inbound    inboundPayloadOpener
	agent      emailDecisionService
	summarizer emailagent.Summarizer
	threads    replyThreadPreparer
	mailer     mailer.Service
	context    ContextLoader
	mutations  MutationApplier
	now        func() time.Time
}

func NewProcessor(config ProcessorConfig) (*Processor, error) {
	switch {
	case config.Store == nil:
		return nil, errors.New("email reply processor store is required")
	case config.Inbound == nil:
		return nil, errors.New("email reply payload opener is required")
	case config.Agent == nil:
		return nil, errors.New("email reply agent is required")
	case config.Threads == nil:
		return nil, errors.New("email reply thread service is required")
	case config.Mailer == nil:
		return nil, errors.New("email reply mailer is required")
	case config.Context == nil:
		return nil, errors.New("email reply context loader is required")
	case config.Mutations == nil:
		return nil, errors.New("email reply mutation applier is required")
	}
	return &Processor{
		log:        config.Log,
		store:      config.Store,
		inbound:    config.Inbound,
		agent:      config.Agent,
		summarizer: config.Summarizer,
		threads:    config.Threads,
		mailer:     config.Mailer,
		context:    config.Context,
		mutations:  config.Mutations,
		now:        time.Now,
	}, nil
}

// ProcessEvent processes the encrypted canonical inbox value. The queue task
// supplies identity only, so retries cannot diverge from the accepted webhook.
func (processor *Processor) ProcessEvent(ctx context.Context, externalWorkspaceID, eventID string) (err error) {
	if processor == nil {
		return errors.New("email reply processor is not configured")
	}
	externalWorkspaceID = strings.TrimSpace(externalWorkspaceID)
	eventID = strings.TrimSpace(eventID)
	if externalWorkspaceID == "" || eventID == "" {
		return errors.New("Brevo email reply scope and event id are required")
	}

	inbox, err := processor.store.GetInboundEvent(ctx, Provider, externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	if terminalInboundStatus(inbox.Status) {
		return nil
	}
	if inbox.PayloadEncrypted == nil || strings.TrimSpace(*inbox.PayloadEncrypted) == "" {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("Brevo email reply inbox event has no encrypted payload"))
	}
	stored, err := processor.inbound.OpenStoredInboundEmail(*inbox.PayloadEncrypted)
	if err != nil {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, err)
	}
	if stored.ThreadID == uuid.Nil || stored.WorkspaceID == uuid.Nil || stored.UserID == uuid.Nil ||
		inbox.WorkspaceID == nil || *inbox.WorkspaceID != stored.WorkspaceID ||
		externalWorkspaceID != stored.WorkspaceID.String()+":"+stored.ThreadID.String() {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("Brevo email reply inbox binding does not match its encrypted payload"))
	}

	receipt, claimed, err := processor.store.StartInboundEvent(ctx, Provider, externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	status := "failed"
	statusMessage := ""
	defer func() {
		if err != nil {
			statusMessage = truncateProcessorError(err)
		}
		stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if completeErr := processor.store.CompleteInboundEvent(stateCtx, receipt.ID, status, statusMessage); completeErr != nil {
			if processor.log != nil {
				processor.log.Error(stateCtx, "failed updating Brevo email reply receipt", "error", completeErr, "event_id", eventID)
			}
			if err == nil {
				err = completeErr
			}
		}
	}()

	thread, err := processor.store.GetEmailThread(ctx, messaging.EmailThreadKey{
		ThreadID: stored.ThreadID, WorkspaceID: stored.WorkspaceID, UserID: stored.UserID,
	})
	if err != nil {
		if errors.Is(err, messaging.ErrInvalidEmailConversation) {
			status = "ignored"
			return nil
		}
		return err
	}
	releaseThread, err := processor.store.AcquireEmailThreadProcessingLease(ctx, messaging.EmailThreadProcessingLease{
		ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
	})
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := releaseThread(); releaseErr != nil {
			if processor.log != nil {
				processor.log.Error(context.WithoutCancel(ctx), "failed releasing Maya email thread lease", "error", releaseErr, "thread_id", thread.ID)
			}
			if err == nil {
				err = releaseErr
			}
		}
	}()
	// Authorization and history may have changed while another event held the
	// thread lease, so reload after acquiring it.
	thread, err = processor.store.GetEmailThread(ctx, messaging.EmailThreadKey{
		ThreadID: stored.ThreadID, WorkspaceID: stored.WorkspaceID, UserID: stored.UserID,
	})
	if err != nil {
		return err
	}
	senderEmail, err := normalizedEmailAddress(stored.Email.From.Address)
	if err != nil {
		status = "ignored"
		return nil
	}
	recipientEmail, err := normalizedEmailAddress(thread.RecipientEmail)
	if err != nil || !strings.EqualFold(senderEmail, recipientEmail) {
		status = "ignored"
		return nil
	}
	currentReply := plainInboundReply(stored.Email)
	if currentReply == "" {
		status = "ignored"
		return nil
	}
	replyTokenHash, err := ReplyTokenHash(stored.Email)
	if err != nil {
		status = "ignored"
		return nil
	}
	internetMessageID := safeInternetMessageID(stored.Email.MessageID)
	earlier, err := processor.store.HasEarlierInboundEvent(ctx, Provider, externalWorkspaceID, receipt.ID)
	if err != nil {
		return err
	}
	if earlier {
		return errors.New("an earlier Maya email reply for this thread is still pending")
	}
	providerMessageID := boundedProviderMessageID(stored.Email.MessageID, eventID)
	inboundMessage, _, err := processor.store.AppendEmailMessage(ctx, messaging.EmailMessageInput{
		ThreadID:           thread.ID,
		WorkspaceID:        thread.WorkspaceID,
		UserID:             thread.UserID,
		InboundEventID:     &receipt.ID,
		IdempotencyKey:     "inbound:" + eventID,
		Direction:          messaging.EmailMessageDirectionInbound,
		Role:               messaging.EmailMessageRoleUser,
		Kind:               messaging.EmailMessageKindReply,
		ProviderMessageID:  providerMessageID,
		InternetMessageID:  internetMessageID,
		InReplyToMessageID: safeOptionalInternetMessageID(stored.Email.InReplyTo),
		Subject:            sanitizeEmailSubject(stored.Email.Subject),
		Content:            currentReply,
		ProviderMetadata:   json.RawMessage(`{"provider":"brevo"}`),
	})
	if err != nil {
		return err
	}
	if inboundMessage.InternetMessageID != nil {
		thread.LatestInternetMessageID = *inboundMessage.InternetMessageID
	}
	delivery, send, err := processor.claimReplyDelivery(ctx, externalWorkspaceID, eventID, receipt.ID, thread)
	if err != nil {
		return err
	}
	if !send {
		status = "completed"
		return nil
	}
	deliveryCompleted := false
	defer func() {
		if deliveryCompleted {
			return
		}
		failure := err
		if failure == nil {
			failure = errors.New("Maya email reply processing ended before delivery")
		}
		failCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = processor.store.FailOutboundDelivery(failCtx, delivery.ID, truncateProcessorError(failure))
	}()
	if len(delivery.ProviderPayload) > 0 {
		currentAuthorization, loadErr := processor.context.Load(ctx, thread)
		if loadErr != nil {
			if errors.Is(loadErr, ErrActionUnauthorized) || errors.Is(loadErr, messaging.ErrInvalidEmailConversation) {
				status = "ignored"
				return nil
			}
			return loadErr
		}
		if authorizeErr := processor.authorizeFrozenDelivery(delivery.ProviderPayload, currentAuthorization); authorizeErr != nil {
			status = "ignored"
			return nil
		}
		if err := processor.sendClaimedDelivery(ctx, delivery, thread); err != nil {
			return err
		}
		deliveryCompleted = true
		status = "completed"
		return nil
	}

	messages, err := processor.listUnsummarizedMessages(ctx, thread)
	if err != nil {
		return err
	}
	thread, err = processor.refreshSummaryForReply(ctx, currentReply, thread, messages)
	if err != nil {
		return err
	}
	authorized, err := processor.context.Load(ctx, thread)
	if err != nil {
		if errors.Is(err, ErrActionUnauthorized) {
			status = "ignored"
			return nil
		}
		return err
	}
	pending, err := processor.store.ListPendingEmailActionProposals(ctx, messaging.EmailActionProposalListInput{
		ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID, Now: processor.now().UTC(),
	})
	if err != nil {
		return err
	}
	if control, exact := emailagent.ParseControlCommand(currentReply); exact && len(pending) == 0 {
		status := messaging.EmailActionProposalConfirmed
		if control == emailagent.ControlCancel {
			status = messaging.EmailActionProposalCancelled
		}
		recovered, found, findErr := processor.store.FindLatestEmailActionProposalForControl(ctx, messaging.EmailActionProposalControlLookup{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID, Control: status,
		})
		if findErr != nil {
			return findErr
		}
		if found {
			pending = []messaging.EmailActionProposalRecord{recovered}
		}
	}
	if _, exactControl := emailagent.ParseControlCommand(currentReply); !exactControl && len(pending) == 1 &&
		pending[0].Status == messaging.EmailActionProposalPending && pending[0].SourceMessageID == inboundMessage.ID {
		resumeReply := resumedProposalReply(stored.Email.Subject, pending[0], eventID)
		var persistedProposal emailagent.ActionProposal
		resumeErr := json.Unmarshal(pending[0].ProposedDiff, &persistedProposal)
		if resumeErr == nil {
			resumeErr = processor.context.AuthorizeProposal(ctx, persistedProposal)
		}
		if resumeErr == nil {
			var currentVersion time.Time
			currentVersion, resumeErr = processor.context.CurrentVersion(ctx, persistedProposal)
			if resumeErr == nil {
				target, targetErr := proposalTarget(persistedProposal)
				if targetErr != nil {
					resumeErr = targetErr
				} else if !currentVersion.Equal(target.ExpectedUpdatedAt.UTC()) {
					resumeErr = ErrActionConflict
				}
			}
		}
		if resumeErr != nil {
			if _, _, supersedeErr := processor.store.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
				ProposalID: pending[0].ID, ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
				ReplyTokenHash: replyTokenHash, Decision: messaging.EmailActionProposalSuperseded, Now: processor.now().UTC(),
			}); supersedeErr != nil && !errors.Is(supersedeErr, messaging.ErrEmailProposalConflict) {
				return supersedeErr
			}
			resumeReply = deterministicReply(stored.Email.Subject,
				"I couldn't safely restore the earlier change preview, so I haven't changed anything.",
				"Reply with the update you want and I'll prepare a fresh preview.",
			)
		}
		if err := processor.deliverReply(ctx, delivery, thread, resumeReply, authorized.AllowedTeamIDs); err != nil {
			return err
		}
		deliveryCompleted = true
		status = "completed"
		return nil
	}

	decision, err := processor.agent.Decide(ctx, emailagent.Request{
		WorkspaceID:      thread.WorkspaceID,
		ActorID:          thread.UserID,
		SafetyIdentifier: thread.UserID.String(),
		AllowedTeamIDs:   authorized.AllowedTeamIDs,
		Subject:          sanitizeEmailSubject(stored.Email.Subject),
		Message:          currentReply,
		Summary:          thread.Summary,
		History:          conversationHistory(messages, thread.SummaryThroughSequence, inboundMessage.ID),
		Facts:            authorized.Facts,
		Targets:          authorized.Targets,
		Choices:          authorized.Choices,
		PendingProposals: pendingProposalPreviews(pending),
	})
	if err != nil {
		return err
	}

	reply, err := processor.resolveDecision(ctx, decisionContext{
		Thread:         thread,
		Inbound:        stored.Email,
		InboundMessage: inboundMessage,
		EventID:        eventID,
		ReceiptID:      receipt.ID,
		ReplyTokenHash: replyTokenHash,
		Decision:       decision,
	})
	if err != nil {
		return err
	}
	if err := processor.deliverReply(ctx, delivery, thread, reply, authorized.AllowedTeamIDs); err != nil {
		return err
	}
	deliveryCompleted = true
	status = "completed"
	return nil
}

type decisionContext struct {
	Thread         messaging.EmailThreadRecord
	Inbound        InboundEmail
	InboundMessage messaging.EmailMessageRecord
	EventID        string
	ReceiptID      uuid.UUID
	ReplyTokenHash []byte
	Decision       emailagent.Decision
}

type resolvedReply struct {
	Copy emailagent.EmailCopy
	Kind string
	Key  string
}

func (processor *Processor) resolveDecision(ctx context.Context, input decisionContext) (resolvedReply, error) {
	decision := input.Decision
	switch decision.Intent {
	case emailagent.IntentAnswer, emailagent.IntentClarify, emailagent.IntentRefuse:
		if decision.Copy == nil {
			return resolvedReply{}, errors.New("email agent returned copy-less response")
		}
		copy := conversationReplyCopy(*decision.Copy, input.Inbound.Subject)
		kind := messaging.EmailMessageKindAnswer
		if decision.Source == emailagent.DecisionSourceFallback {
			kind = messaging.EmailMessageKindError
		}
		return resolvedReply{Copy: copy, Kind: kind, Key: "decision"}, nil
	case emailagent.IntentPropose:
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
		_, _, err = processor.store.RegisterEmailActionProposal(ctx, messaging.EmailActionProposalInput{
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
			Kind: messaging.EmailMessageKindProposal,
			Key:  "proposal",
		}, nil
	case emailagent.IntentCancel:
		return processor.cancelProposal(ctx, input)
	case emailagent.IntentConfirm:
		return processor.confirmProposal(ctx, input)
	default:
		return resolvedReply{}, fmt.Errorf("unsupported email agent intent %q", decision.Intent)
	}
}

func conversationReplyCopy(copy emailagent.EmailCopy, inboundSubject string) emailagent.EmailCopy {
	copy.Subject = replyEmailSubject(inboundSubject)
	return copy
}

func (processor *Processor) cancelProposal(ctx context.Context, input decisionContext) (resolvedReply, error) {
	proposal, _, err := processor.store.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: input.Decision.Command.ProposalID(), ThreadID: input.Thread.ID,
		WorkspaceID: input.Thread.WorkspaceID, UserID: input.Thread.UserID,
		ReplyTokenHash: input.ReplyTokenHash, Decision: messaging.EmailActionProposalCancelled, Now: processor.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, messaging.ErrEmailProposalExpired) || errors.Is(err, messaging.ErrEmailProposalConflict) {
			return deterministicReply(input.Inbound.Subject, "That proposed change is no longer pending, so I haven't changed anything.", "Tell me what you want to update and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, err
	}
	return deterministicReply(input.Inbound.Subject, "Cancelled. I haven't applied the proposed change: "+proposalSummary(proposal)+".", "Tell me if you'd like to take a different next step."), nil
}

func (processor *Processor) confirmProposal(ctx context.Context, input decisionContext) (resolvedReply, error) {
	proposalRecord, _, err := processor.store.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: input.Decision.Command.ProposalID(), ThreadID: input.Thread.ID,
		WorkspaceID: input.Thread.WorkspaceID, UserID: input.Thread.UserID,
		ReplyTokenHash: input.ReplyTokenHash, Decision: messaging.EmailActionProposalConfirmed, Now: processor.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, messaging.ErrEmailProposalExpired) || errors.Is(err, messaging.ErrEmailProposalConflict) {
			return deterministicReply(input.Inbound.Subject, "That proposed change is no longer available, so I haven't changed anything.", "Tell me the latest value and I'll prepare a fresh preview."), nil
		}
		return resolvedReply{}, err
	}
	if proposalRecord.Status == messaging.EmailActionProposalApplied {
		return deterministicReply(input.Inbound.Subject, "That change has already been applied: "+proposalSummary(proposalRecord)+".", "Nothing else was changed."), nil
	}

	claimed, apply, err := processor.store.ClaimEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyClaim{
		ProposalID: proposalRecord.ID, ThreadID: input.Thread.ID, WorkspaceID: input.Thread.WorkspaceID,
		UserID: input.Thread.UserID, Now: processor.now().UTC(), RetryAfter: emailReplyApplyRetryAfter,
	})
	if err != nil {
		return resolvedReply{}, err
	}
	if !apply {
		return deterministicReply(input.Inbound.Subject, "That change has already been applied: "+proposalSummary(claimed)+".", "Nothing else was changed."), nil
	}
	var proposal emailagent.ActionProposal
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
	claimed messaging.EmailActionProposalRecord,
	proposal emailagent.ActionProposal,
) (resolvedReply, error) {
	result, _ := json.Marshal(map[string]any{"summary": proposal.Summary})
	if _, _, err := processor.store.CompleteEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyCompletion{
		ProposalID: claimed.ID, ThreadID: claimed.ThreadID, WorkspaceID: claimed.WorkspaceID,
		UserID: claimed.UserID, ApplyAttempt: claimed.ApplyAttempt, Status: messaging.EmailActionProposalApplied,
		Result: result, Now: processor.now().UTC(),
	}); err != nil {
		return resolvedReply{}, err
	}
	return deterministicReply(input.Inbound.Subject, "Done — I applied this change: "+proposal.Summary+".", "Nothing else was changed. Reply if you'd like me to help with the next update."), nil
}

func (processor *Processor) failProposalApply(ctx context.Context, proposal messaging.EmailActionProposalRecord, cause error) error {
	if err := processor.completeProposalFailure(ctx, proposal, cause); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (processor *Processor) completeProposalFailure(ctx context.Context, proposal messaging.EmailActionProposalRecord, cause error) error {
	_, _, err := processor.store.CompleteEmailActionProposalApply(context.WithoutCancel(ctx), messaging.EmailActionProposalApplyCompletion{
		ProposalID: proposal.ID, ThreadID: proposal.ThreadID, WorkspaceID: proposal.WorkspaceID,
		UserID: proposal.UserID, ApplyAttempt: proposal.ApplyAttempt, Status: messaging.EmailActionProposalFailed,
		Result: json.RawMessage(`{}`), ErrorMessage: truncateProcessorError(cause), Now: processor.now().UTC(),
	})
	return err
}

func (processor *Processor) deliverReply(
	ctx context.Context,
	delivery messagingrepository.OutboundDeliveryRecord,
	thread messaging.EmailThreadRecord,
	reply resolvedReply,
	authorizedTeamIDs []uuid.UUID,
) error {
	htmlBody, err := emailagent.RenderHTML(reply.Copy)
	if err != nil {
		return err
	}
	htmlBody = renderMayaReplyHTML(reply.Copy.Subject, htmlBody)
	eventID := strings.TrimPrefix(delivery.IdempotencyKey, "email-reply:")
	messageID := deterministicReplyMessageID(thread.ID, eventID, reply.Key)
	if len(delivery.ProviderPayload) == 0 {
		replyToken, tokenErr := processor.threads.NewReplyToken(ctx, thread)
		if tokenErr != nil {
			_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(tokenErr))
			return tokenErr
		}
		plainPayload, encodeErr := json.Marshal(emailDeliveryPayload{
			To: []string{thread.RecipientEmail}, Subject: reply.Copy.Subject, HTML: htmlBody, PlainText: reply.Copy.PlainText,
			ReplyToken: replyToken, MessageID: messageID, InReplyTo: thread.LatestInternetMessageID,
			References: emailReplyReferences(thread), Kind: reply.Kind, HistoryIdempotencyKey: "outbound:" + eventID,
			AuthorizationVersion: 1, AuthorizedTeamIDs: append([]uuid.UUID(nil), authorizedTeamIDs...),
		})
		if encodeErr != nil {
			_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(encodeErr))
			return fmt.Errorf("encode Maya email delivery: %w", encodeErr)
		}
		sealed, sealErr := processor.inbound.SealProcessorState(plainPayload)
		if sealErr != nil {
			_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(sealErr))
			return sealErr
		}
		frozenPayload, encodeErr := json.Marshal(emailDeliveryEnvelope{Sealed: sealed})
		if encodeErr != nil {
			_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(encodeErr))
			return fmt.Errorf("encode sealed Maya email delivery: %w", encodeErr)
		}
		if persistErr := processor.store.SetOutboundDeliveryContentAndProviderPayload(ctx, delivery.ID, reply.Copy.PlainText, frozenPayload); persistErr != nil {
			_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(persistErr))
			return persistErr
		}
		delivery.ProviderPayload = frozenPayload
	}
	return processor.sendClaimedDelivery(ctx, delivery, thread)
}

func (processor *Processor) claimReplyDelivery(
	ctx context.Context,
	externalWorkspaceID, eventID string,
	receiptID uuid.UUID,
	thread messaging.EmailThreadRecord,
) (messagingrepository.OutboundDeliveryRecord, bool, error) {
	expiresAt := processor.now().UTC().Add(emailReplyDeliveryLifetime)
	userID := thread.UserID
	return processor.store.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider: Provider, WorkspaceID: thread.WorkspaceID, UserID: &userID,
		ExternalWorkspaceID: externalWorkspaceID, ExternalRecipientUserID: thread.UserID.String(),
		InboundEventID: &receiptID, IdempotencyKey: "email-reply:" + eventID,
		ExternalChannelID: thread.RecipientEmail, ExternalThreadID: thread.ExternalThreadID,
		Purpose: emailReplyPurpose, ExpiresAt: &expiresAt,
	})
}

func (processor *Processor) sendClaimedDelivery(
	ctx context.Context,
	delivery messagingrepository.OutboundDeliveryRecord,
	thread messaging.EmailThreadRecord,
) error {
	persisted, err := processor.decodeEmailDeliveryPayload(delivery.ProviderPayload)
	if err != nil {
		_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(err))
		return err
	}
	prepared, err := processor.threads.PrepareReply(ctx, emailthread.ReplyInput{
		Thread: thread, ReplyToken: persisted.ReplyToken, InternetMessageID: persisted.MessageID,
		InReplyTo: persisted.InReplyTo, Subject: persisted.Subject, Content: persisted.PlainText,
		Kind: persisted.Kind, IdempotencyKey: persisted.HistoryIdempotencyKey,
		Context: json.RawMessage(`{"source":"email_reply"}`),
	})
	if err != nil {
		_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(err))
		return err
	}
	persisted.ReplyTo = prepared.ReplyTo
	if delivery.ExpiresAt != nil && !processor.now().UTC().Before(delivery.ExpiresAt.UTC()) {
		err := errors.New("Maya email reply delivery expired before send")
		_ = processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, err.Error())
		return err
	}
	if err := processor.mailer.Send(ctx, persisted.Email()); err != nil {
		if failErr := processor.store.FailOutboundDelivery(context.WithoutCancel(ctx), delivery.ID, truncateProcessorError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	if err := processor.store.CompleteOutboundDelivery(ctx, delivery.ID, persisted.MessageID); err != nil {
		return err
	}
	return nil
}

type emailDeliveryPayload struct {
	To                    []string    `json:"to"`
	Subject               string      `json:"subject"`
	HTML                  string      `json:"html"`
	PlainText             string      `json:"plainText"`
	ReplyToken            string      `json:"replyToken"`
	ReplyTo               string      `json:"-"`
	MessageID             string      `json:"messageId"`
	InReplyTo             string      `json:"inReplyTo"`
	References            []string    `json:"references"`
	Kind                  string      `json:"kind"`
	HistoryIdempotencyKey string      `json:"historyIdempotencyKey"`
	AuthorizationVersion  int         `json:"authorizationVersion"`
	AuthorizedTeamIDs     []uuid.UUID `json:"authorizedTeamIds"`
}

type emailDeliveryEnvelope struct {
	Sealed string `json:"sealed"`
}

func (payload emailDeliveryPayload) Email() mailer.Email {
	return mailer.Email{
		To: payload.To, Subject: payload.Subject, Body: payload.HTML, PlainTextBody: payload.PlainText, IsHTML: true,
		Sender: mailer.SenderProfileMaya, ReplyTo: payload.ReplyTo,
		MessageID: payload.MessageID, InReplyTo: payload.InReplyTo, References: payload.References,
	}
}

func (processor *Processor) decodeEmailDeliveryPayload(raw []byte) (emailDeliveryPayload, error) {
	var envelope emailDeliveryEnvelope
	if len(raw) == 0 || json.Unmarshal(raw, &envelope) != nil || strings.TrimSpace(envelope.Sealed) == "" {
		return emailDeliveryPayload{}, errors.New("persisted Maya email delivery envelope is invalid")
	}
	opened, err := processor.inbound.OpenProcessorState(envelope.Sealed)
	if err != nil {
		return emailDeliveryPayload{}, err
	}
	var payload emailDeliveryPayload
	if json.Unmarshal(opened, &payload) != nil || len(payload.To) != 1 ||
		strings.TrimSpace(payload.Subject) == "" || strings.TrimSpace(payload.HTML) == "" ||
		strings.TrimSpace(payload.PlainText) == "" || strings.TrimSpace(payload.MessageID) == "" ||
		!validReplyToken(payload.ReplyToken) || strings.TrimSpace(payload.Kind) == "" ||
		strings.TrimSpace(payload.HistoryIdempotencyKey) == "" || payload.AuthorizationVersion != 1 {
		return emailDeliveryPayload{}, errors.New("persisted Maya email delivery is invalid")
	}
	for _, teamID := range payload.AuthorizedTeamIDs {
		if teamID == uuid.Nil {
			return emailDeliveryPayload{}, errors.New("persisted Maya email delivery authorization is invalid")
		}
	}
	return payload, nil
}

func (processor *Processor) authorizeFrozenDelivery(raw []byte, current AuthorizedContext) error {
	payload, err := processor.decodeEmailDeliveryPayload(raw)
	if err != nil {
		return err
	}
	allowed := make(map[uuid.UUID]struct{}, len(current.AllowedTeamIDs))
	for _, teamID := range current.AllowedTeamIDs {
		allowed[teamID] = struct{}{}
	}
	for _, requiredTeamID := range payload.AuthorizedTeamIDs {
		if _, ok := allowed[requiredTeamID]; !ok {
			return ErrActionUnauthorized
		}
	}
	return nil
}

func (processor *Processor) listUnsummarizedMessages(ctx context.Context, thread messaging.EmailThreadRecord) ([]messaging.EmailMessageRecord, error) {
	messages := make([]messaging.EmailMessageRecord, 0, max(0, int(thread.NextMessageSequence-thread.SummaryThroughSequence-1)))
	after := thread.SummaryThroughSequence
	for {
		page, err := processor.store.ListEmailMessages(ctx, messaging.EmailMessagePageInput{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
			AfterSequence: after, Limit: emailReplyHistoryPageSize,
		})
		if err != nil {
			return nil, err
		}
		messages = append(messages, page.Messages...)
		if !page.HasMore {
			return messages, nil
		}
		if page.NextSequence <= after {
			return nil, errors.New("email conversation pagination did not advance")
		}
		after = page.NextSequence
	}
}

func (processor *Processor) refreshSummary(
	ctx context.Context,
	thread messaging.EmailThreadRecord,
	messages []messaging.EmailMessageRecord,
) (messaging.EmailThreadRecord, error) {
	cutoffIndex := len(messages) - emailReplyRecentTurnCount
	if cutoffIndex <= 0 || messages[cutoffIndex-1].Sequence <= thread.SummaryThroughSequence {
		return thread, nil
	}
	if processor.summarizer == nil {
		return thread, errors.New("email conversation summarizer is required once history exceeds the recent context window")
	}
	toSummarize := make([]emailagent.HistoryTurn, 0, cutoffIndex)
	sequences := make([]int64, 0, cutoffIndex)
	for _, message := range messages[:cutoffIndex] {
		if message.Sequence <= thread.SummaryThroughSequence || message.Role == messaging.EmailMessageRoleSystem {
			continue
		}
		role, ok := emailAgentRole(message.Role)
		if !ok || strings.TrimSpace(message.Content) == "" {
			continue
		}
		toSummarize = append(toSummarize, emailagent.HistoryTurn{
			Role: role, Text: truncateRunes(strings.TrimSpace(message.Content), maximumSummaryTurnRunes), SentAt: message.CreatedAt,
		})
		sequences = append(sequences, message.Sequence)
	}
	for len(toSummarize) > 0 {
		batchSize := summaryBatchSize(toSummarize)
		generation, err := processor.summarizer.Summarize(ctx, emailagent.SummaryRequest{
			SafetyIdentifier: thread.UserID.String(), PreviousSummary: thread.Summary,
			OmittedTurns: toSummarize[:batchSize],
		})
		if err != nil {
			return thread, err
		}
		updated, err := processor.store.UpdateEmailThreadSummary(ctx, messaging.EmailThreadSummaryUpdate{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
			ExpectedSummaryThroughSequence: thread.SummaryThroughSequence,
			Summary:                        generation.Summary, ThroughSequence: sequences[batchSize-1],
		})
		if err != nil {
			return thread, err
		}
		thread = updated
		toSummarize = toSummarize[batchSize:]
		sequences = sequences[batchSize:]
	}
	return thread, nil
}

func (processor *Processor) refreshSummaryForReply(
	ctx context.Context,
	currentReply string,
	thread messaging.EmailThreadRecord,
	messages []messaging.EmailMessageRecord,
) (messaging.EmailThreadRecord, error) {
	if _, isControl := emailagent.ParseControlCommand(currentReply); isControl {
		return thread, nil
	}
	return processor.refreshSummary(ctx, thread, messages)
}

func summaryBatchSize(turns []emailagent.HistoryTurn) int {
	count, runes := 0, 0
	for count < len(turns) && count < maximumSummaryBatchTurnCount {
		next := len([]rune(turns[count].Text))
		if count > 0 && runes+next > maximumSummaryBatchRunes {
			break
		}
		runes += next
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}

func conversationHistory(messages []messaging.EmailMessageRecord, afterSequence int64, currentMessageID uuid.UUID) []emailagent.HistoryTurn {
	history := make([]emailagent.HistoryTurn, 0, len(messages))
	for _, message := range messages {
		if message.Sequence <= afterSequence || message.ID == currentMessageID || message.Role == messaging.EmailMessageRoleSystem {
			continue
		}
		role, ok := emailAgentRole(message.Role)
		if !ok || strings.TrimSpace(message.Content) == "" {
			continue
		}
		history = append(history, emailagent.HistoryTurn{Role: role, Text: message.Content, SentAt: message.CreatedAt})
	}
	return history
}

func emailAgentRole(role string) (emailagent.ConversationRole, bool) {
	switch role {
	case messaging.EmailMessageRoleUser:
		return emailagent.RoleUser, true
	case messaging.EmailMessageRoleAssistant:
		return emailagent.RoleAssistant, true
	default:
		return "", false
	}
}

func pendingProposalPreviews(records []messaging.EmailActionProposalRecord) []emailagent.PendingProposal {
	result := make([]emailagent.PendingProposal, 0, len(records))
	for _, record := range records {
		result = append(result, emailagent.PendingProposal{ID: record.ID, Summary: proposalSummary(record)})
	}
	return result
}

func proposalSummary(record messaging.EmailActionProposalRecord) string {
	var proposal emailagent.ActionProposal
	if json.Unmarshal(record.ProposedDiff, &proposal) == nil && strings.TrimSpace(proposal.Summary) != "" {
		return proposal.Summary
	}
	return strings.ReplaceAll(strings.TrimSpace(record.ActionKind), "_", " ")
}

func proposalTarget(proposal emailagent.ActionProposal) (emailagent.TargetSnapshot, error) {
	switch proposal.Kind {
	case emailagent.ActionObjectiveUpdate:
		if proposal.Objective != nil {
			return proposal.Objective.Target, nil
		}
	case emailagent.ActionKeyResultUpdate:
		if proposal.KeyResult != nil {
			return proposal.KeyResult.Target, nil
		}
	case emailagent.ActionStoryUpdate:
		if proposal.Story != nil {
			return proposal.Story.Target, nil
		}
	case emailagent.ActionFeedbackStatus:
		if proposal.Feedback != nil {
			return proposal.Feedback.Target, nil
		}
	}
	return emailagent.TargetSnapshot{}, errors.New("email action proposal has no matching target")
}

func proposalEntityType(kind emailagent.ActionKind) string {
	switch kind {
	case emailagent.ActionObjectiveUpdate:
		return string(emailagent.TargetObjective)
	case emailagent.ActionKeyResultUpdate:
		return string(emailagent.TargetKeyResult)
	case emailagent.ActionStoryUpdate:
		return string(emailagent.TargetStory)
	case emailagent.ActionFeedbackStatus:
		return string(emailagent.TargetFeedback)
	default:
		return "unknown"
	}
}

func deterministicReply(subject string, paragraphs ...string) resolvedReply {
	blocks := make([]emailagent.CopyBlock, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		blocks = append(blocks, emailagent.CopyBlock{Kind: emailagent.CopyBlockParagraph, Text: paragraph})
	}
	copy := emailagent.EmailCopy{
		Subject: replyEmailSubject(subject), Blocks: blocks, PlainText: emailagent.RenderPlainText(blocks),
	}
	return resolvedReply{Copy: copy, Kind: messaging.EmailMessageKindReceipt, Key: "control"}
}

func resumedProposalReply(subject string, proposal messaging.EmailActionProposalRecord, eventID string) resolvedReply {
	blocks := []emailagent.CopyBlock{
		{Kind: emailagent.CopyBlockParagraph, Text: "Here’s the change I’m ready to make:"},
		{Kind: emailagent.CopyBlockCallout, Text: proposalSummary(proposal)},
		{Kind: emailagent.CopyBlockParagraph, Text: "Reply CONFIRM to apply it, or CANCEL to leave everything unchanged."},
	}
	copy := emailagent.EmailCopy{
		Subject: replyEmailSubject(subject), Blocks: blocks, PlainText: emailagent.RenderPlainText(blocks),
	}
	return resolvedReply{Copy: copy, Kind: messaging.EmailMessageKindProposal, Key: "proposal:" + eventID}
}

func deterministicReplyMessageID(threadID uuid.UUID, eventID, key string) string {
	digest := sha256.Sum256([]byte(threadID.String() + ":" + strings.TrimSpace(eventID) + ":" + key))
	return "<maya-email-" + hex.EncodeToString(digest[:16]) + "@fortyone.app>"
}

func boundedProviderMessageID(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		value = "missing"
	}
	if len(value) <= 998 {
		return value
	}
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func emailReplyReferences(thread messaging.EmailThreadRecord) []string {
	values := []string{thread.RootInternetMessageID, thread.LatestInternetMessageID}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (processor *Processor) failUnreadableEvent(ctx context.Context, scope, eventID string, cause error) error {
	receipt, claimed, err := processor.store.StartInboundEvent(ctx, Provider, scope, eventID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := processor.store.CompleteInboundEvent(context.WithoutCancel(ctx), receipt.ID, "failed", truncateProcessorError(cause)); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func terminalInboundStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "completed", "ignored", "cancelled":
		return true
	default:
		return false
	}
}

func truncateProcessorError(err error) string {
	if err == nil {
		return ""
	}
	return truncateRunes(err.Error(), maximumStoredErrorRunes)
}

func truncateRunes(value string, maximum int) string {
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	if maximum <= 1 {
		return "…"
	}
	return string(runes[:maximum-1]) + "…"
}
