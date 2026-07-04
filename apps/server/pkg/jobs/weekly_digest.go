package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type WeeklyDigestRecipient struct {
	UserID        uuid.UUID `db:"user_id"`
	UserEmail     string    `db:"user_email"`
	UserName      string    `db:"user_name"`
	WorkspaceID   uuid.UUID `db:"workspace_id"`
	WorkspaceName string    `db:"workspace_name"`
	WorkspaceSlug string    `db:"workspace_slug"`
}

type WeeklyDigestStats struct {
	UnreadNotifications         int `db:"unread_notifications"`
	UnreadPriorityNotifications int `db:"unread_priority_notifications"`
	OverdueStories              int `db:"overdue_stories"`
	DueThisWeekStories          int `db:"due_this_week_stories"`
	ObjectiveRisks              int `db:"objective_risks"`
	TeamComments                int `db:"team_comments"`
}

func (s WeeklyDigestStats) hasSignal() bool {
	return s.UnreadNotifications+
		s.OverdueStories+
		s.DueThisWeekStories+
		s.ObjectiveRisks+
		s.TeamComments > 0
}

// ProcessWeeklyDigestEmail sends a weekly workspace digest to users with meaningful activity.
func ProcessWeeklyDigestEmail(ctx context.Context, db *sqlx.DB, log *logger.Logger, mailerService mailer.Service) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWeeklyDigestEmail")
	defer span.End()

	log.Info(ctx, "Processing weekly digest emails")
	startTime := time.Now()

	const recipientBatchSize = 100
	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0

	for {
		batchCount++
		recipients, err := getWeeklyDigestRecipients(ctx, db, recipientBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get weekly digest recipients batch %d: %w", batchCount, err)
		}
		if len(recipients) == 0 {
			break
		}

		for _, recipient := range recipients {
			stats, err := getWeeklyDigestStats(ctx, db, recipient.UserID, recipient.WorkspaceID)
			if err != nil {
				log.Error(ctx, "Failed to get weekly digest stats", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", err)
				continue
			}
			if !stats.hasSignal() {
				totalProcessed++
				continue
			}
			if err := sendWeeklyDigestEmail(ctx, log, mailerService, recipient, stats); err != nil {
				log.Error(ctx, "Failed to send weekly digest email", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", err)
				continue
			}
			totalEmailsSent++
			totalProcessed++
		}

		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	span.AddEvent("weekly digest email job completed", trace.WithAttributes(
		attribute.Int("recipients.processed", totalProcessed),
		attribute.Int("emails.sent", totalEmailsSent),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, "Weekly digest email job completed",
		"recipients_processed", totalProcessed,
		"emails_sent", totalEmailsSent,
		"batches_processed", batchCount,
		"duration", duration)
	return nil
}

func getWeeklyDigestRecipients(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]WeeklyDigestRecipient, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getWeeklyDigestRecipients")
	defer span.End()

	query := `
		SELECT
			wm.user_id,
			u.email AS user_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) AS user_name,
			w.workspace_id,
			w.name AS workspace_name,
			w.slug AS workspace_slug
		FROM workspace_members wm
		INNER JOIN users u ON wm.user_id = u.user_id
		INNER JOIN workspaces w ON wm.workspace_id = w.workspace_id
		LEFT JOIN notification_preferences np ON wm.user_id = np.user_id
			AND wm.workspace_id = np.workspace_id
		WHERE u.is_active = true
			AND u.is_system = false
			AND w.deleted_at IS NULL
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND CAST(COALESCE(np.preferences -> 'weekly_digest' ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY w.workspace_id, wm.user_id
		LIMIT :batch_size OFFSET :offset;
	`

	params := map[string]any{
		"batch_size": batchSize,
		"offset":     offset * batchSize,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare weekly digest recipients query: %w", err)
	}
	defer stmt.Close()

	var recipients []WeeklyDigestRecipient
	if err := stmt.SelectContext(ctx, &recipients, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute weekly digest recipients query: %w", err)
	}

	span.AddEvent("weekly digest recipients retrieved", trace.WithAttributes(
		attribute.Int("recipients.count", len(recipients)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))
	return recipients, nil
}

func getWeeklyDigestStats(ctx context.Context, db *sqlx.DB, userID, workspaceID uuid.UUID) (WeeklyDigestStats, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getWeeklyDigestStats")
	defer span.End()

	query := `
		SELECT
			(
				SELECT COUNT(*)
				FROM notifications n
				WHERE n.recipient_id = :user_id
					AND n.workspace_id = :workspace_id
					AND n.read_at IS NULL
			)::int AS unread_notifications,
			(
				SELECT COUNT(*)
				FROM notifications n
				WHERE n.recipient_id = :user_id
					AND n.workspace_id = :workspace_id
					AND n.read_at IS NULL
					AND n.created_at >= NOW() - INTERVAL '7 days'
					AND n.type IN ('mention', 'comment_reply')
			)::int AS unread_priority_notifications,
			(
				SELECT COUNT(*)
				FROM stories s
				INNER JOIN statuses st ON s.status_id = st.status_id
				WHERE s.assignee_id = :user_id
					AND s.workspace_id = :workspace_id
					AND s.end_date < CURRENT_DATE
					AND st.category NOT IN ('completed', 'cancelled', 'paused')
					AND s.deleted_at IS NULL
					AND s.archived_at IS NULL
					AND s.completed_at IS NULL
			)::int AS overdue_stories,
			(
				SELECT COUNT(*)
				FROM stories s
				INNER JOIN statuses st ON s.status_id = st.status_id
				WHERE s.assignee_id = :user_id
					AND s.workspace_id = :workspace_id
					AND s.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '7 days'
					AND st.category NOT IN ('completed', 'cancelled', 'paused')
					AND s.deleted_at IS NULL
					AND s.archived_at IS NULL
					AND s.completed_at IS NULL
			)::int AS due_this_week_stories,
			(
				SELECT COUNT(*)
				FROM objectives o
				INNER JOIN objective_statuses os ON o.status_id = os.status_id
				INNER JOIN workspace_settings ws ON o.workspace_id = ws.workspace_id
				WHERE o.lead_user_id = :user_id
					AND o.workspace_id = :workspace_id
					AND o.end_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE + INTERVAL '7 days'
					AND os.category NOT IN ('completed', 'cancelled', 'paused')
					AND ws.objective_enabled = true
			)::int AS objective_risks,
			(
				SELECT COUNT(*)
				FROM story_comments sc
				INNER JOIN stories s ON sc.story_id = s.id
				WHERE s.workspace_id = :workspace_id
					AND sc.commenter_id <> :user_id
					AND sc.created_at >= NOW() - INTERVAL '7 days'
					AND s.deleted_at IS NULL
			)::int AS team_comments;
	`

	params := map[string]any{
		"user_id":      userID,
		"workspace_id": workspaceID,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return WeeklyDigestStats{}, fmt.Errorf("failed to prepare weekly digest stats query: %w", err)
	}
	defer stmt.Close()

	var stats WeeklyDigestStats
	if err := stmt.GetContext(ctx, &stats, params); err != nil {
		span.RecordError(err)
		return WeeklyDigestStats{}, fmt.Errorf("failed to execute weekly digest stats query: %w", err)
	}

	span.AddEvent("weekly digest stats retrieved", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("unread_notifications", stats.UnreadNotifications),
		attribute.Int("overdue_stories", stats.OverdueStories),
		attribute.Int("due_this_week_stories", stats.DueThisWeekStories),
		attribute.Int("objective_risks", stats.ObjectiveRisks),
		attribute.Int("team_comments", stats.TeamComments),
	))
	return stats, nil
}

func sendWeeklyDigestEmail(ctx context.Context, log *logger.Logger, mailerService mailer.Service, recipient WeeklyDigestRecipient, stats WeeklyDigestStats) error {
	if strings.TrimSpace(recipient.UserEmail) == "" {
		log.Info(ctx, "Skipping weekly digest because user email is empty", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID)
		return nil
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", recipient.WorkspaceSlug)
	title := fmt.Sprintf("Weekly digest: %s", recipient.WorkspaceName)
	data := map[string]any{
		"UserName":                 recipient.UserName,
		"UserEmail":                recipient.UserEmail,
		"WorkspaceName":            recipient.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        title,
		"NotificationMessage":      formatWeeklyDigestEmailContent(stats),
		"NotificationType":         "weekly_digest",
		"NotificationCTAURL":       fmt.Sprintf("%s/my-work?tab=assigned", workspaceURL),
		"NotificationCTALabel":     "Plan my week",
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{recipient.UserEmail},
		Template: "notifications/notification",
		Subject:  title,
		Data:     data,
	}); err != nil {
		return fmt.Errorf("failed to send weekly digest email: %w", err)
	}

	log.Info(ctx, "Successfully sent weekly digest email",
		"user_id", recipient.UserID,
		"workspace_id", recipient.WorkspaceID,
		"user_email", recipient.UserEmail)
	return nil
}

func formatWeeklyDigestEmailContent(stats WeeklyDigestStats) string {
	textStyle := mailer.EmailStyleString("notificationText")
	listStyle := mailer.EmailStyleString("notificationList")
	itemStyle := mailer.EmailStyleString("notificationItem")
	detailStyle := mailer.EmailStyleString("detailValue")

	items := make([]string, 0, 5)
	if stats.UnreadNotifications > 0 {
		label := pluralize(stats.UnreadNotifications, "unread update", "unread updates")
		text := fmt.Sprintf("%d %s", stats.UnreadNotifications, label)
		if stats.UnreadPriorityNotifications > 0 {
			priorityLabel := pluralize(stats.UnreadPriorityNotifications, "mention or reply", "mentions or replies")
			text = fmt.Sprintf("%s, including %d %s", text, stats.UnreadPriorityNotifications, priorityLabel)
		}
		items = append(items, fmt.Sprintf(`<li style="%s"><strong style="%s">%s</strong></li>`, itemStyle, detailStyle, text))
	}
	if stats.OverdueStories > 0 {
		items = append(items, fmt.Sprintf(`<li style="%s"><strong style="%s">%d %s</strong></li>`, itemStyle, detailStyle, stats.OverdueStories, pluralize(stats.OverdueStories, "overdue assigned task", "overdue assigned tasks")))
	}
	if stats.DueThisWeekStories > 0 {
		items = append(items, fmt.Sprintf(`<li style="%s"><strong style="%s">%d %s</strong></li>`, itemStyle, detailStyle, stats.DueThisWeekStories, pluralize(stats.DueThisWeekStories, "assigned task due this week", "assigned tasks due this week")))
	}
	if stats.ObjectiveRisks > 0 {
		items = append(items, fmt.Sprintf(`<li style="%s"><strong style="%s">%d %s</strong></li>`, itemStyle, detailStyle, stats.ObjectiveRisks, pluralize(stats.ObjectiveRisks, "objective needs attention", "objectives need attention")))
	}
	if stats.TeamComments > 0 {
		items = append(items, fmt.Sprintf(`<li style="%s"><strong style="%s">%d %s</strong></li>`, itemStyle, detailStyle, stats.TeamComments, pluralize(stats.TeamComments, "new team comment", "new team comments")))
	}

	if len(items) == 0 {
		return fmt.Sprintf(`<div style="%s"><p style="%s">No major updates need your attention this week.</p></div>`, textStyle, textStyle)
	}

	return fmt.Sprintf(`
		<div style="%s">
			<p style="%s">Here is what needs attention this week:</p>
			<ul style="%s">%s</ul>
		</div>
	`, textStyle, textStyle, listStyle, strings.Join(items, ""))
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
