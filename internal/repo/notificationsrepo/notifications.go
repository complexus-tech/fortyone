package notificationsrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/notifications"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

func (r *repo) Create(ctx context.Context, n notifications.CoreNewNotification) (notifications.CoreNotification, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.Create")
	defer span.End()

	query := `
		INSERT INTO notifications (
			recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, description
		) VALUES (
			:recipient_id, :workspace_id, :type, :entity_type,
			:entity_id, :actor_id, :title, :description	
		) RETURNING notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, description, created_at, read_at;
	`

	var notification dbNotification
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotification{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &notification, toDBNewNotification(n)); err != nil {
		errMsg := fmt.Sprintf("failed to create notification: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create notification"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotification{}, err
	}

	span.AddEvent("notification created", trace.WithAttributes(
		attribute.String("notification.id", notification.ID.String()),
	))

	return toCoreNotification(notification), nil
}

func (r *repo) List(ctx context.Context, userID, workspaceID uuid.UUID, limit, offset int) ([]notifications.CoreNotification, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.List")
	defer span.End()

	query := `
		SELECT notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, description, created_at, read_at
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

	return toCoreNotifications(dbNotifications), nil
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

func (r *repo) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.MarkAsRead")
	defer span.End()

	query := `
		UPDATE notifications
		SET read_at = CURRENT_TIMESTAMP
		WHERE notification_id = :notification_id
		AND recipient_id = :user_id;
	`

	params := map[string]any{
		"notification_id": notificationID,
		"user_id":         userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to mark notification as read: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to mark notification as read"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found or already read")
	}

	return nil
}

func (r *repo) GetPreferences(ctx context.Context, userID, workspaceID uuid.UUID) ([]notifications.CoreNotificationPreference, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.GetPreferences")
	defer span.End()

	query := `
		SELECT preference_id, user_id, workspace_id, notification_type,
			email_enabled, in_app_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = :user_id
		AND workspace_id = :workspace_id;
	`

	params := map[string]interface{}{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	var dbPreferences []dbNotificationPreference
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbPreferences, params); err != nil {
		errMsg := fmt.Sprintf("failed to get notification preferences: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get notification preferences"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	span.AddEvent("preferences retrieved", trace.WithAttributes(
		attribute.Int("preferences.count", len(dbPreferences)),
	))

	return toCoreNotificationPreferences(dbPreferences), nil
}

func (r *repo) UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.UpdatePreference")
	defer span.End()

	if len(updates) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(updates))
	for field := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	query := fmt.Sprintf(`
		UPDATE notification_preferences
		SET %s,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = :user_id
		AND workspace_id = :workspace_id
		AND notification_type = :notification_type;
	`, strings.Join(setClauses, ", "))

	updates["user_id"] = userID
	updates["workspace_id"] = workspaceID
	updates["notification_type"] = NotificationType(notificationType)

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, updates)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update notification preference: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update notification preference"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification preference not found")
	}

	return nil
}

func (r *repo) CreateDefaultPreferences(ctx context.Context, userID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.CreateDefaultPreferences")
	defer span.End()

	query := `
		INSERT INTO notification_preferences (
			user_id, workspace_id, notification_type
		) VALUES (
			:user_id, :workspace_id, :notification_type
		);
	`

	notificationTypes := []NotificationType{
		NotificationTypeStoryUpdate,
		NotificationTypeStoryComment,
		NotificationTypeCommentReply,
		NotificationTypeObjectiveUpdate,
		NotificationTypeKeyResultUpdate,
		NotificationTypeMention,
	}

	for _, nType := range notificationTypes {
		params := map[string]interface{}{
			"user_id":           userID,
			"workspace_id":      workspaceID,
			"notification_type": nType,
		}

		stmt, err := r.db.PrepareNamedContext(ctx, query)
		if err != nil {
			errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}

		if _, err := stmt.ExecContext(ctx, params); err != nil {
			stmt.Close()
			errMsg := fmt.Sprintf("failed to create default preference: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to create default preference"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
		stmt.Close()
	}

	return nil
}

func (r *repo) MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.MarkAllAsRead")
	defer span.End()

	query := `
		UPDATE notifications
		SET read_at = CURRENT_TIMESTAMP
		WHERE recipient_id = :user_id
		AND workspace_id = :workspace_id
		AND read_at IS NULL;
	`

	params := map[string]interface{}{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to mark all notifications as read: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to mark all notifications as read"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("notifications marked as read", trace.WithAttributes(
		attribute.Int64("notifications.count", rows),
	))

	return nil
}

func (r *repo) DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.DeleteNotification")
	defer span.End()

	query := `
		DELETE FROM notifications
		WHERE notification_id = :notification_id
		AND recipient_id = :user_id;
	`

	params := map[string]any{
		"notification_id": notificationID,
		"user_id":         userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete notification: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete notification"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found or not authorized to delete")
	}

	span.AddEvent("notification deleted", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return nil
}

func (r *repo) DeleteAllNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.DeleteAllNotifications")
	defer span.End()

	query := `
		DELETE FROM notifications
		WHERE recipient_id = :user_id
		AND workspace_id = :workspace_id;
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
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete all notifications: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete all notifications"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("all notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", rows),
	))

	return rows, nil
}

func (r *repo) DeleteReadNotifications(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.DeleteReadNotifications")
	defer span.End()

	query := `
		DELETE FROM notifications
		WHERE recipient_id = :user_id
		AND workspace_id = :workspace_id
		AND read_at IS NOT NULL;
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
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete read notifications: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete read notifications"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("read notifications deleted", trace.WithAttributes(
		attribute.Int64("notifications.count", rows),
	))

	return rows, nil
}

func (r *repo) MarkAsUnread(ctx context.Context, notificationID, userID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.MarkAsUnread")
	defer span.End()

	query := `
		UPDATE notifications
		SET read_at = NULL
		WHERE notification_id = :notification_id
		AND recipient_id = :user_id;
	`

	params := map[string]any{
		"notification_id": notificationID,
		"user_id":         userID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to mark notification as unread: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to mark notification as unread"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found or already unread")
	}

	span.AddEvent("notification marked as unread", trace.WithAttributes(
		attribute.String("notification.id", notificationID.String()),
	))

	return nil
}
