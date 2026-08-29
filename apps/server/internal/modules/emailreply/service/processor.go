package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

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

type inboundPayloadOpener interface {
	OpenStoredInboundEmail(sealed string) (StoredInboundEmail, error)
	SealProcessorState(payload []byte) (string, error)
	OpenProcessorState(sealed string) ([]byte, error)
}

// AuthorizedContext is rebuilt from current database state for every inbound
// turn. It contains no target the actor cannot currently access.
type AuthorizedContext struct {
	AllowedTeamIDs []uuid.UUID
	Facts          []GroundedFact
	Targets        []AuthorizedTarget
	Choices        []AuthorizedChoice
}

// ContextLoader reauthorizes the actor and reloads the original guidance
// targets, their current versions, and safe choices.
type ContextLoader interface {
	Load(ctx context.Context, thread Thread) (AuthorizedContext, error)
	AuthorizeProposal(ctx context.Context, proposal ActionProposal) error
	CurrentVersion(ctx context.Context, proposal ActionProposal) (time.Time, error)
	ProposalAlreadyApplied(ctx context.Context, proposal ActionProposal) (bool, error)
}

// MutationApplier invokes domain services only after confirmation-time
// authorization. Implementations must use compare-and-swap domain methods.
type MutationApplier interface {
	Apply(ctx context.Context, proposal ActionProposal) error
}

// ProcessorConfig composes the production email reply worker without exposing
// provider payloads to the task queue.
type ProcessorConfig struct {
	Log        *logger.Logger
	Store      ProcessorStore
	Inbound    inboundPayloadOpener
	Agent      DecisionPort
	Summarizer SummaryPort
	Renderer   CopyRenderer
	Threads    ReplyThreadPort
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
	agent      DecisionPort
	summarizer SummaryPort
	renderer   CopyRenderer
	threads    ReplyThreadPort
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
	case config.Renderer == nil:
		return nil, errors.New("email reply copy renderer is required")
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
		renderer:   config.Renderer,
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
		return errors.New("brevo email reply scope and event id are required")
	}

	inbox, err := processor.store.GetInboundEvent(ctx, Provider, externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	if terminalInboundStatus(inbox.Status) {
		return nil
	}
	if inbox.PayloadEncrypted == nil || strings.TrimSpace(*inbox.PayloadEncrypted) == "" {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("brevo email reply inbox event has no encrypted payload"))
	}
	stored, err := processor.inbound.OpenStoredInboundEmail(*inbox.PayloadEncrypted)
	if err != nil {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, err)
	}
	if stored.ThreadID == uuid.Nil || stored.WorkspaceID == uuid.Nil || stored.UserID == uuid.Nil ||
		inbox.WorkspaceID == nil || *inbox.WorkspaceID != stored.WorkspaceID ||
		externalWorkspaceID != stored.WorkspaceID.String()+":"+stored.ThreadID.String() {
		return processor.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("brevo email reply inbox binding does not match its encrypted payload"))
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
		stateCtx, cancel := processorStateContext(ctx)
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

	thread, err := processor.store.GetThread(ctx, ThreadKey{
		ThreadID: stored.ThreadID, WorkspaceID: stored.WorkspaceID, UserID: stored.UserID,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidConversation) {
			status = "ignored"
			return nil
		}
		return err
	}
	releaseThread, err := processor.store.AcquireThreadLease(ctx, ThreadLease{
		ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
	})
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := releaseThread(); releaseErr != nil {
			if processor.log != nil {
				logCtx, cancel := processorStateContext(ctx)
				processor.log.Error(logCtx, "failed releasing Maya email thread lease", "error", releaseErr, "thread_id", thread.ID)
				cancel()
			}
			if err == nil {
				err = releaseErr
			}
		}
	}()
	// Authorization and history may have changed while another event held the
	// thread lease, so reload after acquiring it.
	thread, err = processor.store.GetThread(ctx, ThreadKey{
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
	inboundMessage, _, err := processor.store.AppendMessage(ctx, MessageInput{
		ThreadID:           thread.ID,
		WorkspaceID:        thread.WorkspaceID,
		UserID:             thread.UserID,
		InboundEventID:     &receipt.ID,
		IdempotencyKey:     "inbound:" + eventID,
		Direction:          MessageDirectionInbound,
		Role:               MessageRoleUser,
		Kind:               MessageKindReply,
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
			failure = errors.New("maya email reply processing ended before delivery")
		}
		_ = processor.failOutboundDelivery(ctx, delivery.ID, failure)
	}()
	if len(delivery.ProviderPayload) > 0 {
		currentAuthorization, loadErr := processor.context.Load(ctx, thread)
		if loadErr != nil {
			if errors.Is(loadErr, ErrActionUnauthorized) || errors.Is(loadErr, ErrInvalidConversation) {
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
	pending, err := processor.store.ListPendingProposals(ctx, ProposalListInput{
		ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID, Now: processor.now().UTC(),
	})
	if err != nil {
		return err
	}
	if control, exact := parseControlCommand(currentReply); exact && len(pending) == 0 {
		status := ProposalConfirmed
		if control == ControlCancel {
			status = ProposalCancelled
		}
		recovered, found, findErr := processor.store.FindLatestProposalForControl(ctx, ProposalControlLookup{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID, Control: status,
		})
		if findErr != nil {
			return findErr
		}
		if found {
			pending = []Proposal{recovered}
		}
	}
	if _, exactControl := parseControlCommand(currentReply); !exactControl && len(pending) == 1 &&
		pending[0].Status == ProposalPending && pending[0].SourceMessageID == inboundMessage.ID {
		resumeReply := resumedProposalReply(stored.Email.Subject, pending[0], eventID)
		var persistedProposal ActionProposal
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
			if _, _, supersedeErr := processor.store.DecideProposal(ctx, ProposalDecision{
				ProposalID: pending[0].ID, ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
				ReplyTokenHash: replyTokenHash, Decision: ProposalSuperseded, Now: processor.now().UTC(),
			}); supersedeErr != nil && !errors.Is(supersedeErr, ErrProposalConflict) {
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

	decision, err := processor.agent.Decide(ctx, AgentRequest{
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
