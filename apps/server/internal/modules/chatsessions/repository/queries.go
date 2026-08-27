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
func (r *repo) GetSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) (chatsessions.CoreChatSession, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetSession")
	defer span.End()

	q := `
		SELECT id, user_id, workspace_id, title, created_at, updated_at, deleted_at
		FROM chat_sessions
		WHERE id = :id AND user_id = :user_id AND workspace_id = :workspace_id AND deleted_at IS NULL
	`

	var cs dbChatSession
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &cs, map[string]any{
		"id":           id,
		"user_id":      userID,
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
func (r *repo) GetMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) ([]any, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetMessages")
	defer span.End()

	q := `
		SELECT COALESCE(messages.messages, CAST('[]' AS jsonb))
		FROM chat_sessions sessions
		LEFT JOIN chat_messages messages ON messages.session_id = sessions.id
		WHERE sessions.id = :session_id
			AND sessions.user_id = :user_id
			AND sessions.workspace_id = :workspace_id
			AND sessions.deleted_at IS NULL
	`

	var messagesJSON json.RawMessage
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &messagesJSON, map[string]any{
		"session_id":   sessionID,
		"user_id":      userID,
		"workspace_id": workspaceID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, chatsessions.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	var messages []any
	if err := json.Unmarshal(messagesJSON, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	return messages, nil
}

// GetLatestAssistantMessage returns the newest assistant message without transferring the full chat payload.
func (r *repo) GetLatestAssistantMessage(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID) (json.RawMessage, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetLatestAssistantMessage")
	defer span.End()

	const q = `
		SELECT (
			SELECT messages.messages -> message_index
			FROM generate_series(
				jsonb_array_length(COALESCE(messages.messages, CAST('[]' AS jsonb))) - 1,
				0,
				-1
			) AS indexes(message_index)
			WHERE (messages.messages -> message_index) ->> 'role' = 'assistant'
			LIMIT 1
		)
		FROM chat_sessions AS sessions
		LEFT JOIN chat_messages AS messages ON messages.session_id = sessions.id
		WHERE sessions.id = :session_id
			AND sessions.user_id = :user_id
			AND sessions.workspace_id = :workspace_id
			AND sessions.deleted_at IS NULL
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("prepare latest assistant message query: %w", err)
	}
	defer stmt.Close()

	var message sql.NullString
	if err := stmt.GetContext(ctx, &message, map[string]any{
		"session_id":   sessionID,
		"user_id":      userID,
		"workspace_id": workspaceID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, chatsessions.ErrNotFound
		}
		return nil, fmt.Errorf("get latest assistant message: %w", err)
	}
	if !message.Valid {
		return nil, nil
	}
	messageJSON := json.RawMessage(message.String)
	if !json.Valid(messageJSON) {
		return nil, errors.New("latest assistant message is not valid JSON")
	}

	return messageJSON, nil
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
