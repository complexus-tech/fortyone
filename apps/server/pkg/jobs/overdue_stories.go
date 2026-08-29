package jobs

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const overdueStoryAssigneeBatchSize = 100

// OverdueStoryStore is the worker-owned persistence capability needed to page
// eligible recipients and load their current story deadline signals.
type OverdueStoryStore interface {
	ListOverdueStoryGuidanceRecipients(context.Context, time.Time, *storydomain.OverdueGuidanceCursor, int) ([]storydomain.OverdueGuidanceRecipient, error)
	ListOverdueStoryGuidanceItems(context.Context, time.Time, uuid.UUID, uuid.UUID) ([]storydomain.OverdueGuidanceStory, error)
}

// ProcessOverdueStoriesEmail processes overdue stories and sends emails directly.
func ProcessOverdueStoriesEmail(ctx context.Context, store OverdueStoryStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer) error {
	return processOverdueStoriesEmailAt(ctx, store, log, mailerService, copyGenerator, threader, time.Now().UTC())
}

func processOverdueStoriesEmailAt(ctx context.Context, store OverdueStoryStore, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, asOf time.Time) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessOverdueStoriesEmail")
	defer span.End()
	if store == nil {
		return errors.New("overdue story store is required")
	}
	if log == nil {
		return errors.New("overdue story logger is required")
	}
	if mailerService == nil {
		return errors.New("overdue story mailer is required")
	}
	asOf, err := overdueGuidanceUTCDate(asOf)
	if err != nil {
		return err
	}

	log.Info(ctx, "Processing overdue stories email notifications")
	startTime := time.Now()

	totalProcessed := 0
	totalEmailsCreated := 0
	batchCount := 0
	var cursor *storydomain.OverdueGuidanceCursor

	for {
		nextBatch := batchCount + 1
		log.Info(ctx, fmt.Sprintf("Processing assignee batch %d", nextBatch))

		// Get next batch of assignees with overdue stories (filtered by email preferences)
		assignees, err := store.ListOverdueStoryGuidanceRecipients(ctx, asOf, cursor, overdueStoryAssigneeBatchSize)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get assignees batch %d: %w", nextBatch, err)
		}

		if len(assignees) == 0 {
			break // No more assignees
		}
		batchCount = nextBatch

		results, batchErr := processGuidanceEmailBatch(ctx, assignees, func(batchCtx context.Context, assignee storydomain.OverdueGuidanceRecipient) guidanceEmailBatchResult {
			return processGuidanceEmailRecipient(batchCtx, func(attemptCtx context.Context) guidanceEmailBatchResult {
				stories, storiesErr := store.ListOverdueStoryGuidanceItems(attemptCtx, asOf, assignee.AssigneeID, assignee.WorkspaceID)
				if storiesErr != nil {
					log.Error(attemptCtx, "Failed to get stories for assignee", "assignee_id", assignee.AssigneeID, "workspace_id", assignee.WorkspaceID, "error", storiesErr)
					return guidanceEmailBatchResult{Err: storiesErr, Retryable: true}
				}
				if len(stories) == 0 {
					return guidanceEmailBatchResult{Processed: true}
				}
				if sendErr := sendOverdueStoriesEmailForAssignee(attemptCtx, log, mailerService, copyGenerator, threader, stories); sendErr != nil {
					log.Error(attemptCtx, "Failed to send email", "assignee_id", assignee.AssigneeID, "error", sendErr)
					return guidanceEmailBatchResult{Err: sendErr}
				}
				return guidanceEmailBatchResult{Processed: true, Sent: true}
			})
		})
		if batchErr != nil {
			span.RecordError(batchErr)
			return fmt.Errorf("overdue task guidance batch cancelled: %w", batchErr)
		}
		for _, result := range results {
			if result.Processed {
				totalProcessed++
			}
			if result.Sent {
				totalEmailsCreated++
			}
		}
		if failureCount := guidanceEmailBatchFailureCount(results); failureCount > 0 {
			log.Error(ctx, "Overdue task recipients failed after in-job processing; continuing without retrying successful deliveries", "failed_recipients", failureCount, "batch", batchCount)
			span.AddEvent("overdue task recipient deliveries failed", trace.WithAttributes(attribute.Int("failed_recipients", failureCount)))
		}

		log.Info(ctx, fmt.Sprintf("Assignee batch %d completed: %d assignees processed", batchCount, len(assignees)))

		lastAssignee := assignees[len(assignees)-1]
		cursor = &storydomain.OverdueGuidanceCursor{
			AssigneeID:  lastAssignee.AssigneeID,
			WorkspaceID: lastAssignee.WorkspaceID,
		}
		if len(assignees) < overdueStoryAssigneeBatchSize {
			break
		}
		if err := waitForNextOverdueGuidanceBatch(ctx); err != nil {
			span.RecordError(err)
			return fmt.Errorf("wait before overdue story assignee batch %d: %w", batchCount+1, err)
		}
	}

	duration := time.Since(startTime)

	span.AddEvent("overdue stories email job completed", trace.WithAttributes(
		attribute.Int64("assignees.processed", int64(totalProcessed)),
		attribute.Int64("emails.created", int64(totalEmailsCreated)),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Overdue stories email job completed: %d assignees processed, %d emails created in %d batches over %v",
		totalProcessed, totalEmailsCreated, batchCount, duration))

	return nil
}

