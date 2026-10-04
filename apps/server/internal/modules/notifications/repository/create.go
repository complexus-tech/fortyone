package notificationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Create returns mutated=true for both a new row and a refreshed unread row.
// Event receipts retain every consumed identity, independently of inbox state.
func (repository *Repository) Create(ctx context.Context, notification notificationsdomain.NewNotification) (notificationsdomain.Notification, bool, error) {
	if err := notification.Validate(); err != nil {
		return notificationsdomain.Notification{}, false, err
	}
	message, dedupeKey, err := marshalNewNotification(notification)
	if err != nil {
		return notificationsdomain.Notification{}, false, err
	}
	title := strings.TrimSpace(notification.Title)
	payload, err := json.Marshal([]any{notification.Type, notification.EntityType, notification.EntityID, notification.ActorID, title, json.RawMessage(message)})
	if err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("marshal notification receipt: %w", err)
	}
	occurredAt := notification.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	coalescingKey := "story_update"
	if notification.Message.ScheduleIssue != nil {
		coalescingKey = "schedule_issue:" + notification.Message.ScheduleIssue.ID.String()
	}

	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("begin notification create: %w", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	queries := notificationssql.New(tx)
	if err := queries.LockNotificationEvent(ctx, notificationssql.LockNotificationEventParams{DedupeKey: dedupeKey}); err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("lock notification event: %w", err)
	}
	receipt, err := queries.GetNotificationEventReceipt(ctx, notificationssql.GetNotificationEventReceiptParams{DedupeKey: dedupeKey, Payload: payload})
	if err == nil {
		if receipt.RecipientID != notification.RecipientID || receipt.WorkspaceID != notification.WorkspaceID {
			return notificationsdomain.Notification{}, false, fmt.Errorf("replay notification event: %w", notificationsdomain.ErrForbidden)
		}
		if !receipt.MatchesPayload {
			return notificationsdomain.Notification{}, false, fmt.Errorf("replay notification event: %w", notificationsdomain.ErrConflict)
		}
		if receipt.NotificationID == nil {
			return notificationsdomain.Notification{}, false, nil
		}
		row, err := queries.GetNotificationReceiptInboxRow(ctx, notificationssql.GetNotificationReceiptInboxRowParams{
			NotificationID: *receipt.NotificationID,
			RecipientID:    notification.RecipientID,
			WorkspaceID:    notification.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return notificationsdomain.Notification{}, false, nil
		}
		if err != nil {
			return notificationsdomain.Notification{}, false, fmt.Errorf("read replayed notification: %w", err)
		}
		mapped, err := toNotification(notificationRecord{
			ID: row.NotificationID, RecipientID: row.RecipientID, WorkspaceID: row.WorkspaceID,
			Type: row.Type, EntityType: row.EntityType, EntityID: row.EntityID,
			ActorID: row.ActorID, Title: row.Title, Message: row.Message,
			InAppEnabled: row.InAppEnabled, CreatedAt: row.CreatedAt, ReadAt: row.ReadAt,
		})
		return mapped, false, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return notificationsdomain.Notification{}, false, fmt.Errorf("read notification event receipt: %w", err)
	}
	if notification.Type == notificationsdomain.NotificationTypeStoryUpdate && notification.EntityType == notificationsdomain.EntityTypeStory {
		if err := queries.LockStoryNotification(ctx, notificationssql.LockStoryNotificationParams{
			RecipientID:   notification.RecipientID.String(),
			WorkspaceID:   notification.WorkspaceID.String(),
			EntityID:      notification.EntityID.String(),
			CoalescingKey: coalescingKey,
		}); err != nil {
			return notificationsdomain.Notification{}, false, fmt.Errorf("lock story notification: %w", err)
		}
	}
	row, err := queries.CreateNotification(ctx, notificationssql.CreateNotificationParams{
		DedupeKey:        dedupeKey,
		OccurredAt:       occurredAt,
		CoalescingKey:    coalescingKey,
		NotificationType: notificationssql.NotificationType(notification.Type),
		EntityType:       notificationssql.EntityType(notification.EntityType),
		EntityID:         notification.EntityID,
		Title:            title,
		Message:          message,
		InAppEnabled:     notification.InAppEnabled,
		ActorID:          notification.ActorID,
		WorkspaceID:      notification.WorkspaceID,
		RecipientID:      notification.RecipientID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return notificationsdomain.Notification{}, false, fmt.Errorf("create notification: %w", notificationsdomain.ErrForbidden)
	}
	if err != nil {
		return notificationsdomain.Notification{}, false, mapWriteError("create notification", err)
	}
	var notificationID *uuid.UUID
	if row.NotificationID != uuid.Nil {
		notificationID = &row.NotificationID
	}
	if err := queries.RecordNotificationEventReceipt(ctx, notificationssql.RecordNotificationEventReceiptParams{
		DedupeKey:        dedupeKey,
		RecipientID:      notification.RecipientID,
		WorkspaceID:      notification.WorkspaceID,
		ActorID:          notification.ActorID,
		EntityID:         notification.EntityID,
		NotificationType: notificationssql.NotificationType(notification.Type),
		EntityType:       notificationssql.EntityType(notification.EntityType),
		CoalescingKey:    coalescingKey,
		OccurredAt:       occurredAt,
		Payload:          payload,
		NotificationID:   notificationID,
	}); err != nil {
		return notificationsdomain.Notification{}, false, mapWriteError("record notification event receipt", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("commit notification create: %w", err)
	}
	if notificationID == nil {
		return notificationsdomain.Notification{}, false, nil
	}
	mapped, err := toNotification(notificationRecord{
		ID: row.NotificationID, RecipientID: row.RecipientID, WorkspaceID: row.WorkspaceID,
		Type: row.Type, EntityType: row.EntityType, EntityID: row.EntityID,
		ActorID: row.ActorID, Title: row.Title, Message: row.Message,
		InAppEnabled: row.InAppEnabled, CreatedAt: row.CreatedAt, ReadAt: row.ReadAt,
	})
	if err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("map created notification: %w", err)
	}
	return mapped, row.Mutated, nil
}
