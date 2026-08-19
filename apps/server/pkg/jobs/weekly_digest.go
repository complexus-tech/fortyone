package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	mayaGuidanceProductVoice = "Calm, concise product guidance that helps the recipient decide what to advance next. Avoid marketing language and notification-system jargon."
	maxGuidanceEmailRows     = 12
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
	return s.UnreadPriorityNotifications+
		s.OverdueStories+
		s.DueThisWeekStories+
		s.ObjectiveRisks > 0
}

// ProcessWeeklyDigestEmail sends a weekly workspace digest to users with meaningful activity.
func ProcessWeeklyDigestEmail(ctx context.Context, db *sqlx.DB, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer) error {
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

		results, batchErr := processGuidanceEmailBatch(ctx, recipients, func(batchCtx context.Context, recipient WeeklyDigestRecipient) guidanceEmailBatchResult {
			return processGuidanceEmailRecipient(batchCtx, func(attemptCtx context.Context) guidanceEmailBatchResult {
				stats, statsErr := getWeeklyDigestStats(attemptCtx, db, recipient.UserID, recipient.WorkspaceID)
				if statsErr != nil {
					log.Error(attemptCtx, "Failed to get weekly digest stats", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", statsErr)
					return guidanceEmailBatchResult{Err: statsErr, Retryable: true}
				}
				if !stats.hasSignal() {
					return guidanceEmailBatchResult{Processed: true}
				}
				if sendErr := sendWeeklyDigestEmail(attemptCtx, log, mailerService, copyGenerator, threader, recipient, stats); sendErr != nil {
					log.Error(attemptCtx, "Failed to send weekly digest email", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", sendErr)
					return guidanceEmailBatchResult{Err: sendErr}
				}
				return guidanceEmailBatchResult{Processed: true, Sent: true}
			})
		})
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("weekly digest batch cancelled: %w", batchErr)
		}
		for _, result := range results {
			if result.Processed {
				totalProcessed++
			}
			if result.Sent {
				totalEmailsSent++
			}
		}
		if failureCount := guidanceEmailBatchFailureCount(results); failureCount > 0 {
			log.Error(ctx, "Weekly digest recipients failed after in-job processing; continuing without retrying successful deliveries", "failed_recipients", failureCount, "batch", batchCount)
			span.AddEvent("weekly digest recipient deliveries failed", trace.WithAttributes(attribute.Int("failed_recipients", failureCount)))
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

	query := weeklyDigestRecipientsQuery()

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

func weeklyDigestRecipientsQuery() string {
	return `
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
			AND wm.role IN ('admin', 'member')
			AND w.deleted_at IS NULL
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND CAST(COALESCE(np.preferences -> 'weekly_digest' ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY w.workspace_id, wm.user_id
		LIMIT :batch_size OFFSET :offset;
	`
}

func getWeeklyDigestStats(ctx context.Context, db *sqlx.DB, userID, workspaceID uuid.UUID) (WeeklyDigestStats, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getWeeklyDigestStats")
	defer span.End()

	query := weeklyDigestStatsQuery()

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

func weeklyDigestStatsQuery() string {
	return `
		WITH recipient_access AS (
			SELECT wm.role
			FROM workspace_members wm
			INNER JOIN workspaces w ON w.workspace_id = wm.workspace_id
			WHERE wm.user_id = :user_id
				AND wm.workspace_id = :workspace_id
				AND wm.role IN ('admin', 'member')
				AND w.deleted_at IS NULL
		),
		visible_teams AS (
			SELECT tm.team_id
			FROM team_members tm
			WHERE tm.user_id = :user_id
		),
		accessible_notifications AS (
			SELECT n.notification_id, n.type, n.created_at
			FROM notifications n
			WHERE n.recipient_id = :user_id
				AND n.workspace_id = :workspace_id
				AND n.read_at IS NULL
				AND (
					(
						CAST(n.entity_type AS TEXT) = 'feedback'
						AND EXISTS (
							SELECT 1
							FROM feedback_items feedback
							WHERE feedback.id = n.entity_id
								AND feedback.workspace_id = n.workspace_id
								AND feedback.deleted_at IS NULL
						)
					)
					OR (
						EXISTS (SELECT 1 FROM recipient_access)
						AND (
							(
								CAST(n.entity_type AS TEXT) = 'story'
								AND EXISTS (
									SELECT 1
									FROM stories story
									WHERE story.id = n.entity_id
										AND story.workspace_id = n.workspace_id
										AND story.deleted_at IS NULL
										AND (
											EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
											OR story.team_id IN (SELECT team_id FROM visible_teams)
										)
								)
							)
							OR (
								CAST(n.entity_type AS TEXT) = 'comment'
								AND EXISTS (
									SELECT 1
									FROM story_comments comment
									INNER JOIN stories story ON story.id = comment.story_id
									WHERE comment.comment_id = n.entity_id
										AND story.workspace_id = n.workspace_id
										AND story.deleted_at IS NULL
										AND (
											EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
											OR story.team_id IN (SELECT team_id FROM visible_teams)
										)
								)
							)
							OR (
								CAST(n.entity_type AS TEXT) = 'objective'
								AND EXISTS (
									SELECT 1
									FROM objectives objective
									WHERE objective.objective_id = n.entity_id
										AND objective.workspace_id = n.workspace_id
										AND (
											EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
											OR objective.team_id IN (SELECT team_id FROM visible_teams)
										)
								)
							)
							OR (
								CAST(n.entity_type AS TEXT) = 'key_result'
								AND EXISTS (
									SELECT 1
									FROM key_results key_result
									INNER JOIN objectives objective ON objective.objective_id = key_result.objective_id
									WHERE key_result.id = n.entity_id
										AND objective.workspace_id = n.workspace_id
										AND (
											EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
											OR objective.team_id IN (SELECT team_id FROM visible_teams)
										)
								)
							)
							OR (
								CAST(n.entity_type AS TEXT) = 'strategy'
								AND (
									EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
									OR n.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
								)
							)
						)
					)
				)
		)
		SELECT
			CAST((
				SELECT COUNT(*)
				FROM accessible_notifications
			) AS int) AS unread_notifications,
			CAST((
				SELECT COUNT(*)
				FROM accessible_notifications notification
				WHERE notification.created_at >= NOW() - INTERVAL '7 days'
					AND notification.type IN ('mention', 'comment_reply')
			) AS int) AS unread_priority_notifications,
			CAST((
				SELECT COUNT(*)
				FROM stories s
				INNER JOIN statuses st ON s.status_id = st.status_id
				WHERE s.assignee_id = :user_id
					AND s.workspace_id = :workspace_id
					AND EXISTS (SELECT 1 FROM recipient_access)
					AND s.end_date < CURRENT_DATE
					AND st.category NOT IN ('completed', 'cancelled', 'paused')
					AND s.deleted_at IS NULL
					AND s.archived_at IS NULL
					AND s.completed_at IS NULL
					AND (
						EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
						OR s.team_id IN (SELECT team_id FROM visible_teams)
					)
			) AS int) AS overdue_stories,
			CAST((
				SELECT COUNT(*)
				FROM stories s
				INNER JOIN statuses st ON s.status_id = st.status_id
				WHERE s.assignee_id = :user_id
					AND s.workspace_id = :workspace_id
					AND EXISTS (SELECT 1 FROM recipient_access)
					AND s.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '7 days'
					AND st.category NOT IN ('completed', 'cancelled', 'paused')
					AND s.deleted_at IS NULL
					AND s.archived_at IS NULL
					AND s.completed_at IS NULL
					AND (
						EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
						OR s.team_id IN (SELECT team_id FROM visible_teams)
					)
			) AS int) AS due_this_week_stories,
			CAST((
				SELECT COUNT(*)
				FROM objectives o
				INNER JOIN objective_statuses os ON o.status_id = os.status_id
				INNER JOIN workspace_settings ws ON o.workspace_id = ws.workspace_id
				WHERE o.lead_user_id = :user_id
					AND o.workspace_id = :workspace_id
					AND EXISTS (SELECT 1 FROM recipient_access)
					AND o.end_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE + INTERVAL '7 days'
					AND os.category NOT IN ('completed', 'cancelled', 'paused')
					AND ws.objective_enabled = true
					AND (
						EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
						OR o.team_id IN (SELECT team_id FROM visible_teams)
					)
			) AS int) AS objective_risks,
			CAST((
				SELECT COUNT(*)
				FROM story_comments sc
				INNER JOIN stories s ON sc.story_id = s.id
				WHERE s.workspace_id = :workspace_id
					AND EXISTS (SELECT 1 FROM recipient_access)
					AND sc.commenter_id <> :user_id
					AND sc.created_at >= NOW() - INTERVAL '7 days'
					AND s.deleted_at IS NULL
					AND (
						EXISTS (SELECT 1 FROM recipient_access WHERE role = 'admin')
						OR s.team_id IN (SELECT team_id FROM visible_teams)
					)
			) AS int) AS team_comments;
	`
}

func sendWeeklyDigestEmail(ctx context.Context, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, recipient WeeklyDigestRecipient, stats WeeklyDigestStats) error {
	if strings.TrimSpace(recipient.UserEmail) == "" {
		log.Info(ctx, "Skipping weekly digest because user email is empty", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID)
		return nil
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", recipient.WorkspaceSlug)
	title := fmt.Sprintf("Weekly digest: %s", recipient.WorkspaceName)
	heading := title
	emailContent := formatWeeklyDigestEmailContent(stats)
	emailContent = appendGuidanceReplyPrompt(emailContent, "I’m Maya, your AI agent. Reply to this email with what has changed since your last update or where you’d like help deciding the next step.")
	ctaURL := fmt.Sprintf("%s/my-work?tab=assigned", workspaceURL)
	ctaLabel := "Plan my week"

	if copyGenerator != nil {
		request, destinations := weeklyDigestEmailCopyRequest(recipient, stats, ctaURL)
		generated, err := copyGenerator.Generate(ctx, request)
		if err != nil {
			log.Warn(ctx, "Falling back to deterministic weekly digest copy", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", err)
		} else if generatedContent, renderErr := renderGeneratedEmailContent(generated, destinations); renderErr != nil {
			log.Warn(ctx, "Falling back to deterministic weekly digest copy after render validation", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", renderErr)
		} else if generatedCTALabel, generatedCTAURL, ok := generatedPrimaryCTA(generated, destinations); !ok {
			log.Warn(ctx, "Falling back to deterministic weekly digest copy because no trusted CTA was generated", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID)
		} else {
			title = generated.Subject.Text
			heading = generated.H1.Text
			emailContent = generatedContent
			ctaLabel = generatedCTALabel
			ctaURL = generatedCTAURL
		}
	}

	data := map[string]any{
		"UserName":                 recipient.UserName,
		"UserEmail":                recipient.UserEmail,
		"WorkspaceName":            recipient.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        heading,
		"NotificationMessage":      emailContent,
		"NotificationType":         "weekly_digest",
		"NotificationCTAURL":       ctaURL,
		"NotificationCTALabel":     ctaLabel,
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}
	messageID := guidanceEmailMessageID("weekly-digest", recipient.WorkspaceID, recipient.UserID, time.Now())
	threadContext, err := emailthread.EncodeThreadContext(emailthread.ThreadContext{
		Source:        "weekly_digest",
		WorkspaceSlug: recipient.WorkspaceSlug,
	})
	if err != nil {
		return err
	}
	plainText := guidancePlainText(heading, emailContent, ctaLabel, ctaURL)
	replyTo, err := prepareGuidanceThread(ctx, threader, emailthread.GuidanceInput{
		WorkspaceID:       recipient.WorkspaceID,
		UserID:            recipient.UserID,
		RecipientEmail:    recipient.UserEmail,
		ExternalThreadID:  messageID,
		InternetMessageID: messageID,
		Subject:           title,
		Content:           plainText,
		Context:           threadContext,
	})
	if err != nil {
		return fmt.Errorf("prepare weekly digest reply thread: %w", err)
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{recipient.UserEmail},
		Template:      "notifications/notification",
		Subject:       title,
		Data:          data,
		PlainTextBody: plainText,
		Sender:        mailer.SenderProfileMaya,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		return fmt.Errorf("failed to send weekly digest email: %w", err)
	}

	log.Info(ctx, "Successfully sent weekly digest email",
		"user_id", recipient.UserID,
		"workspace_id", recipient.WorkspaceID,
		"user_email", recipient.UserEmail)
	return nil
}

func weeklyDigestEmailCopyRequest(recipient WeeklyDigestRecipient, stats WeeklyDigestStats, ctaURL string) (emailcopy.Request, map[string]emailCopyDestination) {
	facts := []emailcopy.Fact{
		{
			ReferenceID:  "workspace_context",
			Text:         fmt.Sprintf("The workspace is named %s.", recipient.WorkspaceName),
			EntityTokens: []string{recipient.WorkspaceName},
		},
	}
	if stats.UnreadNotifications > 0 {
		factText := fmt.Sprintf("There are %d unread updates.", stats.UnreadNotifications)
		protectedTokens := nonEmptyFactTokens(fmt.Sprintf("%d unread updates", stats.UnreadNotifications))
		if stats.UnreadPriorityNotifications > 0 {
			priorityLabel := pluralize(stats.UnreadPriorityNotifications, "mention or reply", "mentions or replies")
			factText = fmt.Sprintf("There are %d unread updates, including %d %s.", stats.UnreadNotifications, stats.UnreadPriorityNotifications, priorityLabel)
			protectedTokens = append(protectedTokens, fmt.Sprintf("including %d %s", stats.UnreadPriorityNotifications, priorityLabel))
		}
		facts = append(facts, emailcopy.Fact{ReferenceID: "unread_updates", Text: factText, ProtectedTokens: protectedTokens, Required: true})
	}
	if stats.OverdueStories > 0 {
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     "overdue_tasks",
			Text:            fmt.Sprintf("There are %d overdue assigned tasks.", stats.OverdueStories),
			ProtectedTokens: nonEmptyFactTokens(fmt.Sprintf("%d overdue assigned tasks", stats.OverdueStories)),
			Required:        true,
		})
	}
	if stats.DueThisWeekStories > 0 {
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     "tasks_due_this_week",
			Text:            fmt.Sprintf("There are %d assigned tasks due this week.", stats.DueThisWeekStories),
			ProtectedTokens: nonEmptyFactTokens(fmt.Sprintf("%d assigned tasks due this week", stats.DueThisWeekStories)),
			Required:        true,
		})
	}
	if stats.ObjectiveRisks > 0 {
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     "objective_risks",
			Text:            fmt.Sprintf("There are %d objectives that need attention.", stats.ObjectiveRisks),
			ProtectedTokens: nonEmptyFactTokens(fmt.Sprintf("%d objectives that need attention", stats.ObjectiveRisks)),
			Required:        true,
		})
	}
	if stats.TeamComments > 0 {
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     "team_comments",
			Text:            fmt.Sprintf("There are %d new team comments.", stats.TeamComments),
			ProtectedTokens: nonEmptyFactTokens(fmt.Sprintf("%d new team comments", stats.TeamComments)),
			Required:        true,
		})
	}

	const actionReference = "review_week"
	request := emailcopy.Request{
		SafetyIdentifier: recipient.UserID.String(),
		Purpose:          "Help the recipient review the week's signals and choose what to advance next.",
		ProductVoice:     mayaGuidanceProductVoice,
		Facts:            facts,
		Actions: []emailcopy.Action{
			{ReferenceID: actionReference, Description: "Open the recipient's assigned work to plan the week."},
		},
		IncludeReplyPrompt: true,
	}
	destinations := map[string]emailCopyDestination{
		actionReference: {Label: "Assigned work", URL: ctaURL},
	}
	return request, destinations
}