type OverdueStory = storydomain.OverdueGuidanceStory

// sendOverdueStoriesEmailForAssignee sends email directly for a specific assignee
func sendOverdueStoriesEmailForAssignee(ctx context.Context, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, stories []OverdueStory) error {
	ctx, span := web.AddSpan(ctx, "jobs.sendOverdueStoriesEmailForAssignee")
	defer span.End()

	if len(stories) == 0 {
		return nil
	}

	if strings.TrimSpace(stories[0].AssigneeEmail) == "" {
		log.Info(ctx, "Skipping overdue stories email because assignee email is empty", "assignee_id", stories[0].AssigneeID)
		return nil
	}

	// Group stories by deadline status
	var dueSoonStories, dueTodayStories, overdueStories []OverdueStory

	for _, story := range stories {
		switch story.DeadlineStatus {
		case "due_in_3_days", "due_tomorrow":
			dueSoonStories = append(dueSoonStories, story)
		case "due_today":
			dueTodayStories = append(dueTodayStories, story)
		case "overdue":
			overdueStories = append(overdueStories, story)
		}
	}

	// Use data from first story for common fields
	firstStory := stories[0]
	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", firstStory.WorkspaceSlug)

	totalCount := len(dueSoonStories) + len(dueTodayStories) + len(overdueStories)
	itemText := "task"
	if totalCount > 1 {
		itemText = "tasks"
	}
	title := fmt.Sprintf("%d %s need attention", totalCount, itemText)
	heading := title
	emailContent := formatOverdueStoriesEmailContent(firstStory, dueSoonStories, dueTodayStories, overdueStories, workspaceURL)
	emailContent = appendGuidanceReplyPrompt(emailContent, "I’m Maya, your AI agent. Reply to this email with the new due date, status, or owner. I’ll show you the exact change before anything is updated.")
	ctaURL := fmt.Sprintf("%s/my-work?tab=assigned", workspaceURL)
	ctaLabel := "View my work"

	if copyGenerator != nil {
		request, destinations := overdueStoriesEmailCopyRequest(stories, workspaceURL, ctaURL)
		generated, err := copyGenerator.Generate(ctx, request)
		if err != nil {
			log.Warn(ctx, "Falling back to deterministic task guidance copy", "assignee_id", firstStory.AssigneeID, "workspace_id", firstStory.WorkspaceID, "error", err)
		} else if generatedContent, renderErr := renderGeneratedEmailContent(generated, destinations); renderErr != nil {
			log.Warn(ctx, "Falling back to deterministic task guidance copy after render validation", "assignee_id", firstStory.AssigneeID, "workspace_id", firstStory.WorkspaceID, "error", renderErr)
		} else if generatedCTALabel, generatedCTAURL, ok := generatedPrimaryCTA(generated, destinations); !ok {
			log.Warn(ctx, "Falling back to deterministic task guidance copy because no trusted CTA was generated", "assignee_id", firstStory.AssigneeID, "workspace_id", firstStory.WorkspaceID)
		} else {
			title = generated.Subject.Text
			heading = generated.H1.Text
			emailContent = generatedContent
			ctaLabel = generatedCTALabel
			ctaURL = generatedCTAURL
		}
	}

	data := map[string]any{
		"UserName":                 firstStory.AssigneeName,
		"UserEmail":                firstStory.AssigneeEmail,
		"WorkspaceName":            firstStory.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        heading,
		"NotificationMessage":      emailContent,
		"NotificationType":         "reminders",
		"NotificationCTAURL":       ctaURL,
		"NotificationCTALabel":     ctaLabel,
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}
	messageID := guidanceEmailMessageID("task-guidance", firstStory.WorkspaceID, firstStory.AssigneeID, time.Now())
	targets := make([]emailthread.TargetContext, 0, len(stories))
	for _, story := range stories {
		targets = append(targets, emailthread.TargetContext{
			Kind:        "story",
			ID:          story.ID,
			TeamID:      story.TeamID,
			DisplayName: story.Title,
		})
	}
	threadContext, err := emailthread.EncodeThreadContext(emailthread.ThreadContext{
		Source:        "task_guidance",
		WorkspaceSlug: firstStory.WorkspaceSlug,
		Targets:       targets,
	})
	if err != nil {
		return err
	}
	plainText := guidancePlainText(heading, emailContent, ctaLabel, ctaURL)
	replyTo, err := prepareGuidanceThread(ctx, threader, emailthread.GuidanceInput{
		WorkspaceID:       firstStory.WorkspaceID,
		UserID:            firstStory.AssigneeID,
		RecipientEmail:    firstStory.AssigneeEmail,
		ExternalThreadID:  messageID,
		InternetMessageID: messageID,
		Subject:           title,
		Content:           plainText,
		Context:           threadContext,
	})
	if err != nil {
		return fmt.Errorf("prepare task guidance reply thread: %w", err)
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{firstStory.AssigneeEmail},
		Template:      "notifications/notification",
		Subject:       title,
		Data:          data,
		PlainTextBody: plainText,
		Sender:        mailer.SenderProfileMaya,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to send overdue stories email: %w", err)
	}

	log.Info(ctx, "Successfully sent overdue stories email",
		"assignee_id", firstStory.AssigneeID,
		"assignee_email", firstStory.AssigneeEmail,
		"workspace_name", firstStory.WorkspaceName,
		"total_stories", totalCount)

	span.AddEvent("email prepared", trace.WithAttributes(
		attribute.String("assignee_id", firstStory.AssigneeID.String()),
		attribute.String("assignee_email", firstStory.AssigneeEmail),
		attribute.String("workspace_name", firstStory.WorkspaceName),
		attribute.Int("stories.count", len(stories)),
	))

	return nil
}

