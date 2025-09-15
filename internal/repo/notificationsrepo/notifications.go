package notificationsrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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
			entity_id, actor_id, title, message
		) 
		SELECT 
			:recipient_id, :workspace_id, :type, :entity_type,
			:entity_id, :actor_id, :title, :message
		FROM users u
		WHERE u.user_id = :recipient_id 
			AND u.is_active = true
		ON CONFLICT (recipient_id, workspace_id, entity_id, entity_type)
		DO UPDATE SET
			title = :title,
			message = :message,
			actor_id = :actor_id,
			read_at = NULL,
			created_at = CURRENT_TIMESTAMP
		RETURNING notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, message, created_at, read_at;
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

	dbNotif, err := toDBNewNotification(n)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert notification: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to convert notification"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotification{}, err
	}

	if err := stmt.GetContext(ctx, &notification, dbNotif); err != nil {
		errMsg := fmt.Sprintf("failed to create notification: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create notification"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotification{}, err
	}

	span.AddEvent("notification created", trace.WithAttributes(
		attribute.String("notification.id", notification.ID.String()),
	))

	coreNotif, err := toCoreNotification(notification)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert notification: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to convert notification"), trace.WithAttributes(attribute.String("error", errMsg)))
		return notifications.CoreNotification{}, err
	}

	return coreNotif, nil
}

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

func (r *repo) UpdatePreference(ctx context.Context, userID, workspaceID uuid.UUID, notificationType string, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.UpdatePreference")
	defer span.End()

	if len(updates) == 0 {
		return nil
	}

	// First get current preferences
	query := `
		SELECT preferences FROM notification_preferences
		WHERE user_id = :user_id AND workspace_id = :workspace_id;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	var prefsJSON json.RawMessage
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowxContext(ctx, params).Scan(&prefsJSON)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Parse current preferences or use defaults
	prefs := make(map[string]map[string]bool)
	if errors.Is(err, sql.ErrNoRows) || len(prefsJSON) == 0 {
		// Initialize with defaults
		prefs = getDefaultPreferences()
	} else {
		if err := json.Unmarshal(prefsJSON, &prefs); err != nil {
			return err
		}
	}

	// Check if notification type exists
	if _, exists := prefs[notificationType]; !exists {
		prefs[notificationType] = map[string]bool{
			"email":  true,
			"in_app": true,
		}
	}

	// Update specific channel preferences
	for key, value := range updates {
		switch key {
		case "email_enabled":
			prefs[notificationType]["email"] = value.(bool)
		case "in_app_enabled":
			prefs[notificationType]["in_app"] = value.(bool)
		}
	}

	// Marshal back to JSON
	updatedJSON, err := json.Marshal(prefs)
	if err != nil {
		return err
	}

	// Update the database
	updateQuery := `
		INSERT INTO notification_preferences (user_id, workspace_id, preferences)
		VALUES (:user_id, :workspace_id, :preferences)
		ON CONFLICT (user_id, workspace_id) 
		DO UPDATE SET preferences = :preferences, updated_at = CURRENT_TIMESTAMP;
	`

	updateParams := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"preferences":  updatedJSON,
	}

	updateStmt, err := r.db.PrepareNamedContext(ctx, updateQuery)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	_, err = updateStmt.ExecContext(ctx, updateParams)
	return err
}

func (r *repo) createDefaultPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (notifications.CoreNotificationPreferences, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.CreateDefaultPreferences")
	defer span.End()

	defaultPrefs := getDefaultPreferences()
	prefsJSON, err := json.Marshal(defaultPrefs)
	if err != nil {
		return notifications.CoreNotificationPreferences{}, err
	}

	query := `
		INSERT INTO notification_preferences (user_id, workspace_id, preferences)
		VALUES (:user_id, :workspace_id, :preferences)
		RETURNING preference_id, user_id, workspace_id, preferences, created_at, updated_at;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"preferences":  prefsJSON,
	}

	var result dbNotificationPreferences
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return notifications.CoreNotificationPreferences{}, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &result, params)
	if err != nil {
		return notifications.CoreNotificationPreferences{}, err
	}

	corePrefs, err := toCoreNotificationPreferences(result)
	if err != nil {
		return notifications.CoreNotificationPreferences{}, err
	}

	return corePrefs, nil
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

	params := map[string]any{
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

// Helper function for default preferences
func getDefaultPreferences() map[string]map[string]bool {
	return map[string]map[string]bool{
		"story_update": {
			"email":  true,
			"in_app": true,
		},
		"story_comment": {
			"email":  true,
			"in_app": true,
		},
		"comment_reply": {
			"email":  true,
			"in_app": true,
		},
		"objective_update": {
			"email":  true,
			"in_app": true,
		},
		"key_result_update": {
			"email":  true,
			"in_app": true,
		},
		"mention": {
			"email":  true,
			"in_app": true,
		},
		"overdue_stories": {
			"email":  true,
			"in_app": true,
		},
	}
}
