package notificationsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Create(ctx context.Context, notification notificationsdomain.NewNotification) (notificationsdomain.Notification, bool, error) {
	if err := notification.Validate(); err != nil {
		return notificationsdomain.Notification{}, false, err
	}
	message, dedupeKey, err := marshalNewNotification(notification)
	if err != nil {
		return notificationsdomain.Notification{}, false, err
	}

	params := notificationssql.CreateNotificationParams{
		DedupeKey:        dedupeKey,
		NotificationType: notificationssql.NotificationType(notification.Type),
		EntityType:       notificationssql.EntityType(notification.EntityType),
		EntityID:         notification.EntityID,
		Title:            strings.TrimSpace(notification.Title),
		Message:          message,
		InAppEnabled:     notification.InAppEnabled,
		ActorID:          notification.ActorID,
		WorkspaceID:      notification.WorkspaceID,
		RecipientID:      notification.RecipientID,
	}
	row, err := repository.queries.CreateNotification(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		// PostgreSQL's INSERT ... ON CONFLICT can observe an uncommitted
		// conflicting row while the rest of the statement cannot see that row in
		// its original snapshot. One bounded retry starts a fresh snapshot, making
		// concurrent exact replays deterministic without retrying other failures.
		row, err = repository.queries.CreateNotification(ctx, params)
	}
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return notificationsdomain.Notification{}, false, mapWriteError("create notification", err)
		}
		exists, existsErr := repository.queries.NotificationDedupeKeyExists(ctx, notificationssql.NotificationDedupeKeyExistsParams{
			DedupeKey:   dedupeKey,
			RecipientID: notification.RecipientID,
			WorkspaceID: notification.WorkspaceID,
			ActorID:     notification.ActorID,
		})
		if existsErr != nil {
			return notificationsdomain.Notification{}, false, fmt.Errorf("classify notification create: %w", existsErr)
		}
		if exists {
			return notificationsdomain.Notification{}, false, fmt.Errorf("create notification: %w", notificationsdomain.ErrConflict)
		}
		return notificationsdomain.Notification{}, false, fmt.Errorf("create notification: %w", notificationsdomain.ErrForbidden)
	}

	mapped, err := toNotification(notificationRecord{
		ID:           row.NotificationID,
		RecipientID:  row.RecipientID,
		WorkspaceID:  row.WorkspaceID,
		Type:         row.Type,
		EntityType:   row.EntityType,
		EntityID:     row.EntityID,
		ActorID:      row.ActorID,
		Title:        row.Title,
		Message:      row.Message,
		InAppEnabled: row.InAppEnabled,
		CreatedAt:    row.CreatedAt,
		ReadAt:       row.ReadAt,
	})
	if err != nil {
		return notificationsdomain.Notification{}, false, fmt.Errorf("map created notification: %w", err)
	}
	return mapped, row.Inserted, nil
}
