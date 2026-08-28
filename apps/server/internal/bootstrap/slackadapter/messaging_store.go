package slackadapter

import (
	"context"
	"errors"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
)

type MessagingBackend interface {
	CreateNonce(context.Context, messagingrepository.NonceInput) error
	ConsumeNonce(context.Context, messagingrepository.NonceConsumeInput) (messagingrepository.NonceRecord, error)
	UpsertConversation(context.Context, messagingrepository.ConversationInput) (uuid.UUID, error)
	FindConversation(context.Context, messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error)
	FindChannelConversation(context.Context, messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error)
	AppendMessage(context.Context, uuid.UUID, string, string, string) error
	ListRecentMessages(context.Context, uuid.UUID, int) ([]messagingrepository.MessageRecord, error)
	StartOutboundDelivery(context.Context, messagingrepository.OutboundDeliveryInput) (messagingrepository.OutboundDeliveryRecord, bool, error)
	SetOutboundDeliveryContent(context.Context, uuid.UUID, string) error
	SetOutboundDeliveryContentAndDestination(context.Context, uuid.UUID, string, string, string) error
	SetOutboundDeliveryContentAndProviderPayload(context.Context, uuid.UUID, string, []byte) error
	CompleteOutboundDelivery(context.Context, uuid.UUID, string) error
	FailOutboundDelivery(context.Context, uuid.UUID, string) error
	CancelOutboundDelivery(context.Context, uuid.UUID, string) error
	ListRecoverableOutboundDeliveries(context.Context, string, int) ([]messagingrepository.OutboundDeliveryRecord, error)
}

type MessagingStore struct {
	backend MessagingBackend
}

func NewMessagingStore(backend MessagingBackend) *MessagingStore {
	if backend == nil {
		return nil
	}
	return &MessagingStore{backend: backend}
}

func (adapter *MessagingStore) CreateNonce(ctx context.Context, input slack.NonceInput) error {
	return mapMessagingError(adapter.backend.CreateNonce(ctx, messagingrepository.NonceInput{
		Provider: input.Provider, Purpose: input.Purpose, NonceHash: append([]byte(nil), input.NonceHash...),
		WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalUserID: input.ExternalUserID,
		Payload: append([]byte(nil), input.Payload...), ExpiresAt: input.ExpiresAt,
	}))
}

func (adapter *MessagingStore) ConsumeNonce(ctx context.Context, input slack.NonceConsumeInput) (slack.NonceRecord, error) {
	record, err := adapter.backend.ConsumeNonce(ctx, messagingrepository.NonceConsumeInput{
		Provider: input.Provider, Purpose: input.Purpose, NonceHash: append([]byte(nil), input.NonceHash...),
		WorkspaceID: input.WorkspaceID, UserID: input.UserID, Now: input.Now,
	})
	return slack.NonceRecord{
		ID: record.ID, Provider: record.Provider, Purpose: record.Purpose,
		WorkspaceID: record.WorkspaceID, UserID: record.UserID,
		ExternalWorkspaceID: record.ExternalWorkspaceID, ExternalUserID: record.ExternalUserID,
		Payload: append([]byte(nil), record.Payload...), ExpiresAt: record.ExpiresAt, ConsumedAt: record.ConsumedAt,
	}, mapMessagingError(err)
}

func (adapter *MessagingStore) UpsertConversation(ctx context.Context, input slack.ConversationInput) (uuid.UUID, error) {
	id, err := adapter.backend.UpsertConversation(ctx, mapConversationInput(input))
	return id, mapMessagingError(err)
}

func (adapter *MessagingStore) FindConversation(ctx context.Context, input slack.ConversationInput) (slack.ConversationRecord, error) {
	record, err := adapter.backend.FindConversation(ctx, mapConversationInput(input))
	return slack.ConversationRecord{ID: record.ID, UpdatedAt: record.UpdatedAt}, mapMessagingError(err)
}

func (adapter *MessagingStore) FindChannelConversation(ctx context.Context, input slack.ConversationInput) (slack.ConversationRecord, error) {
	record, err := adapter.backend.FindChannelConversation(ctx, mapConversationInput(input))
	return slack.ConversationRecord{ID: record.ID, UpdatedAt: record.UpdatedAt}, mapMessagingError(err)
}

func (adapter *MessagingStore) AppendMessage(ctx context.Context, conversationID uuid.UUID, externalMessageID, role, content string) error {
	return mapMessagingError(adapter.backend.AppendMessage(ctx, conversationID, externalMessageID, role, content))
}

func (adapter *MessagingStore) ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]slack.MessageRecord, error) {
	records, err := adapter.backend.ListRecentMessages(ctx, conversationID, limit)
	if err != nil {
		return nil, mapMessagingError(err)
	}
	result := make([]slack.MessageRecord, 0, len(records))
	for _, record := range records {
		result = append(result, slack.MessageRecord{
			ExternalMessageID: record.ExternalMessageID, Role: record.Role,
			Content: record.Content, CreatedAt: record.CreatedAt,
		})
	}
	return result, nil
}

