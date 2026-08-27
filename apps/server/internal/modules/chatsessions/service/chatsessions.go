package chatsessions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotFound                  = errors.New("chat session not found")
	ErrMutationApprovalConflict  = errors.New("mutation approval does not match the prepared tool call")
	ErrMutationApprovalUncertain = errors.New("an identical mutation has an unresolved execution")
	ErrMutationApprovalLease     = errors.New("mutation approval execution lease is invalid or expired")
	ErrMessageWriteConflict      = errors.New("chat transcript changed before this write could be applied")
	ErrMessageWriteApprovalOpen  = errors.New("chat transcript has an unresolved mutation approval")
	ErrMessageWriteInvalid       = errors.New("chat transcript write is invalid")
)

const (
	MutationApprovalFailureStartUncertain      = "start_transition_uncertain"
	MutationApprovalFailureCompletionUncertain = "completion_persistence_uncertain"
)

// Repository provides access to the chat session storage.
type Repository interface {
	CreateSessionWithMessages(ctx context.Context, session *CoreChatSession, messages []any) (CoreChatSession, error)
	GetSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) (CoreChatSession, error)
	ListSessions(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreChatSession, error)
	UpdateSession(ctx context.Context, id string, userID, workspaceID uuid.UUID, title string) error
	DeleteSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) error
	SaveMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID, messages []any) error
	BeginMessageWrite(ctx context.Context, params BeginMessageWriteParams) (CoreMessageWriteReservation, error)
	FinalizeMessageWrite(ctx context.Context, params FinalizeMessageWriteParams) (CoreMessageWriteResult, error)
	RecoverMutationApprovalOutput(ctx context.Context, params RecoverMutationApprovalOutputParams) (CoreMessageWriteResult, error)
	GetMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) ([]any, error)
	GetLatestAssistantMessage(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) (json.RawMessage, error)
	ClaimMutationApproval(ctx context.Context, params MutationApprovalExecutionParams) (CoreMutationApprovalExecution, error)
	StartMutationApproval(ctx context.Context, params MutationApprovalExecutionParams) (CoreMutationApprovalExecution, error)
	CompleteMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, output json.RawMessage) (CoreMutationApprovalExecution, error)
	FailMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, failureCode string) (CoreMutationApprovalExecution, error)
	ReconcileMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, reconciliation MutationApprovalReconciliation) (CoreMutationApprovalExecution, error)
	CountUserMessages(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, start, end time.Time) (int, error)
}

// BeginMessageWrite reserves the only generation that may finalize a model or
// approval response. The repository validates and stores the request-side
// transcript in the same transaction as the reservation.
func (s *Service) BeginMessageWrite(ctx context.Context, params BeginMessageWriteParams) (CoreMessageWriteReservation, error) {
	s.log.Info(ctx, "business.core.chatsessions.BeginMessageWrite")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.BeginMessageWrite")
	defer span.End()

	reservation, err := s.repo.BeginMessageWrite(ctx, params)
	if err != nil {
		span.RecordError(err)
		return CoreMessageWriteReservation{}, err
	}
	span.AddEvent("message write reserved", trace.WithAttributes(
		attribute.String("session.id", params.Session.ID),
		attribute.String("write.operation", string(params.Operation)),
		attribute.Int64("write.generation", reservation.Generation),
	))
	return reservation, nil
}

// FinalizeMessageWrite applies a response only when its opaque token still
// owns the active generation. Superseded responses are successful no-ops.
func (s *Service) FinalizeMessageWrite(ctx context.Context, params FinalizeMessageWriteParams) (CoreMessageWriteResult, error) {
	s.log.Info(ctx, "business.core.chatsessions.FinalizeMessageWrite")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.FinalizeMessageWrite")
	defer span.End()

	if params.Generation <= 0 || params.Token == uuid.Nil {
		return CoreMessageWriteResult{}, ErrMessageWriteInvalid
	}
	result, err := s.repo.FinalizeMessageWrite(ctx, params)
	if err != nil {
		span.RecordError(err)
		return CoreMessageWriteResult{}, err
	}
	span.SetAttributes(attribute.Bool("write.applied", result.Applied))
	return result, nil
}

