package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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

// DeleteSession performs soft delete of the chat session with the specified ID.
func (r *repo) DeleteSession(ctx context.Context, id string, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.DeleteSession")
	defer span.End()

	q := `
		UPDATE chat_sessions
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = :id AND workspace_id = :workspace_id AND deleted_at IS NULL
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Soft deleting chat session #%s", id), "id", id)
	result, err := stmt.ExecContext(ctx, map[string]any{
		"id":           id,
		"workspace_id": workspaceID,
	})
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to soft delete chat session: %s", err), "id", id)
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return chatsessions.ErrNotFound
	}

	r.log.Info(ctx, fmt.Sprintf("Chat session #%s soft deleted successfully", id), "id", id)
	span.AddEvent("Chat session soft deleted.", trace.WithAttributes(attribute.String("session.id", id)))

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
