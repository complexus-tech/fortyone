package notificationsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) List(ctx context.Context, userID, workspaceID uuid.UUID, search string, limit, offset int) ([]notifications.CoreNotification, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.List")
	defer span.End()

	query := `
		SELECT notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, message, created_at, read_at
		FROM notifications
		WHERE recipient_id = :user_id 
		AND workspace_id = :workspace_id
		AND entity_type::text <> 'feedback'
		AND (
			:search = ''
			OR title ILIKE '%' || :search || '%'
			OR message::text ILIKE '%' || :search || '%'
		)
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
		"search":       strings.TrimSpace(search),
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
		AND entity_type::text <> 'feedback'
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

func (r *repo) ListPortalFeedback(ctx context.Context, userID uuid.UUID, portalSlug string, unreadOnly bool, limit, offset int) ([]notifications.CorePortalNotification, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.ListPortalFeedback")
	defer span.End()

	var rows []dbPortalNotification
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT n.notification_id, n.recipient_id, n.workspace_id, n.type, n.entity_type,
			n.entity_id, n.actor_id, n.title, n.message, n.created_at, n.read_at,
			COALESCE(NULLIF(actor.full_name, ''), NULLIF(actor.username, ''), actor.email, 'Someone') AS actor_name,
			actor.avatar_url AS actor_avatar,
			fi.title AS feedback_title,
			fi.slug AS feedback_slug
		FROM notifications n
		INNER JOIN feedback_items fi ON fi.id = n.entity_id
		INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.workspace_id = n.workspace_id
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		LEFT JOIN users actor ON actor.user_id = n.actor_id
		WHERE n.recipient_id = $1
			AND w.slug = $2
			AND fp.is_public = true
			AND n.entity_type::text = 'feedback'
			AND n.type::text IN ('feedback_comment', 'feedback_status_update')
			AND ($3 = false OR n.read_at IS NULL)
		ORDER BY n.created_at DESC
		LIMIT $4 OFFSET $5
	`, userID, portalSlug, unreadOnly, limit, offset); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("list portal feedback notifications: %w", err)
	}

	result, err := toCorePortalNotifications(rows)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("convert portal feedback notifications: %w", err)
	}
	span.SetAttributes(attribute.Int("notifications.count", len(result)))
	return result, nil
}

func (r *repo) GetPortalFeedbackUnreadCount(ctx context.Context, userID uuid.UUID, portalSlug string) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.GetPortalFeedbackUnreadCount")
	defer span.End()

	var count int
	if err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM notifications n
		INNER JOIN feedback_items fi ON fi.id = n.entity_id
		INNER JOIN feedback_portals fp ON fp.id = fi.portal_id AND fp.workspace_id = n.workspace_id
		INNER JOIN workspaces w ON w.workspace_id = fp.workspace_id
		WHERE n.recipient_id = $1
			AND w.slug = $2
			AND fp.is_public = true
			AND n.entity_type::text = 'feedback'
			AND n.type::text IN ('feedback_comment', 'feedback_status_update')
			AND n.read_at IS NULL
	`, userID, portalSlug); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("count unread portal feedback notifications: %w", err)
	}

	span.SetAttributes(attribute.Int("notifications.unread_count", count))
	return count, nil
}
