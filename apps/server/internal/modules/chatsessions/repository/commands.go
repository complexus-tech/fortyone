package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CreateSessionWithMessages idempotently creates a chat session and its initial
// transcript. Existing transcripts are never replaced by this legacy endpoint;
// subsequent writes must use the reservation protocol.
func (r *repo) CreateSessionWithMessages(ctx context.Context, session *chatsessions.CoreChatSession, messages []any) (chatsessions.CoreChatSession, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.CreateSessionWithMessages")
	defer span.End()

	var stored chatsessionssql.ChatSession
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		var err error
		stored, err = queries.UpsertChatSession(ctx, chatsessionssql.UpsertChatSessionParams{
			SessionID:   session.ID,
			UserID:      session.UserID,
			WorkspaceID: session.WorkspaceID,
			Title:       session.Title,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return chatsessions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("create chat session: %w", err)
		}
		if len(messages) == 0 {
			return nil
		}

		encoded, err := json.Marshal(messages)
		if err != nil {
			return fmt.Errorf("marshal initial chat messages: %w", err)
		}
		if _, err := queries.InsertInitialChatMessages(ctx, chatsessionssql.InsertInitialChatMessagesParams{
			SessionID: stored.ID,
			Messages:  encoded,
		}); err != nil {
			return fmt.Errorf("save initial chat messages: %w", err)
		}
		return nil
	})
	if err != nil {
		return chatsessions.CoreChatSession{}, err
	}

	r.log.Info(ctx, "Chat session persisted successfully", "session_id", stored.ID)
	span.AddEvent("Chat session persisted with messages", trace.WithAttributes(
		attribute.String("session.id", stored.ID),
		attribute.String("session.title", stored.Title),
		attribute.Int("message.count", len(messages)),
	))
	return toCoreChatSession(stored), nil
}

// UpdateSession updates the title of a chat session.
func (r *repo) UpdateSession(ctx context.Context, id string, userID, workspaceID uuid.UUID, title string) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.UpdateSession")
	defer span.End()

	rows, err := r.queries.UpdateChatSessionTitle(ctx, chatsessionssql.UpdateChatSessionTitleParams{
		Title:       title,
		SessionID:   id,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("update chat session: %w", err)
	}
	if rows == 0 {
		return chatsessions.ErrNotFound
	}
	return nil
}

// DeleteSession performs soft delete of the chat session with the specified ID.
func (r *repo) DeleteSession(ctx context.Context, id string, userID, workspaceID uuid.UUID) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.DeleteSession")
	defer span.End()

	r.log.Info(ctx, "Soft deleting chat session", "session_id", id)
	rows, err := r.queries.SoftDeleteChatSession(ctx, chatsessionssql.SoftDeleteChatSessionParams{
		SessionID:   id,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		r.log.Error(ctx, "Failed to soft delete chat session", "session_id", id, "error", err)
		return fmt.Errorf("delete chat session: %w", err)
	}
	if rows == 0 {
		return chatsessions.ErrNotFound
	}

	r.log.Info(ctx, "Chat session soft deleted successfully", "session_id", id)
	span.AddEvent("Chat session soft deleted", trace.WithAttributes(attribute.String("session.id", id)))
	return nil
}

// SaveMessages only initializes a missing legacy transcript. Replacing an
// existing whole array is intentionally rejected because it cannot prove that
// the caller observed the latest generation.
func (r *repo) SaveMessages(ctx context.Context, sessionID string, userID, workspaceID uuid.UUID, messages []any) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.SaveMessages")
	defer span.End()

	encoded, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("marshal chat messages: %w", err)
	}
	rows, err := r.queries.InitializeLegacyChatMessages(ctx, chatsessionssql.InitializeLegacyChatMessagesParams{
		Messages:    encoded,
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("save chat messages: %w", err)
	}
	if rows > 0 {
		return nil
	}

	exists, err := r.queries.ChatSessionExists(ctx, chatsessionssql.ChatSessionExistsParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("check chat session after rejected legacy save: %w", err)
	}
	if !exists {
		return chatsessions.ErrNotFound
	}
	return chatsessions.ErrMessageWriteConflict
}
