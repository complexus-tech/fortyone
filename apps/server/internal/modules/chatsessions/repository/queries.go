package chatsessionsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// GetSession returns the chat session with the specified ID.
func (r *repo) GetSession(ctx context.Context, id string, workspaceID uuid.UUID) (chatsessions.CoreChatSession, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetSession")
	defer span.End()

	q := `
		SELECT id, user_id, workspace_id, title, created_at, updated_at, deleted_at
		FROM chat_sessions
		WHERE id = :id AND workspace_id = :workspace_id AND deleted_at IS NULL
	`

	var cs dbChatSession
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &cs, map[string]any{
		"id":           id,
		"workspace_id": workspaceID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chatsessions.CoreChatSession{}, chatsessions.ErrNotFound
		}
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to get chat session: %w", err)
	}

	return toCoreChatSession(cs), nil
}

// ListSessions returns a list of chat sessions for a user in a workspace.
func (r *repo) ListSessions(ctx context.Context, userID, workspaceID uuid.UUID) ([]chatsessions.CoreChatSession, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.ListSessions")
	defer span.End()

	q := `
		SELECT id, user_id, workspace_id, title, created_at, updated_at, deleted_at
		FROM chat_sessions
		WHERE user_id = :user_id AND workspace_id = :workspace_id AND deleted_at IS NULL
		ORDER BY updated_at DESC LIMIT 25
	`

	var sessions []dbChatSession
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &sessions, map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}); err != nil {
		return nil, fmt.Errorf("failed to list chat sessions: %w", err)
	}

	return toCoreChatSessions(sessions), nil
}

// GetMessages returns the messages for a chat session.
func (r *repo) GetMessages(ctx context.Context, sessionID string) ([]any, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetMessages")
	defer span.End()

	q := `
		SELECT messages
		FROM chat_messages
		WHERE session_id = :session_id
	`

	var messagesJSON json.RawMessage
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &messagesJSON, map[string]any{
		"session_id": sessionID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []any{}, nil // Return empty array if no messages found
		}
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	var messages []any
	if err := json.Unmarshal(messagesJSON, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	return messages, nil
}

// CountUserMessages counts the number of user messages for a user in a given time range.
func (r *repo) CountUserMessages(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID, start, end time.Time) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.CountUserMessages")
	defer span.End()

	q := `
		SELECT count(*)
		FROM chat_sessions s
		JOIN chat_messages m ON s.id = m.session_id
		CROSS JOIN LATERAL jsonb_array_elements(m.messages) AS msg
		WHERE s.user_id = :user_id
		AND s.workspace_id = :workspace_id
		AND s.created_at >= :start_date 
		AND s.created_at < :end_date
		AND msg->>'role' = 'user';
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"start_date":   start,
		"end_date":     end,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		return 0, fmt.Errorf("failed to count user messages: %w", err)
	}

	return count, nil
}
