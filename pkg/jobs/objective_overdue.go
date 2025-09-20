package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ProcessObjectiveOverdue processes overdue objectives and sends emails directly
func ProcessObjectiveOverdue(ctx context.Context, db *sqlx.DB, log *logger.Logger, brevoService *brevo.Service) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessObjectiveOverdue")
	defer span.End()

	log.Info(ctx, "Processing objective overdue notifications")
	startTime := time.Now()

	const leadBatchSize = 100 // Process 100 leads at a time
	totalProcessed := 0
	totalEmailsCreated := 0
	batchCount := 0

	for {
		batchCount++
		log.Info(ctx, fmt.Sprintf("Processing objective lead batch %d", batchCount))

		// Get next batch of leads with overdue objectives (filtered by email preferences)
		leads, err := getLeadsWithOverdueObjectives(ctx, db, leadBatchSize, batchCount-1)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get leads batch %d: %w", batchCount, err)
		}

		if len(leads) == 0 {
			break // No more leads
		}

		// Process each lead in this batch
		for _, lead := range leads {
			// Get objectives for this specific lead
			objectives, err := getOverdueObjectivesForLead(ctx, db, lead.LeadUserID)
			if err != nil {
				log.Error(ctx, "Failed to get objectives for lead", "lead_id", lead.LeadUserID, "error", err)
				continue
			}

			if len(objectives) > 0 {
				// Send email directly for this lead
				err := sendObjectiveOverdueEmailForLead(ctx, log, brevoService, objectives)
				if err != nil {
					log.Error(ctx, "Failed to send email", "lead_id", lead.LeadUserID, "error", err)
					continue
				}
				totalEmailsCreated++
			}
			totalProcessed++
		}

		log.Info(ctx, fmt.Sprintf("Lead batch %d completed: %d leads processed", batchCount, len(leads)))

		// Small delay to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)

	span.AddEvent("objective overdue job completed", trace.WithAttributes(
		attribute.Int64("leads.processed", int64(totalProcessed)),
		attribute.Int64("emails.created", int64(totalEmailsCreated)),
		attribute.Int("batches.processed", batchCount),
		attribute.String("duration", duration.String()),
	))

	log.Info(ctx, fmt.Sprintf("Objective overdue job completed: %d leads processed, %d emails created in %d batches over %v",
		totalProcessed, totalEmailsCreated, batchCount, duration))

	return nil
}

// OverdueObjective represents an objective that needs attention
type OverdueObjective struct {
	ID             uuid.UUID `db:"objective_id"`
	Name           string    `db:"name"`
	EndDate        time.Time `db:"end_date"`
	LeadUserID     uuid.UUID `db:"lead_user_id"`
	LeadEmail      string    `db:"lead_email"`
	LeadName       string    `db:"lead_name"`
	WorkspaceID    uuid.UUID `db:"workspace_id"`
	WorkspaceName  string    `db:"workspace_name"`
	WorkspaceSlug  string    `db:"workspace_slug"`
	DeadlineStatus string    `db:"deadline_status"`
	DaysDifference int       `db:"days_difference"`
}

// getLeadsWithOverdueObjectives gets a batch of leads who have objectives needing attention and email enabled
func getLeadsWithOverdueObjectives(ctx context.Context, db *sqlx.DB, batchSize int, offset int) ([]OverdueObjective, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getLeadsWithOverdueObjectives")
	defer span.End()

	query := `
		SELECT DISTINCT 
			o.lead_user_id,
			u.email as lead_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) as lead_name,
			w.workspace_id,
			w.name as workspace_name,
			w.slug as workspace_slug,
			CAST(COALESCE(np.preferences -> 'reminders' ->> 'email', 'true') AS BOOLEAN) AS email_enabled
		FROM objectives o
		JOIN users u ON o.lead_user_id = u.user_id
		JOIN workspaces w ON o.workspace_id = w.workspace_id
		LEFT JOIN notification_preferences np ON o.lead_user_id = np.user_id AND o.workspace_id = np.workspace_id
		WHERE o.end_date IS NOT NULL
			AND o.deleted_at IS NULL
			AND o.lead_user_id IS NOT NULL
			AND o.end_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE + INTERVAL '7 days'
			AND u.is_active = true
			AND CAST(COALESCE(np.preferences -> 'reminders' ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY o.lead_user_id
		LIMIT :batch_size OFFSET :offset`

	params := map[string]any{
		"batch_size": batchSize,
		"offset":     offset * batchSize,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare leads query: %w", err)
	}
	defer stmt.Close()

	var leads []OverdueObjective
	if err := stmt.SelectContext(ctx, &leads, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute leads query: %w", err)
	}

	span.AddEvent("leads retrieved", trace.WithAttributes(
		attribute.Int("leads.count", len(leads)),
		attribute.Int("batch_size", batchSize),
		attribute.Int("offset", offset),
	))

	return leads, nil
}

