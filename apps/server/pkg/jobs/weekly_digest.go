package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	mayaGuidanceProductVoice = "Calm, concise product guidance that helps the recipient decide what to advance next. Avoid marketing language and notification-system jargon."
	maxGuidanceEmailRows     = mailer.DigestDetailLimit + 1
	weeklyDigestBatchSize    = 100
	weeklyDigestBatchDelay   = 100 * time.Millisecond
)

type WeeklyDigestRecipient = notificationsdomain.WeeklyDigestRecipient
type WeeklyDigestStats = notificationsdomain.WeeklyDigestStats

// WeeklyDigestStore is the worker-owned persistence capability needed to page
// current recipients and load one scoped aggregate for each recipient.
type WeeklyDigestStore interface {
	ListWeeklyDigestRecipients(context.Context, *notificationsdomain.WeeklyDigestCursor, int) ([]notificationsdomain.WeeklyDigestRecipient, error)
	GetWeeklyDigestStats(context.Context, notificationsdomain.WeeklyDigestStatsQuery) (notificationsdomain.WeeklyDigestStats, error)
}

func hasWeeklyDigestSignal(stats WeeklyDigestStats) bool {
	return stats.UnreadPriorityNotifications+
		stats.OverdueStories+
		stats.DueThisWeekStories+
		stats.ObjectiveRisks > 0
}

// ProcessWeeklyDigestEmail sends a weekly workspace digest to users with meaningful activity.
func ProcessWeeklyDigestEmail(ctx context.Context, store WeeklyDigestStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer) error {
	return processWeeklyDigestEmailAt(ctx, store, log, mailerService, copyGenerator, threader, time.Now().UTC())
}

func processWeeklyDigestEmailAt(
	ctx context.Context,
	store WeeklyDigestStore,
	log *logger.Logger,
	mailerService mailer.Service,
	copyGenerator emailcopy.Generator,
	threader emailthread.GuidancePreparer,
	asOf time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessWeeklyDigestEmail")
	defer span.End()
	if store == nil {
		return errors.New("weekly digest store is required")
	}
	if log == nil {
		return errors.New("weekly digest logger is required")
	}
	if mailerService == nil {
		return errors.New("weekly digest mailer is required")
	}
	if asOf.IsZero() {
		return errors.New("weekly digest as-of time is required")
	}
	asOf = asOf.UTC()

	log.Info(ctx, "Processing weekly digest emails")
	startTime := time.Now()

	totalProcessed := 0
	totalEmailsSent := 0
	batchCount := 0
	var cursor *notificationsdomain.WeeklyDigestCursor

	for {
		nextBatch := batchCount + 1
		recipients, err := store.ListWeeklyDigestRecipients(ctx, cursor, weeklyDigestBatchSize)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("list weekly digest recipients batch %d: %w", nextBatch, err)
		}
		if len(recipients) == 0 {
			break
		}
		batchCount = nextBatch

		results, batchErr := processGuidanceEmailBatch(ctx, recipients, func(batchCtx context.Context, recipient WeeklyDigestRecipient) guidanceEmailBatchResult {
			return processGuidanceEmailRecipient(batchCtx, func(attemptCtx context.Context) guidanceEmailBatchResult {
				stats, statsErr := store.GetWeeklyDigestStats(attemptCtx, notificationsdomain.WeeklyDigestStatsQuery{
					UserID:      recipient.UserID,
					WorkspaceID: recipient.WorkspaceID,
					AsOf:        asOf,
				})
				if statsErr != nil {
					log.Error(attemptCtx, "Failed to get weekly digest stats", "user_id", recipient.UserID, "workspace_id", recipient.WorkspaceID, "error", statsErr)
					return guidanceEmailBatchResult{Err: statsErr, Retryable: true}
				}
				if !hasWeeklyDigestSignal(stats) {
					return guidanceEmailBatchResult{Processed: true}
				}
				if sendErr := sendWeeklyDigestEmail(attemptCtx, log, mailerService, copyGenerator, threader, recipient, stats, asOf); sendErr != nil {
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

		lastRecipient := recipients[len(recipients)-1]
		cursor = &notificationsdomain.WeeklyDigestCursor{
			WorkspaceID: lastRecipient.WorkspaceID,
			UserID:      lastRecipient.UserID,
		}
		if len(recipients) < weeklyDigestBatchSize {
			break
		}
		if err := waitForNextWeeklyDigestBatch(ctx); err != nil {
			span.RecordError(err)
			return fmt.Errorf("wait before weekly digest batch %d: %w", batchCount+1, err)
		}
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

func waitForNextWeeklyDigestBatch(ctx context.Context) error {
	timer := time.NewTimer(weeklyDigestBatchDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func sendWeeklyDigestEmail(ctx context.Context, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, recipient WeeklyDigestRecipient, stats WeeklyDigestStats, asOf time.Time) error {
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
	messageID := guidanceEmailMessageID("weekly-digest", recipient.WorkspaceID, recipient.UserID, asOf)
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
		"workspace_id", recipient.WorkspaceID)
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
