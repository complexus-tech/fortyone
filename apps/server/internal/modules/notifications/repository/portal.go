package notificationsrepository

import (
	"context"
	"errors"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) ListPortalFeedback(ctx context.Context, query notificationsdomain.PortalListQuery) ([]notificationsdomain.PortalNotification, error) {
	query.Access = query.Access.Normalized()
	if err := query.Validate(); err != nil {
		return nil, err
	}
	if err := repository.authorizePortal(ctx, query.Access); err != nil {
		return nil, err
	}
	limit, err := safecast.Int32(query.Limit)
	if err != nil {
		return nil, fmt.Errorf("convert portal notification limit: %w", err)
	}
	offset, err := safecast.Int32(query.Offset)
	if err != nil {
		return nil, fmt.Errorf("convert portal notification offset: %w", err)
	}
	rows, err := repository.queries.ListPortalFeedbackNotifications(ctx, notificationssql.ListPortalFeedbackNotificationsParams{
		ResultOffset: offset,
		ResultLimit:  limit,
		PortalSlug:   query.Access.PortalSlug,
		ActorID:      query.Access.ActorID,
		UnreadOnly:   query.UnreadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("list portal feedback notifications: %w", err)
	}
	result := make([]notificationsdomain.PortalNotification, 0, len(rows))
	for _, row := range rows {
		notification, err := toNotification(notificationRecord{
			ID: row.NotificationID, RecipientID: row.RecipientID, WorkspaceID: row.WorkspaceID,
			Type: row.Type, EntityType: row.EntityType, EntityID: row.EntityID, ActorID: row.ActorID,
			Title: row.Title, Message: row.Message, InAppEnabled: row.InAppEnabled,
			CreatedAt: row.CreatedAt, ReadAt: row.ReadAt,
		})
		if err != nil {
			return nil, fmt.Errorf("map portal feedback notification: %w", err)
		}
		result = append(result, notificationsdomain.PortalNotification{
			Notification:  notification,
			ActorName:     row.ActorName,
			ActorAvatar:   row.ActorAvatar,
			FeedbackTitle: row.FeedbackTitle,
			FeedbackSlug:  row.FeedbackSlug,
		})
	}
	return result, nil
}

func (repository *Repository) CountUnreadPortalFeedback(ctx context.Context, access notificationsdomain.PortalAccess) (int, error) {
	access = access.Normalized()
	if err := access.Validate(); err != nil {
		return 0, err
	}
	if err := repository.authorizePortal(ctx, access); err != nil {
		return 0, err
	}
	count, err := repository.queries.CountUnreadPortalFeedbackNotifications(ctx, notificationssql.CountUnreadPortalFeedbackNotificationsParams{
		PortalSlug: access.PortalSlug,
		ActorID:    access.ActorID,
	})
	if err != nil {
		return 0, fmt.Errorf("count unread portal feedback notifications: %w", err)
	}
	converted, err := safecast.Int64(count)
	if err != nil {
		return 0, fmt.Errorf("convert portal notification count: %w", err)
	}
	return converted, nil
}

func (repository *Repository) MarkPortalFeedbackRead(ctx context.Context, command notificationsdomain.PortalNotificationMutation) error {
	command.Access = command.Access.Normalized()
	if err := command.Validate(); err != nil {
		return err
	}
	_, err := repository.queries.MarkPortalFeedbackNotificationRead(ctx, notificationssql.MarkPortalFeedbackNotificationReadParams{
		ReadAt:         command.At,
		PortalSlug:     command.Access.PortalSlug,
		ActorID:        command.Access.ActorID,
		NotificationID: command.NotificationID,
	})
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return mapWriteError("mark portal feedback notification read", err)
	}
	if authErr := repository.authorizePortal(ctx, command.Access); authErr != nil {
		return authErr
	}
	return fmt.Errorf("mark portal feedback notification read: %w", notificationsdomain.ErrNotFound)
}

func (repository *Repository) MarkAllPortalFeedbackRead(ctx context.Context, command notificationsdomain.PortalMutation) (int, error) {
	command.Access = command.Access.Normalized()
	if err := command.Validate(); err != nil {
		return 0, err
	}
	if err := repository.authorizePortal(ctx, command.Access); err != nil {
		return 0, err
	}
	count, err := repository.queries.MarkAllPortalFeedbackNotificationsRead(ctx, notificationssql.MarkAllPortalFeedbackNotificationsReadParams{
		PortalSlug: command.Access.PortalSlug,
		ActorID:    command.Access.ActorID,
		ReadAt:     command.At,
	})
	if err != nil {
		return 0, mapWriteError("mark all portal feedback notifications read", err)
	}
	converted, err := safecast.Int64(count)
	if err != nil {
		return 0, fmt.Errorf("convert marked portal notification count: %w", err)
	}
	return converted, nil
}

func (repository *Repository) authorizePortal(ctx context.Context, access notificationsdomain.PortalAccess) error {
	authorized, err := repository.queries.PortalNotificationActorAuthorized(ctx, notificationssql.PortalNotificationActorAuthorizedParams{
		PortalSlug: access.PortalSlug,
		ActorID:    access.ActorID,
	})
	if err != nil {
		return fmt.Errorf("authorize portal notifications: %w", err)
	}
	if !authorized {
		return fmt.Errorf("authorize portal notifications: %w", notificationsdomain.ErrForbidden)
	}
	return nil
}
