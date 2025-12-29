package chatsessions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotFound = errors.New("chat session not found")
)

// Repository provides access to the chat session storage.
type Repository interface {
	CreateSessionWithMessages(ctx context.Context, session *CoreChatSession, messages []any) (CoreChatSession, error)
	GetSession(ctx context.Context, id string, workspaceID uuid.UUID) (CoreChatSession, error)
	ListSessions(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreChatSession, error)
	UpdateSession(ctx context.Context, id string, workspaceID uuid.UUID, title string) error
	DeleteSession(ctx context.Context, id string, workspaceID uuid.UUID) error
	SaveMessages(ctx context.Context, sessionID string, messages []any) error
	GetMessages(ctx context.Context, sessionID string) ([]any, error)
	CountUserMessages(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)
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
func (s *Service) GetSession(ctx context.Context, id string, workspaceID uuid.UUID) (CoreChatSession, error) {
	s.log.Info(ctx, "business.core.chatsessions.GetSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.GetSession")
	defer span.End()

	session, err := s.repo.GetSession(ctx, id, workspaceID)
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
func (s *Service) UpdateSession(ctx context.Context, id string, workspaceID uuid.UUID, title string) error {
	s.log.Info(ctx, "business.core.chatsessions.UpdateSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.UpdateSession")
	defer span.End()

	if err := s.repo.UpdateSession(ctx, id, workspaceID, title); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// DeleteSession deletes the chat session with the specified ID.
func (s *Service) DeleteSession(ctx context.Context, id string, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.chatsessions.DeleteSession")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.DeleteSession")
	defer span.End()

	if err := s.repo.DeleteSession(ctx, id, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// SaveMessages saves messages for a chat session.
func (s *Service) SaveMessages(ctx context.Context, sessionID string, messages []any) error {
	s.log.Info(ctx, "business.core.chatsessions.SaveMessages")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.SaveMessages")
	defer span.End()

	if err := s.repo.SaveMessages(ctx, sessionID, messages); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("messages saved", trace.WithAttributes(
		attribute.String("session.id", sessionID),
		attribute.Int("message.count", len(messages)),
	))
	return nil
}

// GetMessages returns the messages for a chat session.
func (s *Service) GetMessages(ctx context.Context, sessionID string) ([]any, error) {
	s.log.Info(ctx, "business.core.chatsessions.GetMessages")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.GetMessages")
	defer span.End()

	messages, err := s.repo.GetMessages(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return messages, nil
}

// CountUserMessagesCurrentMonth returns the number of messages sent by the user in the current month.
func (s *Service) CountUserMessagesCurrentMonth(ctx context.Context, userID uuid.UUID) (int, error) {
	s.log.Info(ctx, "business.core.chatsessions.CountUserMessagesCurrentMonth")
	ctx, span := web.AddSpan(ctx, "business.core.chatsessions.CountUserMessagesCurrentMonth")
	defer span.End()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now.Add(time.Duration(24 * time.Hour))

	count, err := s.repo.CountUserMessages(ctx, userID, start, end)
	if err != nil {
		return 0, fmt.Errorf("counting user messages: %w", err)
	}

	return count, nil
}