func overdueStoriesEmailCopyRequest(stories []OverdueStory, workspaceURL, ctaURL string) (emailcopy.Request, map[string]emailCopyDestination) {
	firstStory := stories[0]
	orderedStories := make([]OverdueStory, 0, len(stories))
	for _, deadlineStatus := range []string{"overdue", "due_today", "due_tomorrow", "due_in_3_days"} {
		for _, story := range stories {
			if story.DeadlineStatus == deadlineStatus {
				orderedStories = append(orderedStories, story)
			}
		}
	}

	itemLimit := maxGuidanceEmailRows - 1
	if len(orderedStories) < itemLimit {
		itemLimit = len(orderedStories)
	}
	hiddenCount := len(orderedStories) - itemLimit
	summaryText := fmt.Sprintf("There are %d assigned tasks that need attention.", len(orderedStories))
	summaryTokens := nonEmptyFactTokens(fmt.Sprintf("%d assigned tasks that need attention", len(orderedStories)))
	if hiddenCount > 0 {
		summaryText = fmt.Sprintf("There are %d assigned tasks that need attention. This email includes %d of them; %d more %s available in assigned work.", len(orderedStories), itemLimit, hiddenCount, pluralize(hiddenCount, "is", "are"))
		summaryTokens = append(summaryTokens,
			fmt.Sprintf("includes %d of them", itemLimit),
			fmt.Sprintf("%d more %s available in assigned work", hiddenCount, pluralize(hiddenCount, "is", "are")),
		)
	}

	facts := []emailcopy.Fact{
		{
			ReferenceID:  "workspace_context",
			Text:         fmt.Sprintf("The workspace is named %s.", firstStory.WorkspaceName),
			EntityTokens: []string{firstStory.WorkspaceName},
		},
		{ReferenceID: "task_summary", Text: summaryText, ProtectedTokens: summaryTokens, Required: true},
	}
	destinations := make(map[string]emailCopyDestination, itemLimit+1)
	for _, story := range orderedStories[:itemLimit] {
		referenceID := "task:" + story.ID.String()
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     referenceID,
			Text:            overdueStoryEmailCopyFact(story),
			EntityTokens:    []string{story.Title},
			ProtectedTokens: overdueStoryProtectedTokens(story),
			Required:        true,
		})
		destinations[referenceID] = emailCopyDestination{
			Label: story.Title,
			URL:   overdueStoryURL(workspaceURL, story),
		}
	}

	const actionReference = "review_assigned_work"
	destinations[actionReference] = emailCopyDestination{Label: "Assigned work", URL: ctaURL}
	return emailcopy.Request{
		SafetyIdentifier: firstStory.AssigneeID.String(),
		Purpose:          "Help the recipient protect commitments by reviewing assigned tasks with approaching or missed due dates.",
		ProductVoice:     mayaGuidanceProductVoice,
		Facts:            facts,
		Actions: []emailcopy.Action{
			{ReferenceID: actionReference, Description: "Open assigned work to review and update these tasks."},
		},
		IncludeReplyPrompt: true,
	}, destinations
}