// RecoverMutationApprovalOutput projects an already durable ledger result onto
// the exact surviving tool call without replacing any surrounding history.
func (s *Service) RecoverMutationApprovalOutput(ctx context.Context, params RecoverMutationApprovalOutputParams) (CoreMessageWriteResult, error) {
	s.log.Info(ctx, "business.core.chatsessions.RecoverMutationApprovalOutput")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.RecoverMutationApprovalOutput")
	defer span.End()

	result, err := s.repo.RecoverMutationApprovalOutput(ctx, params)
	if err != nil {
		span.RecordError(err)
		return CoreMessageWriteResult{}, err
	}
	span.SetAttributes(attribute.Bool("write.applied", result.Applied))
	return result, nil
}

// Service provides chat session-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new chat sessions service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// CreateSession creates a new chat session with initial messages.
func (s *Service) CreateSession(ctx context.Context, ncs CoreNewChatSession) (CoreChatSession, error) {
	s.log.Info(ctx, "business.core.chatsessions.create")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.Create")
	defer span.End()
	if err := validateLegacyMessageInitialization(ncs.Messages); err != nil {
		span.RecordError(err)
		return CoreChatSession{}, err
	}

	session := CoreChatSession{
		ID:          ncs.ID,
		UserID:      ncs.UserID,
		WorkspaceID: ncs.WorkspaceID,
		Title:       ncs.Title,
	}

	cs, err := s.repo.CreateSessionWithMessages(ctx, &session, ncs.Messages)
	if err != nil {
		span.RecordError(err)
		return CoreChatSession{}, err
	}

	span.AddEvent("chat session created with messages", trace.WithAttributes(
		attribute.String("session.id", cs.ID),
		attribute.String("session.title", cs.Title),
		attribute.Int("message.count", len(ncs.Messages)),
	))
	return cs, nil
}

// GetSession returns the chat session with the specified ID.
func (s *Service) GetSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) (CoreChatSession, error) {
	s.log.Info(ctx, "business.core.chatsessions.GetSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.GetSession")
	defer span.End()

	session, err := s.repo.GetSession(ctx, id, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreChatSession{}, err
	}

	return session, nil
}