// getOverdueObjectivesForLead gets all objectives needing attention for a specific lead
func getOverdueObjectivesForLead(ctx context.Context, db *sqlx.DB, leadID uuid.UUID) ([]OverdueObjective, error) {
	ctx, span := web.AddSpan(ctx, "jobs.getOverdueObjectivesForLead")
	defer span.End()

	query := `
		WITH objective_deadlines AS (
			SELECT 
				o.objective_id, o.name, o.end_date, o.lead_user_id, o.workspace_id,
				u.email as lead_email, 
				COALESCE(NULLIF(u.full_name, ''), u.username) as lead_name,
				w.name as workspace_name, w.slug as workspace_slug,
				CASE 
					WHEN o.end_date = CURRENT_DATE THEN 'due_today'
					WHEN o.end_date = CURRENT_DATE + INTERVAL '1 day' THEN 'due_tomorrow'
					WHEN o.end_date = CURRENT_DATE + INTERVAL '7 days' THEN 'due_in_7_days'
					WHEN o.end_date < CURRENT_DATE THEN 'overdue'
					ELSE 'future'
				END as deadline_status,
				CASE 
					WHEN o.end_date < CURRENT_DATE THEN CAST(CURRENT_DATE - o.end_date AS int)
					ELSE CAST(o.end_date - CURRENT_DATE AS int)
				END as days_difference
			FROM objectives o
			JOIN users u ON o.lead_user_id = u.user_id
			JOIN workspaces w ON o.workspace_id = w.workspace_id
			WHERE o.lead_user_id = :lead_id
				AND o.end_date IS NOT NULL
				AND o.deleted_at IS NULL
				AND o.end_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE + INTERVAL '7 days'
				AND u.is_active = true
		)
		SELECT * 
		FROM objective_deadlines 
		WHERE deadline_status IN ('due_today', 'due_tomorrow', 'due_in_7_days', 'overdue')
		ORDER BY deadline_status, end_date;
`

	params := map[string]any{
		"lead_id": leadID,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare objectives query: %w", err)
	}
	defer stmt.Close()

	var objectives []OverdueObjective
	if err := stmt.SelectContext(ctx, &objectives, params); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute objectives query: %w", err)
	}

	span.AddEvent("objectives retrieved", trace.WithAttributes(
		attribute.String("lead_id", leadID.String()),
		attribute.Int("objectives.count", len(objectives)),
	))

	return objectives, nil
}

