package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessOverdueStoriesEmail processes overdue stories and sends emails directly
func ProcessOverdueStoriesEmail(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
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

		// Get next batch of assignees with overdue stories
		assignees, err := getAssigneesWithOverdueStories(ctx, db, assigneeBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get assignees batch %d: %w", batchCount, err)
		}

		if len(assignees) == 0 {
			break // No more assignees
		}

		// Process each assignee in this batch
		for _, assigneeID := range assignees {
			stories, err := getOverdueStoriesForAssignee(ctx, db, assigneeID)
			if err != nil {
				log.Error(ctx, "Failed to get stories for assignee", "assignee_id", assigneeID, "error", err)
				continue
			}

			if len(stories) > 0 {
				// Send email directly for this assignee
				err := sendOverdueStoriesEmailForAssignee(ctx, log, stories)
				if err != nil {
					log.Error(ctx, "Failed to send email", "assignee_id", assigneeID, "error", err)
					continue
				}
				totalEmailsCreated++
			}
			totalProcessed++
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
	StatusName     string    `db:"status_name"`
	StatusCategory string    `db:"status_category"`
	DeadlineStatus string    `db:"deadline_status"`
	DaysDifference int       `db:"days_difference"`
}

// getAssigneesWithOverdueStories gets a batch of assignees who have stories needing attention
func getAssigneesWithOverdueStories(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getAssigneesWithOverdueStories")
	defer span.End()

	query := `
		SELECT DISTINCT s.assignee_id
		FROM stories s
		JOIN users u ON s.assignee_id = u.user_id
		JOIN statuses st ON s.status_id = st.status_id
		WHERE s.end_date IS NOT NULL
			AND st.category NOT IN ('completed', 'cancelled')
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.completed_at IS NULL
			AND s.assignee_id IS NOT NULL
			AND s.end_date BETWEEN CURRENT_DATE - INTERVAL '3 days' AND CURRENT_DATE + INTERVAL '3 days'
			AND u.is_active = true
		ORDER BY s.assignee_id
		LIMIT :batch_size OFFSET :offset`

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

	var assigneeIDs []uuid.UUID
	if err := stmt.SelectContext(ctx, &assigneeIDs, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute assignees query: %w", err)
	}

	span.AddEvent("assignees retrieved", trace.WithAttributes(
		attribute.Int("assignees.count", len(assigneeIDs)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))

	return assigneeIDs, nil
}

// getOverdueStoriesForAssignee gets all stories needing attention for a specific assignee
func getOverdueStoriesForAssignee(ctx context.Context, db *sqlx.DB, assigneeID uuid.UUID) ([]OverdueStory, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getOverdueStoriesForAssignee")
	defer span.End()

	query := `
		WITH story_deadlines AS (
    SELECT 
        s.id, s.title, s.end_date, s.assignee_id, s.workspace_id, s.team_id,
        u.email as assignee_email, 
        COALESCE(NULLIF(u.full_name, ''), u.username) as assignee_name,
        w.name as workspace_name, w.slug as workspace_slug,
        t.name as team_name, 
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
    JOIN teams t ON s.team_id = t.team_id
    JOIN statuses st ON s.status_id = st.status_id
    WHERE s.assignee_id = :assignee_id
        AND s.end_date IS NOT NULL
        AND st.category NOT IN ('completed', 'cancelled')
        AND s.deleted_at IS NULL
        AND s.archived_at IS NULL
        AND s.completed_at IS NULL
        AND s.end_date BETWEEN CURRENT_DATE - INTERVAL '3 days' AND CURRENT_DATE + INTERVAL '3 days'
        AND u.is_active = true
		)
		SELECT * 
		FROM story_deadlines 
		WHERE deadline_status IN ('due_today', 'due_tomorrow', 'due_in_3_days', 'overdue')
		ORDER BY deadline_status, end_date;
`

	params := map[string]any{
		"assignee_id": assigneeID,
	}

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
		attribute.Int("stories.count", len(stories)),
	))

	return stories, nil
}

// sendOverdueStoriesEmailForAssignee sends email directly for a specific assignee
func sendOverdueStoriesEmailForAssignee(ctx context.Context, log *logger.Logger, stories []OverdueStory) error {
	ctx, span := web.AddSpan(ctx, "jobs.sendOverdueStoriesEmailForAssignee")
	defer span.End()

	if len(stories) == 0 {
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

	// Format email content
	emailContent := formatOverdueStoriesEmailContent(firstStory, dueSoonStories, dueTodayStories, overdueStories, workspaceURL)

	// TODO: Send email via Brevo service
	// This would need to be implemented with the actual Brevo service
	log.Info(ctx, "Would send overdue stories email",
		"assignee_id", firstStory.AssigneeID,
		"assignee_email", firstStory.AssigneeEmail,
		"workspace_name", firstStory.WorkspaceName,
		"total_stories", len(stories),
		"email_content_length", len(emailContent),
		"email_content", emailContent)

	span.AddEvent("email prepared", trace.WithAttributes(
		attribute.String("assignee_id", firstStory.AssigneeID.String()),
		attribute.String("assignee_email", firstStory.AssigneeEmail),
		attribute.String("workspace_name", firstStory.WorkspaceName),
		attribute.Int("stories.count", len(stories)),
	))

	return nil
}

// formatOverdueStoriesEmailContent formats the email content
func formatOverdueStoriesEmailContent(firstStory OverdueStory, dueSoonStories, dueTodayStories, overdueStories []OverdueStory, workspaceURL string) string {
	content := fmt.Sprintf(`
		<h2>Hi %s,</h2>
		<p>You have %d stories that need attention in %s.</p>
	`, firstStory.AssigneeName, len(dueSoonStories)+len(dueTodayStories)+len(overdueStories), firstStory.WorkspaceName)

	if len(dueSoonStories) > 0 {
		content += fmt.Sprintf(`
			<h3>DUE SOON (%d stories)</h3>
			<ul>
		`, len(dueSoonStories))
		for _, story := range dueSoonStories {
			content += fmt.Sprintf(`
				<li><strong>%s</strong> - Due %s (%s) <a href="%s/story/%s">View</a></li>
			`, story.Title, story.EndDate.Format("January 2, 2006"), story.TeamName, workspaceURL, story.ID.String())
		}
		content += "</ul>"
	}

	if len(dueTodayStories) > 0 {
		content += fmt.Sprintf(`
			<h3>DUE TODAY (%d stories)</h3>
			<ul>
		`, len(dueTodayStories))
		for _, story := range dueTodayStories {
			content += fmt.Sprintf(`
				<li><strong>%s</strong> - Due today (%s) <a href="%s/story/%s">View</a></li>
			`, story.Title, story.TeamName, workspaceURL, story.ID.String())
		}
		content += "</ul>"
	}

	if len(overdueStories) > 0 {
		content += fmt.Sprintf(`
			<h3>OVERDUE (%d stories)</h3>
			<ul>
		`, len(overdueStories))
		for _, story := range overdueStories {
			daysText := "day"
			if story.DaysDifference > 1 {
				daysText = "days"
			}
			content += fmt.Sprintf(`
				<li><strong>%s</strong> - %d %s overdue (%s) <a href="%s/story/%s">View</a></li>
			`, story.Title, story.DaysDifference, daysText, story.TeamName, workspaceURL, story.ID.String())
		}
		content += "</ul>"
	}

	content += fmt.Sprintf(`
		<p><a href="%s/stories?filter=my-overdue">View all stories</a></p>
	`, workspaceURL)

	return content
}