// ListSessions returns a list of chat sessions for a user in a workspace.
func (s *Service) ListSessions(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreChatSession, error) {
	s.log.Info(ctx, "business.core.chatsessions.ListSessions")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.ListSessions")
	defer span.End()

	sessions, err := s.repo.ListSessions(ctx, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("chat sessions retrieved", trace.WithAttributes(
		attribute.Int("session.count", len(sessions)),
	))
	return sessions, nil
}

// UpdateSession updates the title of a chat session.
func (s *Service) UpdateSession(ctx context.Context, id string, userID, workspaceID uuid.UUID, title string) error {
	s.log.Info(ctx, "business.core.chatsessions.UpdateSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.UpdateSession")
	defer span.End()

	if err := s.repo.UpdateSession(ctx, id, userID, workspaceID, title); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// DeleteSession deletes the chat session with the specified ID.
func (s *Service) DeleteSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.chatsessions.DeleteSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.DeleteSession")
	defer span.End()

	if err := s.repo.DeleteSession(ctx, id, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// SaveMessages saves messages for a chat session.
func (s *Service) SaveMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID, messages []any) error {
	s.log.Info(ctx, "business.core.chatsessions.SaveMessages")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.SaveMessages")
	defer span.End()
	if err := validateLegacyMessageInitialization(messages); err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.repo.SaveMessages(ctx, sessionID, userID, workspaceID, messages); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("messages saved", trace.WithAttributes(
		attribute.String("session.id", sessionID),
		attribute.Int("message.count", len(messages)),
	))
	return nil
}

// validateLegacyMessageInitialization keeps rolling clients able to seed safe
// text/file history without letting the legacy endpoints manufacture an
// already-approved mutation outside the reservation protocol.
func validateLegacyMessageInitialization(messages []any) error {
	for _, rawMessage := range messages {
		message, ok := rawMessage.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: legacy message must be an object", ErrMessageWriteInvalid)
		}
		parts, ok := message["parts"].([]any)
		if !ok {
			return fmt.Errorf("%w: legacy message parts are invalid", ErrMessageWriteInvalid)
		}
		for _, rawPart := range parts {
			part, ok := rawPart.(map[string]any)
			if !ok {
				return fmt.Errorf("%w: legacy message part must be an object", ErrMessageWriteInvalid)
			}
			partType, _ := part["type"].(string)
			_, hasToolCallID := part["toolCallId"]
			_, hasApproval := part["approval"]
			if strings.HasPrefix(partType, "tool-") ||
				partType == "dynamic-tool" ||
				partType == "tool-invocation" ||
				hasToolCallID ||
				hasApproval {
				return fmt.Errorf("%w: legacy initialization cannot contain tool state", ErrMessageWriteInvalid)
			}
		}
	}
	return nil
}

// GetMessages returns the messages for a chat session.
func (s *Service) GetMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) ([]any, error) {
	s.log.Info(ctx, "business.core.chatsessions.GetMessages")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.GetMessages")
	defer span.End()

	messages, err := s.repo.GetMessages(ctx, sessionID, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return messages, nil
}

// GetLatestAssistantMessage returns only the newest persisted assistant message for approval validation.
func (s *Service) GetLatestAssistantMessage(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) (json.RawMessage, error) {
	s.log.Info(ctx, "business.core.chatsessions.GetLatestAssistantMessage")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.GetLatestAssistantMessage")
	defer span.End()

	message, err := s.repo.GetLatestAssistantMessage(ctx, sessionID, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return message, nil
}

// ClaimMutationApproval atomically claims a prepared mutation or returns its durable state.
func (s *Service) ClaimMutationApproval(ctx context.Context, params MutationApprovalExecutionParams) (CoreMutationApprovalExecution, error) {
	s.log.Info(ctx, "business.core.chatsessions.ClaimMutationApproval")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.ClaimMutationApproval")
	defer span.End()

	execution, err := s.repo.ClaimMutationApproval(ctx, params)
	if err != nil {
		span.RecordError(err)
		return CoreMutationApprovalExecution{}, err
	}

	return execution, nil
}

// StartMutationApproval crosses the durable no-retry boundary immediately
// before invoking the approved tool.
func (s *Service) StartMutationApproval(ctx context.Context, params MutationApprovalExecutionParams) (CoreMutationApprovalExecution, error) {
	s.log.Info(ctx, "business.core.chatsessions.StartMutationApproval")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.StartMutationApproval")
	defer span.End()

	if params.LeaseToken == uuid.Nil {
		return CoreMutationApprovalExecution{}, ErrMutationApprovalLease
	}
	execution, err := s.repo.StartMutationApproval(ctx, params)
	if err != nil {
		span.RecordError(err)
		return CoreMutationApprovalExecution{}, err
	}

	return execution, nil
}

// CompleteMutationApproval durably records the first result produced by a claimed mutation.
func (s *Service) CompleteMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, output json.RawMessage) (CoreMutationApprovalExecution, error) {
	s.log.Info(ctx, "business.core.chatsessions.CompleteMutationApproval")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.CompleteMutationApproval")
	defer span.End()

	if !json.Valid(output) {
		return CoreMutationApprovalExecution{}, errors.New("mutation approval output is not valid JSON")
	}
	if params.LeaseToken == uuid.Nil {
		return CoreMutationApprovalExecution{}, ErrMutationApprovalLease
	}

	execution, err := s.repo.CompleteMutationApproval(ctx, params, output)
	if err != nil {
		span.RecordError(err)
		return CoreMutationApprovalExecution{}, err
	}

	return execution, nil
}

// FailMutationApproval terminally records an ambiguous transition or
// completion. It never makes an approval retryable.
func (s *Service) FailMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, failureCode string) (CoreMutationApprovalExecution, error) {
	s.log.Info(ctx, "business.core.chatsessions.FailMutationApproval")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.FailMutationApproval")
	defer span.End()

	if params.LeaseToken == uuid.Nil {
		return CoreMutationApprovalExecution{}, ErrMutationApprovalLease
	}
	switch failureCode {
	case MutationApprovalFailureStartUncertain,
		MutationApprovalFailureCompletionUncertain:
	default:
		return CoreMutationApprovalExecution{}, errors.New("unsupported mutation approval failure code")
	}

	execution, err := s.repo.FailMutationApproval(ctx, params, failureCode)
	if err != nil {
		span.RecordError(err)
		return CoreMutationApprovalExecution{}, err
	}

	return execution, nil
}