func formatWeeklyDigestEmailContent(stats WeeklyDigestStats) string {
	textStyle := mailer.EmailStyleString("notificationText")

	items := make([]string, 0, 5)
	if stats.UnreadNotifications > 0 {
		label := pluralize(stats.UnreadNotifications, "unread update", "unread updates")
		text := fmt.Sprintf("%d %s", stats.UnreadNotifications, label)
		if stats.UnreadPriorityNotifications > 0 {
			priorityLabel := pluralize(stats.UnreadPriorityNotifications, "mention or reply", "mentions or replies")
			text = fmt.Sprintf("%s, including %d %s", text, stats.UnreadPriorityNotifications, priorityLabel)
		}
		items = append(items, fmt.Sprintf("You have %s.", formatEmailStrong(text)))
	}
	if stats.OverdueStories > 0 {
		text := fmt.Sprintf("%d %s", stats.OverdueStories, pluralize(stats.OverdueStories, "overdue assigned task", "overdue assigned tasks"))
		items = append(items, fmt.Sprintf("You have %s.", formatEmailStrong(text)))
	}
	if stats.DueThisWeekStories > 0 {
		text := fmt.Sprintf("%d %s", stats.DueThisWeekStories, pluralize(stats.DueThisWeekStories, "assigned task due this week", "assigned tasks due this week"))
		items = append(items, fmt.Sprintf("You have %s.", formatEmailStrong(text)))
	}
	if stats.ObjectiveRisks > 0 {
		text := fmt.Sprintf("%d %s", stats.ObjectiveRisks, pluralize(stats.ObjectiveRisks, "objective needs attention", "objectives need attention"))
		items = append(items, fmt.Sprintf("You have %s.", formatEmailStrong(text)))
	}
	if stats.TeamComments > 0 {
		text := fmt.Sprintf("%d %s", stats.TeamComments, pluralize(stats.TeamComments, "new team comment", "new team comments"))
		items = append(items, fmt.Sprintf("You have %s.", formatEmailStrong(text)))
	}

	if len(items) == 0 {
		return fmt.Sprintf(`<div style="%s"><p style="%s">No major updates need your attention this week.</p></div>`, textStyle, textStyle)
	}

	return formatCompactNotificationRows("Here is what needs attention this week:", items)
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
