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

	query := listWorkspaceNotificationsQuery()

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

func listWorkspaceNotificationsQuery() string {
	return `
		SELECT notification_id, recipient_id, workspace_id, type, entity_type,
			entity_id, actor_id, title, message, created_at, read_at
		FROM notifications notification
		WHERE notification.recipient_id = :user_id
		AND notification.workspace_id = :workspace_id
		AND CAST(notification.entity_type AS text) <> 'feedback'
		AND ` + workspaceNotificationAccessPredicate("notification") + `
		AND (
			:search = ''
			OR title ILIKE '%' || :search || '%'
			OR (
				CAST(notification.entity_type AS text) <> 'strategy'
				AND CAST(message AS text) ILIKE '%' || :search || '%'
			)
		)
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset;
	`
}

func (r *repo) GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.notifications.GetUnreadCount")
	defer span.End()

	query := unreadWorkspaceNotificationsQuery()

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

func unreadWorkspaceNotificationsQuery() string {
	return `
		SELECT COUNT(*) FROM notifications notification
		WHERE notification.recipient_id = :user_id
		AND notification.workspace_id = :workspace_id
		AND CAST(notification.entity_type AS text) <> 'feedback'
		AND notification.read_at IS NULL
		AND ` + workspaceNotificationAccessPredicate("notification") + `;
	`
}

func workspaceNotificationAccessPredicate(alias string) string {
	return fmt.Sprintf(`EXISTS (
		SELECT 1
		FROM workspace_members notification_member
		WHERE notification_member.workspace_id = %[1]s.workspace_id
			AND notification_member.user_id = %[1]s.recipient_id
			AND notification_member.role IN ('admin', 'member', 'guest')
			AND (
			(
				CAST(%[1]s.entity_type AS TEXT) = 'story'
				AND EXISTS (
					SELECT 1
					FROM stories notification_story
					WHERE notification_story.id = %[1]s.entity_id
						AND notification_story.workspace_id = %[1]s.workspace_id
						AND notification_story.deleted_at IS NULL
						AND (
							notification_member.role = 'admin'
							OR EXISTS (
								SELECT 1
								FROM team_members story_member
								WHERE story_member.team_id = notification_story.team_id
									AND story_member.user_id = %[1]s.recipient_id
							)
						)
				)
			)
			OR (
				CAST(%[1]s.entity_type AS TEXT) = 'comment'
				AND EXISTS (
					SELECT 1
					FROM story_comments notification_comment
					INNER JOIN stories notification_comment_story
						ON notification_comment_story.id = notification_comment.story_id
					WHERE notification_comment.comment_id = %[1]s.entity_id
						AND notification_comment_story.workspace_id = %[1]s.workspace_id
						AND notification_comment_story.deleted_at IS NULL
						AND (
							notification_member.role = 'admin'
							OR EXISTS (
								SELECT 1
								FROM team_members comment_member
								WHERE comment_member.team_id = notification_comment_story.team_id
									AND comment_member.user_id = %[1]s.recipient_id
							)
						)
				)
			)
			OR (
				CAST(%[1]s.entity_type AS TEXT) = 'objective'
				AND EXISTS (
					SELECT 1
					FROM objectives notification_objective
					WHERE notification_objective.objective_id = %[1]s.entity_id
						AND notification_objective.workspace_id = %[1]s.workspace_id
						AND (
							notification_member.role = 'admin'
							OR EXISTS (
								SELECT 1
								FROM team_members objective_member
								WHERE objective_member.team_id = notification_objective.team_id
									AND objective_member.user_id = %[1]s.recipient_id
							)
						)
				)
			)
			OR (
				CAST(%[1]s.entity_type AS TEXT) = 'key_result'
				AND EXISTS (
					SELECT 1
					FROM key_results notification_key_result
					INNER JOIN objectives notification_key_result_objective
						ON notification_key_result_objective.objective_id = notification_key_result.objective_id
					WHERE notification_key_result.id = %[1]s.entity_id
						AND notification_key_result_objective.workspace_id = %[1]s.workspace_id
						AND (
							notification_member.role = 'admin'
							OR EXISTS (
								SELECT 1
								FROM team_members key_result_member
								WHERE key_result_member.team_id = notification_key_result_objective.team_id
									AND key_result_member.user_id = %[1]s.recipient_id
							)
						)
				)
			)
			OR (
				CAST(%[1]s.entity_type AS TEXT) = 'strategy'
				AND (
					notification_member.role = 'admin'
					OR %[1]s.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
				)
			)
			)
	)`, alias)
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
			AND fi.deleted_at IS NULL
			AND CAST(n.entity_type AS text) = 'feedback'
			AND CAST(n.type AS text) IN ('feedback_comment', 'feedback_status_update')
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
			AND fi.deleted_at IS NULL
			AND CAST(n.entity_type AS text) = 'feedback'
			AND CAST(n.type AS text) IN ('feedback_comment', 'feedback_status_update')
			AND n.read_at IS NULL
	`, userID, portalSlug); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("count unread portal feedback notifications: %w", err)
	}

	span.SetAttributes(attribute.Int("notifications.unread_count", count))
	return count, nil
}