func overdueStoryEmailCopyFact(story OverdueStory) string {
	taskContext := fmt.Sprintf("The task %s", story.Title)
	if teamName := strings.TrimSpace(story.TeamName); teamName != "" {
		taskContext += fmt.Sprintf(" in the %s team", teamName)
	}
	if statusName := strings.TrimSpace(story.StatusName); statusName != "" {
		taskContext += fmt.Sprintf(" has status %s and", statusName)
	}

	switch story.DeadlineStatus {
	case "overdue":
		return fmt.Sprintf("%s is %d %s overdue; its due date is %s.", taskContext, story.DaysDifference, pluralize(story.DaysDifference, "day", "days"), story.EndDate.Format("January 2, 2006"))
	case "due_today":
		return fmt.Sprintf("%s is due today, %s.", taskContext, story.EndDate.Format("January 2, 2006"))
	case "due_tomorrow":
		return fmt.Sprintf("%s is due tomorrow, %s.", taskContext, story.EndDate.Format("January 2, 2006"))
	default:
		return fmt.Sprintf("%s is due on %s.", taskContext, story.EndDate.Format("January 2, 2006"))
	}
}

func overdueStoryProtectedTokens(story OverdueStory) []string {
	tokens := deadlineSemanticFactTokens(
		story.DeadlineStatus,
		story.DaysDifference,
		story.EndDate.Format("January 2, 2006"),
	)
	if teamName := strings.TrimSpace(story.TeamName); teamName != "" {
		tokens = append(tokens, "in the "+teamName+" team")
	}
	if statusName := strings.TrimSpace(story.StatusName); statusName != "" {
		tokens = append(tokens, "has status "+statusName)
	}
	return nonEmptyFactTokens(tokens...)
}

// formatOverdueStoriesEmailContent formats the email content
func formatOverdueStoriesEmailContent(firstStory OverdueStory, dueSoonStories, dueTodayStories, overdueStories []OverdueStory, workspaceURL string) string {
	totalItems := len(dueSoonStories) + len(dueTodayStories) + len(overdueStories)
	itemText := "task"
	if totalItems > 1 {
		itemText = "tasks"
	}

	detailRows := make([]string, 0, totalItems)

	if len(dueSoonStories) > 0 {
		for _, story := range dueSoonStories {
			detailRows = append(detailRows, fmt.Sprintf(
				"Task %s is due %s.",
				formatEmailLink(overdueStoryURL(workspaceURL, story), story.Title),
				html.EscapeString(story.EndDate.Format("January 2, 2006")),
			))
		}
	}

	if len(dueTodayStories) > 0 {
		for _, story := range dueTodayStories {
			detailRows = append(detailRows, fmt.Sprintf(
				"Task %s is due today.",
				formatEmailLink(overdueStoryURL(workspaceURL, story), story.Title),
			))
		}
	}

	if len(overdueStories) > 0 {
		for _, story := range overdueStories {
			daysText := "day"
			if story.DaysDifference > 1 {
				daysText = "days"
			}
			detailRows = append(detailRows, fmt.Sprintf(
				"Task %s is %s overdue.",
				formatEmailLink(overdueStoryURL(workspaceURL, story), story.Title),
				formatEmailStrong(fmt.Sprintf("%d %s", story.DaysDifference, daysText)),
			))
		}
	}

	visibleRows, hiddenCount := capGuidanceEmailDetailRows(detailRows)
	summary := fmt.Sprintf("You have %s that need attention.", formatEmailStrong(fmt.Sprintf("%d %s", totalItems, itemText)))
	if hiddenCount > 0 {
		summary += fmt.Sprintf(
			" This email includes %d of them; %d more %s available in assigned work.",
			len(visibleRows),
			hiddenCount,
			pluralize(hiddenCount, "is", "are"),
		)
	}
	rows := append([]string{summary}, visibleRows...)

	return formatCompactNotificationRows("Here's what needs attention.", rows)
}

func overdueStoryURL(workspaceURL string, story OverdueStory) string {
	reference := story.ID.String()
	if teamCode := strings.ToUpper(strings.TrimSpace(story.TeamCode)); teamCode != "" && story.SequenceID > 0 {
		reference = fmt.Sprintf("%s-%d", teamCode, story.SequenceID)
	}
	return fmt.Sprintf("%s/work/%s", strings.TrimRight(workspaceURL, "/"), reference)
}
