package chatrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) ClaimEvents(ctx context.Context, limit int) ([]domain.Event, error) {
	if limit < 1 || limit > 500 {
		return nil, domain.ErrInvalid
	}
	rows, err := r.query(ctx).ClaimEvents(ctx, chatsql.ClaimEventsParams{PageLimit: int32(limit)})
	if err != nil {
		return nil, err
	}
	result := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		if row.LeaseToken == nil {
			continue
		}
		result = append(result, domain.Event{ID: row.ID, LeaseToken: *row.LeaseToken, Type: row.EventType, WorkspaceID: row.WorkspaceID, ConversationID: row.ConversationID, MessageID: row.MessageID, ActorID: row.ActorID, ThreadRootID: row.ThreadRootID, Sequence: row.Sequence, Version: int(row.Version), CreatedAt: row.CreatedAt})
	}
	return result, nil
}

func (r *Repository) CompleteEvent(ctx context.Context, eventID, leaseToken uuid.UUID) error {
	count, err := r.query(ctx).CompleteEvent(ctx, chatsql.CompleteEventParams{EventID: eventID, LeaseToken: &leaseToken})
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) ReleaseEvent(ctx context.Context, eventID, leaseToken uuid.UUID) error {
	count, err := r.query(ctx).ReleaseEvent(ctx, chatsql.ReleaseEventParams{EventID: eventID, LeaseToken: &leaseToken})
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) RecipientIDs(ctx context.Context, workspaceID, conversationID uuid.UUID) ([]uuid.UUID, error) {
	return r.query(ctx).RecipientIDs(ctx, chatsql.RecipientIDsParams{WorkspaceID: workspaceID, ConversationID: conversationID})
}

func (r *Repository) NotificationRecipients(ctx context.Context, event domain.Event) ([]domain.NotificationRecipient, error) {
	if event.Type != "message.created" {
		return []domain.NotificationRecipient{}, nil
	}
	rows, err := r.query(ctx).NotificationRecipients(ctx, chatsql.NotificationRecipientsParams{WorkspaceID: event.WorkspaceID, EventID: event.ID})
	if err != nil {
		return nil, err
	}
	result := make([]domain.NotificationRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.NotificationRecipient{UserID: row.UserID, Reason: row.Reason})
	}
	return result, nil
}

func (r *Repository) ReadRecipientIDs(ctx context.Context, event domain.Event) ([]uuid.UUID, error) {
	return r.query(ctx).ReadRecipientIDs(ctx, chatsql.ReadRecipientIDsParams{WorkspaceID: event.WorkspaceID, ConversationID: event.ConversationID, ActorID: event.ActorID})
}

func (r *Repository) RenewEventLease(ctx context.Context, eventID, leaseToken uuid.UUID) error {
	count, err := r.query(ctx).RenewEventLease(ctx, chatsql.RenewEventLeaseParams{EventID: eventID, LeaseToken: &leaseToken})
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrConflict
	}
	return nil
}
