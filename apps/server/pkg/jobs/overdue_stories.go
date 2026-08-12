package jobs

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessOverdueStoriesEmail processes overdue stories and sends emails directly
func ProcessOverdueStoriesEmail(ctx context.Context, db *sqlx.DB, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessOverdueStoriesEmail")
	defer span.End()

	log.Info(ctx, "Processing overdue stories email notifications")
	startTime := time.Now()

	const assigneeBatchSize = 100 // Process 100 assignees at a time
	totalProcessed := 0
	totalEmailsCreated := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing assignee batch %d", batchCount))

		// Get next batch of assignees with overdue stories (filtered by email preferences)
		assignees, err := getAssigneesWithOverdueStories(ctx, db, assigneeBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get assignees batch %d: %w", batchCount, err)
		}

		if len(assignees) == 0 {
			break // No more assignees
		}

		results, batchErr := processGuidanceEmailBatch(ctx, assignees, func(batchCtx context.Context, assignee OverdueStory) guidanceEmailBatchResult {
			return processGuidanceEmailRecipient(batchCtx, func(attemptCtx context.Context) guidanceEmailBatchResult {
				stories, storiesErr := getOverdueStoriesForAssignee(attemptCtx, db, assignee.AssigneeID, assignee.WorkspaceID)
				if storiesErr != nil {
					log.Error(attemptCtx, "Failed to get stories for assignee", "assignee_id", assignee.AssigneeID, "workspace_id", assignee.WorkspaceID, "error", storiesErr)
					return guidanceEmailBatchResult{Err: storiesErr}
				}
				if len(stories) == 0 {
					return guidanceEmailBatchResult{Processed: true}
				}
				if sendErr := sendOverdueStoriesEmailForAssignee(attemptCtx, log, mailerService, copyGenerator, stories); sendErr != nil {
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

		// Small delay to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
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

// OverdueStory represents a story that needs attention
type OverdueStory struct {
	ID             uuid.UUID `db:"id"`
	Title          string    `db:"title"`
	EndDate        time.Time `db:"end_date"`
	AssigneeID     uuid.UUID `db:"assignee_id"`
	AssigneeEmail  string    `db:"assignee_email"`
	AssigneeName   string    `db:"assignee_name"`
	WorkspaceID    uuid.UUID `db:"workspace_id"`
	WorkspaceName  string    `db:"workspace_name"`
	WorkspaceSlug  string    `db:"workspace_slug"`
	TeamID         uuid.UUID `db:"team_id"`
	TeamName       string    `db:"team_name"`
	TeamCode       string    `db:"team_code"`
	SequenceID     int       `db:"sequence_id"`
	StatusName     string    `db:"status_name"`
	StatusCategory string    `db:"status_category"`
	DeadlineStatus string    `db:"deadline_status"`
	DaysDifference int       `db:"days_difference"`
	EmailEnabled   bool      `db:"email_enabled"`
}

// getAssigneesWithOverdueStories gets a batch of assignees who have stories needing attention and email enabled
func getAssigneesWithOverdueStories(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]OverdueStory, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getAssigneesWithOverdueStories")
	defer span.End()

	query := overdueStoryRecipientsQuery()

	params := map[string]any{
		"batch_size": batchSize,
		"offset":     offset * batchSize,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare assignees query: %w", err)
	}
	defer stmt.Close()

	var assignees []OverdueStory
	if err := stmt.SelectContext(ctx, &assignees, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute assignees query: %w", err)
	}

	span.AddEvent("assignees retrieved", trace.WithAttributes(
		attribute.Int("assignees.count", len(assignees)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))

	return assignees, nil
}

func overdueStoryRecipientsQuery() string {
	return `
		SELECT DISTINCT
			s.assignee_id,
			u.email as assignee_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) as assignee_name,
			w.workspace_id,
			w.name as workspace_name,
			w.slug as workspace_slug,
			CAST(COALESCE(np.preferences -> 'reminders' ->> 'email', 'true') AS BOOLEAN) AS email_enabled
		FROM stories s
		JOIN users u ON s.assignee_id = u.user_id
		JOIN workspaces w ON s.workspace_id = w.workspace_id
		JOIN workspace_members wm
			ON wm.workspace_id = s.workspace_id
			AND wm.user_id = s.assignee_id
			AND wm.role IN ('admin', 'member', 'guest')
		JOIN statuses st ON s.status_id = st.status_id
		LEFT JOIN notification_preferences np ON s.assignee_id = np.user_id AND s.workspace_id = np.workspace_id
		WHERE s.end_date IS NOT NULL
			AND st.category NOT IN ('completed', 'cancelled', 'paused')
			AND w.deleted_at IS NULL
			AND (
				wm.role = 'admin'
				OR EXISTS (
					SELECT 1
					FROM team_members tm
					WHERE tm.team_id = s.team_id
						AND tm.user_id = s.assignee_id
				)
			)
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.completed_at IS NULL
			AND s.assignee_id IS NOT NULL
			AND s.end_date BETWEEN CURRENT_DATE - INTERVAL '3 days' AND CURRENT_DATE + INTERVAL '3 days'
			AND u.is_active = true
			AND u.is_system = false
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND CAST(COALESCE(np.preferences -> 'reminders' ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY s.assignee_id, w.workspace_id
		LIMIT :batch_size OFFSET :offset`
}

func overdueStoriesForAssigneeQuery() string {
	return `
		WITH story_deadlines AS (
    SELECT 
        s.id, s.sequence_id, s.title, s.end_date, s.assignee_id, s.workspace_id, s.team_id,
        u.email as assignee_email, 
        COALESCE(NULLIF(u.full_name, ''), u.username) as assignee_name,
        w.name as workspace_name, w.slug as workspace_slug,
        t.name as team_name, t.code as team_code,
        st.name as status_name, st.category as status_category,
        CASE 
            WHEN s.end_date = CURRENT_DATE THEN 'due_today'
            WHEN s.end_date = CURRENT_DATE + INTERVAL '1 day' THEN 'due_tomorrow'
            WHEN s.end_date = CURRENT_DATE + INTERVAL '3 days' THEN 'due_in_3_days'
            WHEN s.end_date < CURRENT_DATE THEN 'overdue'
            ELSE 'future'
        END as deadline_status,
        CASE 
            WHEN s.end_date < CURRENT_DATE THEN CAST(CURRENT_DATE - s.end_date AS int)
            ELSE CAST(s.end_date - CURRENT_DATE AS int)
        END as days_difference
    FROM stories s
    JOIN users u ON s.assignee_id = u.user_id
    JOIN workspaces w ON s.workspace_id = w.workspace_id
	JOIN workspace_members wm
		ON wm.workspace_id = s.workspace_id
		AND wm.user_id = s.assignee_id
		AND wm.role IN ('admin', 'member', 'guest')
    JOIN teams t ON s.team_id = t.team_id
    JOIN statuses st ON s.status_id = st.status_id
    WHERE s.assignee_id = :assignee_id
        AND s.workspace_id = :workspace_id
		AND w.deleted_at IS NULL
		AND (
			wm.role = 'admin'
			OR EXISTS (
				SELECT 1
				FROM team_members tm
				WHERE tm.team_id = s.team_id
					AND tm.user_id = s.assignee_id
			)
		)
        AND s.end_date IS NOT NULL
        AND st.category NOT IN ('completed', 'cancelled', 'paused')
        AND s.deleted_at IS NULL
        AND s.archived_at IS NULL
        AND s.completed_at IS NULL
        AND s.end_date BETWEEN CURRENT_DATE - INTERVAL '3 days' AND CURRENT_DATE + INTERVAL '3 days'
        AND u.is_active = true
        AND u.is_system = false
        AND NULLIF(TRIM(u.email), '') IS NOT NULL
		)
		SELECT * 
		FROM story_deadlines 
		WHERE deadline_status IN ('due_today', 'due_tomorrow', 'due_in_3_days', 'overdue')
		ORDER BY deadline_status, end_date;
`
}

func overdueStoriesForAssigneeParams(assigneeID, workspaceID uuid.UUID) map[string]any {
	return map[string]any{
		"assignee_id":  assigneeID,
		"workspace_id": workspaceID,
	}
}

// getOverdueStoriesForAssignee gets all stories needing attention for a specific assignee in one workspace.
func getOverdueStoriesForAssignee(ctx context.Context, db *sqlx.DB, assigneeID, workspaceID uuid.UUID) ([]OverdueStory, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getOverdueStoriesForAssignee")
	defer span.End()

	query := overdueStoriesForAssigneeQuery()
	params := overdueStoriesForAssigneeParams(assigneeID, workspaceID)

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare stories query: %w", err)
	}
	defer stmt.Close()

	var stories []OverdueStory
	if err := stmt.SelectContext(ctx, &stories, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute stories query: %w", err)
	}

	span.AddEvent("stories retrieved", trace.WithAttributes(
		attribute.String("assignee_id", assigneeID.String()),
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("stories.count", len(stories)),
	))

	return stories, nil
}

// sendOverdueStoriesEmailForAssignee sends email directly for a specific assignee
func sendOverdueStoriesEmailForAssignee(ctx context.Context, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, stories []OverdueStory) error {
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

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{firstStory.AssigneeEmail},
		Template: "notifications/notification",
		Subject:  title,
		Data:     data,
		Sender:   mailer.SenderProfileMaya,
		MessageID: guidanceEmailMessageID(
			"task-guidance",
			firstStory.WorkspaceID,
			firstStory.AssigneeID,
			time.Now(),
		),
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
