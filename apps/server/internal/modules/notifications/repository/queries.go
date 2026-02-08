package notificationsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]notifications.CoreNotification, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.List")
	defer span.End()

	query := `
		SELECT notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, message, created_at, read_at
		FROM notifications
		WHERE recipient_id = :user_id 
		AND workspace_id = :workspace_id
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"limit":        limit,
		"offset":       offset,
	}

	var dbNotifications []dbNotification
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbNotifications, params); err != nil {
		errMsg := fmt.Sprintf("failed to get notifications: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get notifications"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	span.AddEvent("notifications retrieved", trace.WithAttributes(
		attribute.Int("notifications.count", len(dbNotifications)),
	))

	coreNotifications, err := toCoreNotifications(dbNotifications)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert notifications: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to convert notifications"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return coreNotifications, nil
}

func (r *repo) GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.GetUnreadCount")
	defer span.End()

	query := `
		SELECT COUNT(*) FROM notifications
		WHERE recipient_id = :user_id
		AND workspace_id = :workspace_id
		AND read_at IS NULL;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to get unread count: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get unread count"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	span.AddEvent("unread count retrieved", trace.WithAttributes(
		attribute.Int("unread_count", count),
	))

	return count, nil
}

func (r *repo) GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (notifications.CoreNotificationPreferences, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.GetPreferences")
	defer span.End()

	query := `
		SELECT preference_id, user_id, workspace_id, preferences, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = :user_id AND workspace_id = :workspace_id;
	`

	params := map[string]interface{}{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	var dbPreferences dbNotificationPreferences
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotificationPreferences{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &dbPreferences, params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create default preferences if not found
			return r.createDefaultPreferences(ctx, userID, workspaceID)
		}
		errMsg := fmt.Sprintf("failed to get notification preferences: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get notification preferences"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotificationPreferences{}, err
	}

	span.AddEvent("preferences retrieved")
	corePrefs, err := toCoreNotificationPreferences(dbPreferences)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert preferences: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to convert preferences"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotificationPreferences{}, err
	}
	return corePrefs, nil
}
