package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
)

type ConversationInput struct {
	Provider            string
	WorkspaceID         uuid.UUID
	ExternalWorkspaceID string
	ExternalChannelID   string
	ExternalThreadID    string
	UserID              uuid.UUID
	AudienceScope       string
	AudienceFingerprint string
}

const (
	ConversationAudienceActor   = "actor"
	ConversationAudienceChannel = "channel"
)

type MessageRecord struct {
	ExternalMessageID *string
	Role              string
	Content           string
	CreatedAt         time.Time
}

type ConversationRecord struct {
	ID        uuid.UUID
	UpdatedAt time.Time
}

func (repository *Repository) UpsertConversation(ctx context.Context, input ConversationInput) (uuid.UUID, error) {
	audienceScope := normalizeConversationAudience(input.AudienceScope)
	audienceFingerprint := strings.TrimSpace(input.AudienceFingerprint)
	if audienceScope == ConversationAudienceChannel && audienceFingerprint == "" {
		return uuid.Nil, errors.New("channel conversation audience fingerprint is required")
	}
	if audienceScope == ConversationAudienceActor && audienceFingerprint != "" {
		return uuid.Nil, errors.New("actor conversation cannot have an audience fingerprint")
	}
	if !repository.configured() {
		return uuid.Nil, errors.New("messaging repository is not configured")
	}
	var (
		id  uuid.UUID
		err error
	)
	if audienceScope == ConversationAudienceChannel {
		id, err = repository.queries.UpsertChannelConversation(ctx, messagingsql.UpsertChannelConversationParams{
			Provider: input.Provider, WorkspaceID: input.WorkspaceID,
			ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
			ExternalThreadID: input.ExternalThreadID, UserID: input.UserID,
			AudienceFingerprint: audienceFingerprint,
		})
	} else {
		id, err = repository.queries.UpsertActorConversation(ctx, messagingsql.UpsertActorConversationParams{
			Provider: input.Provider, WorkspaceID: input.WorkspaceID,
			ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
			ExternalThreadID: input.ExternalThreadID, UserID: input.UserID,
		})
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert messaging conversation: %w", err)
	}
	return id, nil
}

func (repository *Repository) FindConversation(ctx context.Context, input ConversationInput) (ConversationRecord, error) {
	if !repository.configured() {
		return ConversationRecord{}, errors.New("messaging repository is not configured")
	}
	audienceScope := normalizeConversationAudience(input.AudienceScope)
	var (
		record ConversationRecord
		err    error
	)
	if audienceScope == ConversationAudienceActor {
		row, queryErr := repository.queries.FindActorConversation(ctx, messagingsql.FindActorConversationParams{
			Provider: input.Provider, WorkspaceID: input.WorkspaceID,
			ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
			ExternalThreadID: input.ExternalThreadID, UserID: input.UserID,
		})
		record, err = ConversationRecord{ID: row.ID, UpdatedAt: row.UpdatedAt}, queryErr
	} else if fingerprint := strings.TrimSpace(input.AudienceFingerprint); fingerprint != "" {
		row, queryErr := repository.queries.FindChannelConversationByFingerprint(ctx, messagingsql.FindChannelConversationByFingerprintParams{
			Provider: input.Provider, WorkspaceID: input.WorkspaceID,
			ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
			ExternalThreadID: input.ExternalThreadID, AudienceFingerprint: fingerprint,
		})
		record, err = ConversationRecord{ID: row.ID, UpdatedAt: row.UpdatedAt}, queryErr
	} else {
		row, queryErr := repository.queries.FindLatestChannelConversation(ctx, messagingsql.FindLatestChannelConversationParams{
			Provider: input.Provider, WorkspaceID: input.WorkspaceID,
			ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
			ExternalThreadID: input.ExternalThreadID,
		})
		record, err = ConversationRecord{ID: row.ID, UpdatedAt: row.UpdatedAt}, queryErr
	}
	if err != nil {
		return ConversationRecord{}, fmt.Errorf("find messaging conversation: %w", err)
	}
	return record, nil
}

func (repository *Repository) FindChannelConversation(ctx context.Context, input ConversationInput) (ConversationRecord, error) {
	input.AudienceScope = ConversationAudienceChannel
	return repository.FindConversation(ctx, input)
}

func normalizeConversationAudience(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), ConversationAudienceChannel) {
		return ConversationAudienceChannel
	}
	return ConversationAudienceActor
}

func (repository *Repository) AppendMessage(
	ctx context.Context,
	conversationID uuid.UUID,
	externalMessageID, role, content string,
) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	if err := repository.queries.AppendConversationMessage(ctx, messagingsql.AppendConversationMessageParams{
		ConversationID: conversationID, ExternalMessageID: externalMessageID, Role: role, Content: content,
	}); err != nil {
		return fmt.Errorf("append messaging message: %w", err)
	}
	return nil
}

func (repository *Repository) ListRecentMessages(
	ctx context.Context,
	conversationID uuid.UUID,
	limit int,
) ([]MessageRecord, error) {
	if !repository.configured() {
		return nil, errors.New("messaging repository is not configured")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := repository.queries.ListRecentConversationMessages(ctx, messagingsql.ListRecentConversationMessagesParams{
		ConversationID: conversationID, RowLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list messaging messages: %w", err)
	}
	records := make([]MessageRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, MessageRecord{
			ExternalMessageID: row.ExternalMessageID, Role: row.Role,
			Content: row.Content, CreatedAt: row.CreatedAt,
		})
	}
	return records, nil
}