// ReconcileMutationApproval explicitly resolves a terminally uncertain
// execution from independently verified evidence. It never retries a mutation
// by itself: verified-not-applied rows become claimable only through a later
// approval request.
func (s *Service) ReconcileMutationApproval(ctx context.Context, params MutationApprovalExecutionParams, reconciliation MutationApprovalReconciliation) (CoreMutationApprovalExecution, error) {
	s.log.Info(ctx, "business.core.chatsessions.ReconcileMutationApproval")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.ReconcileMutationApproval")
	defer span.End()

	if err := validateMutationApprovalReconciliation(reconciliation); err != nil {
		return CoreMutationApprovalExecution{}, err
	}
	execution, err := s.repo.ReconcileMutationApproval(ctx, params, reconciliation)
	if err != nil {
		span.RecordError(err)
		return CoreMutationApprovalExecution{}, err
	}

	return execution, nil
}

func validateMutationApprovalReconciliation(reconciliation MutationApprovalReconciliation) error {
	reference := strings.TrimSpace(reconciliation.Evidence.Reference)
	summary := strings.TrimSpace(reconciliation.Evidence.Summary)
	if len(reference) == 0 || len(reference) > 255 {
		return errors.New("mutation approval reconciliation evidence reference is invalid")
	}
	if len(summary) == 0 || len(summary) > 500 {
		return errors.New("mutation approval reconciliation evidence summary is invalid")
	}
	switch reconciliation.Evidence.Kind {
	case MutationApprovalReconciliationEvidenceIdempotencyLookup,
		MutationApprovalReconciliationEvidenceWorkspaceProbe:
	default:
		return errors.New("mutation approval reconciliation evidence kind is unsupported")
	}

	switch reconciliation.Resolution {
	case MutationApprovalReconciliationVerifiedCompleted:
		if !json.Valid(reconciliation.Output) {
			return errors.New("verified completion requires a valid JSON output")
		}
	case MutationApprovalReconciliationVerifiedNotApplied:
		if len(reconciliation.Output) > 0 {
			return errors.New("verified-not-applied reconciliation cannot include output")
		}
	default:
		return errors.New("mutation approval reconciliation resolution is unsupported")
	}

	return nil
}

// CountUserMessagesCurrentMonth returns the number of messages sent by the user in the current month.
func (s *Service) CountUserMessagesCurrentMonth(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) (int, error) {
	s.log.Info(ctx, "business.core.chatsessions.CountUserMessagesCurrentMonth")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.CountUserMessagesCurrentMonth")
	defer span.End()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now.Add(time.Duration(24 * time.Hour))

	count, err := s.repo.CountUserMessages(ctx, userID, workspaceID, start, end)
	if err != nil {
		return 0, fmt.Errorf("counting user messages: %w", err)
	}

	return count, nil
}