// sendObjectiveOverdueEmailForLead sends an email to a lead about their overdue objectives
func sendObjectiveOverdueEmailForLead(ctx context.Context, log *logger.Logger, brevoService *brevo.Service, objectives []OverdueObjective) error {
	ctx, span := web.AddSpan(ctx, "jobs.sendObjectiveOverdueEmailForLead")
	defer span.End()

	if len(objectives) == 0 {
		return nil
	}

	// Group objectives by deadline status
	var dueSoonObjectives, dueTodayObjectives, overdueObjectives []OverdueObjective

	for _, objective := range objectives {
		switch objective.DeadlineStatus {
		case "due_in_7_days", "due_tomorrow":
			dueSoonObjectives = append(dueSoonObjectives, objective)
		case "due_today":
			dueTodayObjectives = append(dueTodayObjectives, objective)
		case "overdue":
			overdueObjectives = append(overdueObjectives, objective)
		}
	}

	// Use data from first objective for common fields
	firstObjective := objectives[0]
	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", firstObjective.WorkspaceSlug)

	// Format email content
	emailContent := formatObjectiveOverdueEmailContent(firstObjective, dueSoonObjectives, dueTodayObjectives, overdueObjectives, workspaceURL)

	// Send email via Brevo service
	totalCount := len(dueSoonObjectives) + len(dueTodayObjectives) + len(overdueObjectives)
	itemText := "objective"
	if totalCount > 1 {
		itemText = "objectives"
	}
	title := fmt.Sprintf("%d %s need attention", totalCount, itemText)

	params := brevo.EmailNotificationParams{
		Subject:             title,
		UserName:            firstObjective.LeadName,
		UserEmail:           firstObjective.LeadEmail,
		WorkspaceName:       firstObjective.WorkspaceName,
		WorkspaceURL:        workspaceURL,
		NotificationTitle:   title,
		NotificationMessage: emailContent,
		NotificationType:    "reminders",
	}

	if err := brevoService.SendEmailNotification(ctx, brevo.TemplateObjectiveOverdue, params); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to send objective overdue email: %w", err)
	}

	log.Info(ctx, "Successfully sent objective overdue email",
		"lead_id", firstObjective.LeadUserID,
		"lead_email", firstObjective.LeadEmail,
		"workspace_name", firstObjective.WorkspaceName,
		"total_objectives", totalCount)

	span.AddEvent("email prepared", trace.WithAttributes(
		attribute.String("lead_id", firstObjective.LeadUserID.String()),
		attribute.String("lead_email", firstObjective.LeadEmail),
		attribute.String("workspace_name", firstObjective.WorkspaceName),
		attribute.Int("objectives.count", len(objectives)),
	))

	return nil
}

// formatObjectiveOverdueEmailContent formats the email content
func formatObjectiveOverdueEmailContent(firstObjective OverdueObjective, dueSoonObjectives, dueTodayObjectives, overdueObjectives []OverdueObjective, workspaceURL string) string {
	totalItems := len(dueSoonObjectives) + len(dueTodayObjectives) + len(overdueObjectives)
	itemText := "objective"
	if totalItems > 1 {
		itemText = "objectives"
	}
	content := fmt.Sprintf(`
		<h3>Hi %s,</h3>
		<p>You have %d %s that need attention</p>
	`, firstObjective.LeadName, totalItems, itemText)

	if len(dueSoonObjectives) > 0 {
		itemText := "objective"
		if len(dueSoonObjectives) > 1 {
			itemText = "objectives"
		}
		content += fmt.Sprintf(`
			<p><strong>Due soon (%d %s)</strong></p>
			<ul>
		`, len(dueSoonObjectives), itemText)
		for _, objective := range dueSoonObjectives {
			content += fmt.Sprintf(`
				<li><a href="%s/objectives/%s" style="color: #000000; text-decoration: underline;">%s</a> - Due %s</li>
			`, workspaceURL, objective.ID.String(), objective.Name, objective.EndDate.Format("January 2, 2006"))
		}
		content += "</ul>"
	}

	if len(dueTodayObjectives) > 0 {
		itemText := "objective"
		if len(dueTodayObjectives) > 1 {
			itemText = "objectives"
		}
		content += fmt.Sprintf(`
			<p><strong>Due today (%d %s)</strong></p>
			<ul>
		`, len(dueTodayObjectives), itemText)
		for _, objective := range dueTodayObjectives {
			content += fmt.Sprintf(`
				<li><a href="%s/objectives/%s" style="color: #000000; text-decoration: underline;">%s</a> - Due today</li>
			`, workspaceURL, objective.ID.String(), objective.Name)
		}
		content += "</ul>"
	}

	if len(overdueObjectives) > 0 {
		itemText := "objective"
		if len(overdueObjectives) > 1 {
			itemText = "objectives"
		}
		content += fmt.Sprintf(`
			<p><strong>Overdue (%d %s)</strong></p>
			<ul>
		`, len(overdueObjectives), itemText)
		for _, objective := range overdueObjectives {
			daysText := "day"
			if objective.DaysDifference > 1 {
				daysText = "days"
			}
			content += fmt.Sprintf(`
				<li><a href="%s/objectives/%s" style="color: #000000; text-decoration: underline;">%s</a> - %d %s overdue</li>
			`, workspaceURL, objective.ID.String(), objective.Name, objective.DaysDifference, daysText)
		}
		content += "</ul>"
	}

	return content
}