func (adapter *MessagingStore) StartOutboundDelivery(ctx context.Context, input slack.OutboundDeliveryInput) (slack.OutboundDeliveryRecord, bool, error) {
	record, claimed, err := adapter.backend.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		InstallGeneration: input.InstallGeneration, ExternalWorkspaceID: input.ExternalWorkspaceID,
		ExternalRecipientUserID: input.ExternalRecipientUserID, InboundEventID: input.InboundEventID,
		IdempotencyKey: input.IdempotencyKey, ExternalChannelID: input.ExternalChannelID,
		ExternalThreadID: input.ExternalThreadID, Content: input.Content,
		ProviderPayload: append([]byte(nil), input.ProviderPayload...), Purpose: input.Purpose, ExpiresAt: input.ExpiresAt,
	})
	return mapOutboundDelivery(record), claimed, mapMessagingError(err)
}

func (adapter *MessagingStore) SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error {
	return mapMessagingError(adapter.backend.SetOutboundDeliveryContent(ctx, id, content))
}

func (adapter *MessagingStore) SetOutboundDeliveryContentAndDestination(ctx context.Context, id uuid.UUID, content, externalChannelID, externalThreadID string) error {
	return mapMessagingError(adapter.backend.SetOutboundDeliveryContentAndDestination(ctx, id, content, externalChannelID, externalThreadID))
}

func (adapter *MessagingStore) SetOutboundDeliveryContentAndProviderPayload(ctx context.Context, id uuid.UUID, content string, payload []byte) error {
	return mapMessagingError(adapter.backend.SetOutboundDeliveryContentAndProviderPayload(
		ctx,
		id,
		content,
		append([]byte(nil), payload...),
	))
}

func (adapter *MessagingStore) CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error {
	return mapMessagingError(adapter.backend.CompleteOutboundDelivery(ctx, id, externalMessageID))
}

func (adapter *MessagingStore) FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	return mapMessagingError(adapter.backend.FailOutboundDelivery(ctx, id, message))
}

func (adapter *MessagingStore) CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	return mapMessagingError(adapter.backend.CancelOutboundDelivery(ctx, id, message))
}

func (adapter *MessagingStore) ListRecoverableOutboundDeliveries(ctx context.Context, provider string, limit int) ([]slack.OutboundDeliveryRecord, error) {
	records, err := adapter.backend.ListRecoverableOutboundDeliveries(ctx, provider, limit)
	if err != nil {
		return nil, mapMessagingError(err)
	}
	result := make([]slack.OutboundDeliveryRecord, 0, len(records))
	for _, record := range records {
		result = append(result, mapOutboundDelivery(record))
	}
	return result, nil
}

func mapConversationInput(input slack.ConversationInput) messagingrepository.ConversationInput {
	return messagingrepository.ConversationInput{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID,
		ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalChannelID: input.ExternalChannelID,
		ExternalThreadID: input.ExternalThreadID, UserID: input.UserID,
		AudienceScope: input.AudienceScope, AudienceFingerprint: input.AudienceFingerprint,
	}
}

func mapOutboundDelivery(record messagingrepository.OutboundDeliveryRecord) slack.OutboundDeliveryRecord {
	return slack.OutboundDeliveryRecord{
		ID: record.ID, WorkspaceID: record.WorkspaceID, UserID: record.UserID,
		InstallGeneration: record.InstallGeneration, ExternalWorkspaceID: record.ExternalWorkspaceID,
		ExternalRecipientUserID: record.ExternalRecipientUserID, InboundEventID: record.InboundEventID,
		IdempotencyKey: record.IdempotencyKey, ExternalChannelID: record.ExternalChannelID,
		ExternalThreadID: record.ExternalThreadID, ExternalMessageID: record.ExternalMessageID,
		Content: record.Content, ProviderPayload: append([]byte(nil), record.ProviderPayload...),
		Status: record.Status, AttemptCount: record.AttemptCount, Purpose: record.Purpose, ExpiresAt: record.ExpiresAt,
	}
}

func mapMessagingError(err error) error {
	switch {
	case errors.Is(err, messagingrepository.ErrLeaseBusy):
		return errors.Join(slack.ErrOutboundDeliveryBusy, err)
	case errors.Is(err, messagingrepository.ErrNotFound):
		return errors.Join(slack.ErrMessagingRecordNotFound, err)
	default:
		return err
	}
}

var (
	_ slack.EventInbox      = (*MessagingStore)(nil)
	_ slack.OutboundStore   = (*MessagingStore)(nil)
	_ slack.NonceStore      = (*MessagingStore)(nil)
	_ slack.SlackEventStore = (*MessagingStore)(nil)
)
