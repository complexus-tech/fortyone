package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetSession returns the chat session with the specified ID.
func (r *repo) GetSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) (chatsessions.CoreChatSession, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.GetSession")
	defer span.End()

	stored, err := r.queries.GetChatSession(ctx, chatsessionssql.GetChatSessionParams{
		SessionID:   id,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return chatsessions.CoreChatSession{}, chatsessions.ErrNotFound
	}
	if err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("get chat session: %w", err)
	}
	return toCoreChatSession(stored), nil
}

// ListSessions returns a list of chat sessions for a user in a workspace.
func (r *repo) ListSessions(ctx context.Context, userID, workspaceID uuid.UUID) ([]chatsessions.CoreChatSession, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.ListSessions")
	defer span.End()

	stored, err := r.queries.ListChatSessions(ctx, chatsessionssql.ListChatSessionsParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	return toCoreChatSessions(stored), nil
}

// GetMessages returns the messages for a chat session.
func (r *repo) GetMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) ([]any, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.GetMessages")
	defer span.End()

	encoded, err := r.queries.GetChatMessages(ctx, chatsessionssql.GetChatMessagesParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, chatsessions.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}

	var messages []any
	if err := json.Unmarshal(encoded, &messages); err != nil {
		return nil, fmt.Errorf("decode chat messages: %w", err)
	}
	return messages, nil
}

// GetLatestAssistantMessage returns the newest assistant message without transferring the full chat payload.
func (r *repo) GetLatestAssistantMessage(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) (json.RawMessage, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.GetLatestAssistantMessage")
	defer span.End()

	message, err := r.queries.GetLatestAssistantMessage(ctx, chatsessionssql.GetLatestAssistantMessageParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, chatsessions.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get latest assistant message: %w", err)
	}
	if len(message) == 0 {
		return nil, nil
	}
	if !json.Valid(message) {
		return nil, errors.New("latest assistant message is not valid JSON")
	}
	return json.RawMessage(message), nil
}

// CountUserMessages counts the number of user messages for a user in a given time range.
func (r *repo) CountUserMessages(ctx context.Context, userID, workspaceID uuid.UUID, start, end time.Time) (int, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.CountUserMessages")
	defer span.End()

	count, err := r.queries.CountUserMessages(ctx, chatsessionssql.CountUserMessagesParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StartDate:   start,
		EndDate:     end,
	})
	if err != nil {
		return 0, fmt.Errorf("count user chat messages: %w", err)
	}
	converted, err := safecast.Int64(count)
	if err != nil {
		return 0, fmt.Errorf("convert user chat message count: %w", err)
	}
	return converted, nil
}
