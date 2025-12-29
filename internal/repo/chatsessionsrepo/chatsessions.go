package chatsessionsrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/chatsessions"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// CreateSessionWithMessages creates a new chat session with initial messages in a transaction.
func (r *repo) CreateSessionWithMessages(ctx context.Context, session *chatsessions.CoreChatSession, messages []any) (chatsessions.CoreChatSession, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.CreateSessionWithMessages")
	defer span.End()

	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create session
	sessionQuery := `
		INSERT INTO chat_sessions (id, user_id, workspace_id, title)
		VALUES (:id, :user_id, :workspace_id, :title)
		RETURNING id, user_id, workspace_id, title, created_at, updated_at
	`

	var cs dbChatSession
	stmt, err := tx.PrepareNamedContext(ctx, sessionQuery)
	if err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to prepare session statement: %w", err)
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &cs, map[string]any{
		"id":           session.ID,
		"user_id":      session.UserID,
		"workspace_id": session.WorkspaceID,
		"title":        session.Title,
	}); err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to create chat session: %w", err)
	}

	// Save messages if provided
	if len(messages) > 0 {
		messagesJSON, err := json.Marshal(messages)
		if err != nil {
			return chatsessions.CoreChatSession{}, fmt.Errorf("failed to marshal messages: %w", err)
		}

		messagesQuery := `
			INSERT INTO chat_messages (session_id, messages)
			VALUES (:session_id, :messages)
		`

		messagesStmt, err := tx.PrepareNamedContext(ctx, messagesQuery)
		if err != nil {
			return chatsessions.CoreChatSession{}, fmt.Errorf("failed to prepare messages statement: %w", err)
		}
		defer messagesStmt.Close()

		if _, err := messagesStmt.ExecContext(ctx, map[string]any{
			"session_id": cs.ID,
			"messages":   messagesJSON,
		}); err != nil {
			return chatsessions.CoreChatSession{}, fmt.Errorf("failed to save messages: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return chatsessions.CoreChatSession{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, "Chat session created successfully", "session_id", cs.ID)
	span.AddEvent("Chat session created with messages", trace.WithAttributes(
		attribute.String("session.id", cs.ID),
		attribute.String("session.title", cs.Title),
		attribute.Int("message.count", len(messages)),
	))

	return toCoreChatSession(cs), nil
}

// GetSession returns the chat session with the specified ID.
func (r *repo) GetSession(ctx context.Context, id string, workspaceID uuid.UUID) (chatsessions.CoreChatSession, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.GetSession")
	defer span.End()

	q := `
		SELECT id, user_id, workspace_id, title, created_at, updated_at
		FROM chat_sessions
		WHERE id = :id AND workspace_id = :workspace_id
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
		SELECT id, user_id, workspace_id, title, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = :user_id AND workspace_id = :workspace_id
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

// UpdateSession updates the title of a chat session.
func (r *repo) UpdateSession(ctx context.Context, id string, workspaceID uuid.UUID, title string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.UpdateSession")
	defer span.End()

	q := `
		UPDATE chat_sessions
		SET title = :title, updated_at = NOW()
		WHERE id = :id AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, map[string]any{
		"id":           id,
		"workspace_id": workspaceID,
		"title":        title,
	})
	if err != nil {
		return fmt.Errorf("failed to update chat session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return chatsessions.ErrNotFound
	}

	return nil
}

// DeleteSession deletes the chat session with the specified ID.
func (r *repo) DeleteSession(ctx context.Context, id string, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.DeleteSession")
	defer span.End()

	q := `
		DELETE FROM chat_sessions
		WHERE id = :id AND workspace_id = :workspace_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, map[string]any{
		"id":           id,
		"workspace_id": workspaceID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return chatsessions.ErrNotFound
	}

	return nil
}

// SaveMessages saves messages for a chat session.
func (r *repo) SaveMessages(ctx context.Context, sessionID string, messages []any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.SaveMessages")
	defer span.End()

	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	q := `
		INSERT INTO chat_messages (session_id, messages)
		VALUES (:session_id, :messages)
		ON CONFLICT (session_id) 
		DO UPDATE SET 
			messages = :messages,
			updated_at = NOW()
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, map[string]any{
		"session_id": sessionID,
		"messages":   messagesJSON,
	}); err != nil {
		return fmt.Errorf("failed to save messages: %w", err)
	}

	return nil
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
