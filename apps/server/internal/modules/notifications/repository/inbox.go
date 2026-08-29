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

func (repository *Repository) List(ctx context.Context, query notificationsdomain.ListQuery) ([]notificationsdomain.Notification, error) {
	query = query.Normalized()
	if err := query.Validate(); err != nil {
		return nil, err
	}
	if err := repository.authorizeWorkspace(ctx, query.Access); err != nil {
		return nil, err
	}
	limit, err := safecast.Int32(query.Limit)
	if err != nil {
		return nil, fmt.Errorf("convert notification limit: %w", err)
	}
	offset, err := safecast.Int32(query.Offset)
	if err != nil {
		return nil, fmt.Errorf("convert notification offset: %w", err)
	}
	rows, err := repository.queries.ListWorkspaceNotifications(ctx, notificationssql.ListWorkspaceNotificationsParams{
		Search:       query.Search,
		ResultOffset: offset,
		ResultLimit:  limit,
		WorkspaceID:  query.Access.WorkspaceID,
		ActorID:      query.Access.ActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace notifications: %w", err)
	}
	result := make([]notificationsdomain.Notification, 0, len(rows))
	for _, row := range rows {
		notification, err := toNotification(notificationRecord{
			ID: row.NotificationID, RecipientID: row.RecipientID, WorkspaceID: row.WorkspaceID,
			Type: row.Type, EntityType: row.EntityType, EntityID: row.EntityID, ActorID: row.ActorID,
			Title: row.Title, Message: row.Message, InAppEnabled: row.InAppEnabled,
			CreatedAt: row.CreatedAt, ReadAt: row.ReadAt,
		})
		if err != nil {
			return nil, fmt.Errorf("map workspace notification: %w", err)
		}
		result = append(result, notification)
	}
	return result, nil
}

func (repository *Repository) CountUnread(ctx context.Context, access notificationsdomain.WorkspaceAccess) (int, error) {
	if err := access.Validate(); err != nil {
		return 0, err
	}
	if err := repository.authorizeWorkspace(ctx, access); err != nil {
		return 0, err
	}
	count, err := repository.queries.CountUnreadWorkspaceNotifications(ctx, notificationssql.CountUnreadWorkspaceNotificationsParams{
		ActorID: access.ActorID, WorkspaceID: access.WorkspaceID,
	})
	if err != nil {
		return 0, fmt.Errorf("count unread workspace notifications: %w", err)
	}
	converted, err := safecast.Int64(int64(count))
	if err != nil {
		return 0, fmt.Errorf("convert unread notification count: %w", err)
	}
	return converted, nil
}

func (repository *Repository) Mutate(ctx context.Context, command notificationsdomain.NotificationMutation) error {
	if err := command.Validate(); err != nil {
		return err
	}
	_, err := repository.queries.MutateWorkspaceNotification(ctx, notificationssql.MutateWorkspaceNotificationParams{
		WorkspaceID:        command.Access.WorkspaceID,
		ActorID:            command.Access.ActorID,
		NotificationID:     command.NotificationID,
		DeleteNotification: command.Kind == notificationsdomain.NotificationMutationDelete,
		MarkRead:           command.Kind == notificationsdomain.NotificationMutationRead,
		MutatedAt:          command.At,
	})
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return mapWriteError("mutate workspace notification", err)
	}
	if authErr := repository.authorizeWorkspace(ctx, command.Access); authErr != nil {
		return authErr
	}
	return fmt.Errorf("mutate workspace notification: %w", notificationsdomain.ErrNotFound)
}

func (repository *Repository) MutateAll(ctx context.Context, command notificationsdomain.WorkspaceMutation) (int, error) {
	if err := command.Validate(); err != nil {
		return 0, err
	}
	if err := repository.authorizeWorkspace(ctx, command.Access); err != nil {
		return 0, err
	}
	count, err := repository.queries.MutateWorkspaceNotifications(ctx, notificationssql.MutateWorkspaceNotificationsParams{
		WorkspaceID:         command.Access.WorkspaceID,
		ActorID:             command.Access.ActorID,
		DeleteNotifications: command.Kind != notificationsdomain.WorkspaceMutationReadAll,
		OnlyRead:            command.Kind == notificationsdomain.WorkspaceMutationDeleteRead,
		MutatedAt:           command.At,
	})
	if err != nil {
		return 0, mapWriteError("mutate workspace notifications", err)
	}
	converted, err := safecast.Int64(int64(count))
	if err != nil {
		return 0, fmt.Errorf("convert mutated notification count: %w", err)
	}
	return converted, nil
}

func (repository *Repository) authorizeWorkspace(ctx context.Context, access notificationsdomain.WorkspaceAccess) error {
	authorized, err := repository.queries.WorkspaceNotificationActorAuthorized(ctx, notificationssql.WorkspaceNotificationActorAuthorizedParams{
		WorkspaceID: access.WorkspaceID,
		ActorID:     access.ActorID,
	})
	if err != nil {
		return fmt.Errorf("authorize workspace notifications: %w", err)
	}
	if !authorized {
		return fmt.Errorf("authorize workspace notifications: %w", notificationsdomain.ErrForbidden)
	}
	return nil
}
